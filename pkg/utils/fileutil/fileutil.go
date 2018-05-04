package fileutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	log "github.com/sirupsen/logrus"
	bindata "github.com/javefang/kaptain/data"
	"github.com/javefang/kaptain/pkg/api"
)

// TODO: this should probabaly be configurable
const defaultFileMode = 0644

func GetAsset(assetPath string) ([]byte, error) {
	return bindata.Asset(assetPath)
}

func RenderTemplate(templatePath string, args interface{}) ([]byte, error) {
	log.Debugf("Rendering template %s", templatePath)

	tmplData, err := GetAsset(templatePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read template %s: %v", templatePath, err)
	}

	tmpl, err := template.New(templatePath).Parse(string(tmplData))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse template %s: %v", templatePath, err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, args); err != nil {
		return nil, fmt.Errorf("Failed to render template %s: %v", templatePath, err)
	}

	return buffer.Bytes(), nil
}

func WriteAll(prefix string, files []*api.ClusterFile) error {
	for _, file := range files {
		fullPath := path.Join(prefix, file.Path)

		data, err := file.GetData()
		if err != nil {
			return fmt.Errorf("Failed to decode data for %s: %v", fullPath, err)
		}

		log.Infof("writing file: %s (len: %d)", fullPath, len(data))
		if err = Write(data, fullPath); err != nil {
			return fmt.Errorf("Error writing file %s: %v", fullPath, err)
		}
	}

	return nil
}

func EnsureDirExists(dir string) error {
	return os.MkdirAll(dir, 0700)
}

func Write(data []byte, outfile string) error {
	log.Debugf("Writing file %s (data length: %d)", outfile, len(data))
	return ioutil.WriteFile(outfile, data, defaultFileMode)
}
