package version

import (
	"encoding/json"
)

var (
	version      string
	gitCommit    string
	gitTreeState string
)

type Version struct {
	Version      string
	GitCommit    string
	GitTreeState string
}

func GetVersion() Version {
	return Version{
		Version:      version,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
	}
}

func GetVersionString() string {
	ver := GetVersion()
	vStr, _ := json.Marshal(ver)
	return string(vStr)
}
