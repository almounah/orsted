package grumblecli

import "github.com/BurntSushi/toml"

type ClientConf struct {
	DefaultHTTPSCert      string
	DefaultHTTPSKey       string
	NetAssemblyPath       string
	DonutPath             string
	DefaultAmsiBypass     string
	DefaultLoaderTemplate string
	BeaconShellCodePath   string
	ExePath               string
	BatcaveJsonPath       string
	BatcaveURL            string
	Ps1ScriptPath         string
	WindowsModulePath     string
	LinuxModulePath       string
	LinuxCommands         []string
	WindowsCommands       []string
	GlobalCommands        []string
}

var Conf ClientConf

func ParseClientConf(path string) error {
	_, err := toml.DecodeFile(path, &Conf)
	return err
}
