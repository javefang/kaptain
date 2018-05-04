package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

const defaultServerSideEncryption = "AES256"

type S3Store struct {
	Bucket   string
	S3Client *s3.S3
}

func createS3Store(bucket string, region string, assumeRoleARN string) Store {
	s3Client := getS3Client(region, assumeRoleARN)

	return &S3Store{
		Bucket:   bucket,
		S3Client: s3Client,
	}
}

func getS3Client(region string, assumeRoleARN string) *s3.S3 {
	sess := session.Must(session.NewSession())
	conf := &aws.Config{}

	if region != "" {
		conf.Region = aws.String(region)
	}

	if assumeRoleARN != "" {
		creds := stscreds.NewCredentials(sess, assumeRoleARN)
		conf.Credentials = creds
	}

	return s3.New(sess, conf)
}

func (store *S3Store) makeError(action string, key string, err error) error {
	return fmt.Errorf("failed to %s key '%s' from %s: %v", action, key, store, err)
}

func (store *S3Store) List(key string) ([]string, error) {
	store.log(fmt.Sprintf("List key %s", key))

	req := &s3.ListObjectsV2Input{
		Bucket:    aws.String(store.Bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(key),
	}

	resp, err := store.S3Client.ListObjectsV2(req)
	if err != nil {
		return nil, store.makeError("list", key, err)
	}

	names := make([]string, len(resp.CommonPrefixes))
	for i, elem := range resp.CommonPrefixes {
		names[i] = path.Base(aws.StringValue(elem.Prefix))
	}
	return names, nil
}

func (store *S3Store) Exists(key string) (bool, error) {
	store.log(fmt.Sprintf("Head key %s", key))

	params := &s3.HeadObjectInput{
		Bucket: aws.String(store.Bucket),
		Key:    aws.String(key),
	}

	req, _ := store.S3Client.HeadObjectRequest(params)
	err := req.Send()
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				return false, store.makeError("head", key, err)
			}
		} else {
			return false, store.makeError("head", key, err)
		}
	}

	return true, nil
}

func (store *S3Store) Get(key string) ([]byte, error) {
	store.log(fmt.Sprintf("Get key %s", key))

	req := &s3.GetObjectInput{
		Bucket: aws.String(store.Bucket),
		Key:    aws.String(key),
	}
	resp, err := store.S3Client.GetObject(req)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, store.makeError("get", key, err)
			default:
				return nil, store.makeError("get", key, err)
			}
		}
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, store.makeError("read", key, err)
	}

	return data, nil
}

func (store *S3Store) Set(key string, data []byte) error {
	store.log(fmt.Sprintf("Set key %s (len: %d bytes)", key, len(data)))

	req := &s3.PutObjectInput{
		Body:                 bytes.NewReader(data),
		Bucket:               aws.String(store.Bucket),
		Key:                  aws.String(key),
		ServerSideEncryption: aws.String(defaultServerSideEncryption),
	}

	_, err := store.S3Client.PutObject(req)
	if err != nil {
		return store.makeError("write", key, err)
	}

	return nil
}

func (store *S3Store) Delete(key string) error {
	store.log(fmt.Sprintf("Delete key %s", key))

	req := &s3.DeleteObjectInput{
		Bucket: aws.String(store.Bucket),
		Key:    aws.String(key),
	}

	_, err := store.S3Client.DeleteObject(req)
	if err != nil {
		return store.makeError("delete", key, err)
	}

	return nil
}

func (store *S3Store) DeleteAll(key string) error {
	store.log(fmt.Sprintf("DeleteAll key %s", key))

	// list all keys to delete
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(store.Bucket),
		Prefix: aws.String(key),
	}

	listOutput, err := store.S3Client.ListObjectsV2(listInput)
	if err != nil {
		return store.makeError("delete", key, err)
	}

	delCount := len(listOutput.Contents)
	if delCount == 0 {
		store.log("DeleteAll: nothing to delete")
		return nil
	}

	// delete all listed keys
	store.log(fmt.Sprintf("DeleteAll: deleting %d keys", delCount))
	objIdentifiers := make([]*s3.ObjectIdentifier, 0)
	for _, e := range listOutput.Contents {
		store.log(fmt.Sprintf("DeleteAll: delete %s", aws.StringValue(e.Key)))
		objIdentifiers = append(objIdentifiers, &s3.ObjectIdentifier{Key: e.Key})
	}

	// delete all objects
	delInput := &s3.DeleteObjectsInput{
		Bucket: aws.String(store.Bucket),
		Delete: &s3.Delete{
			Objects: objIdentifiers,
		},
	}

	delOutput, err := store.S3Client.DeleteObjects(delInput)
	if err != nil {
		return store.makeError("delete", key, err)
	}
	store.log(fmt.Sprintf("DeleteAll: %d keys deleted", len(delOutput.Deleted)))

	return nil
}

func (store *S3Store) String() string {
	return fmt.Sprintf("s3://%s", store.Bucket)
}

func (store *S3Store) log(msg string) {
	log.WithField("s3Bucket", store.Bucket).Debugf("S3_STORE: %s", msg)
}
