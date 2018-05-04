package store

import (
	"fmt"
	"net/url"
	"strings"
)

func CreateStoreFromUrlOrDie(storeUrl string) Store {
	parsedURL, err := url.Parse(strings.ToLower(storeUrl))
	if err != nil {
		panic(fmt.Errorf("failed to parse store url '%s': %v", storeUrl, err))
	}

	queries := parsedURL.Query()

	switch parsedURL.Scheme {
	case "s3":
		region := getFirstOrEmpty(queries, "region")
		assumeRole := getFirstOrEmpty(queries, "assume_role")
		return createS3Store(parsedURL.Host, region, assumeRole)
	case "vault":
		roleID := getFirstOrEmpty(queries, "role_id")
		secretID := getFirstOrEmpty(queries, "secret_id")
		vaultPath := parsedURL.Host + parsedURL.Path
		return createVaultStoreOrDie(vaultPath, roleID, secretID)
	default:
		panic(fmt.Errorf("failed to create store '%s': unknown scheme '%s'", storeUrl, parsedURL.Scheme))
	}
}

func getFirstOrEmpty(queries url.Values, key string) string {
	if len(queries[key]) > 0 {
		return queries[key][0]
	}
	return ""
}
