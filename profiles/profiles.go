package profiles

import (
	_ "embed"
	"encoding/json"
	"orsted/server/utils"
)

//go:embed default.json
var profile []byte

type ProfileConfig struct {
	Endpoints        map[string]string `json:"endpoints"`
	Domain           string            `json:"domain"`
	Port             string            `json:"port"`
	Interval         int64             `json:"interval"`
	Jitter           int64             `json:"jitter"`
	HeadersHttp      map[string]string `json:"headersHttp"`
	WinModulePath    string            `json:"windows-modules-path"`
}

var Config ProfileConfig

func InitialiseProfile() error {
    utils.PrintInfo("Unmarshelling Profile Config")
	err := json.Unmarshal(profile, &Config)
	if err != nil {
		return err
	}
	return nil
}
