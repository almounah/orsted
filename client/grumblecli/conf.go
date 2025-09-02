package grumblecli

import "github.com/BurntSushi/toml"

type ClientConf struct {
	DefaultHTTPSCert  string
	DefaultHTTPSKey   string
	NetAssemblyPath   string
	Ps1ScriptPath     string
	WindowsModulePath string
	LinuxModulePath   string
	LinuxCommands     []string
	WindowsCommands   []string
	GlobalCommands    []string
}

var Conf ClientConf

func ParseClientConf(path string) error {
	_, err := toml.DecodeFile(path, &Conf)
	return err
}
