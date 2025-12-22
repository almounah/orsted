package grumblecli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/desertbit/grumble"
)

func SetGenerateBeaconCommand() {
	generateCmd := &grumble.Command{
		Name: "generate",
		Help: "generate beacon or implant from a given beacon",
	}

	generateBeaconCmd := &grumble.Command{
		Name: "beacon",
		Help: "generate beacon (HTTP, HTTPS, TCP or SMB _ windows only _",
		Args: func(a *grumble.Args) {
			a.String("os", "windows or linux")
			a.String("type", "http[|dll|svc] https[|dll|svc] tcp[|dll|svc] or smb[|dll|svc]")
			a.String("address", "address for http or tcp")
		},
		Flags: func(f *grumble.Flags) {
			f.Bool("n", "no-console", false, "If Specified, beacon will not have a console on windows")
			f.String("t", "http-proxy-type", "none", "HTTP Proxy Type, can be HTTP or HTTPS")
			f.String("a", "http-proxy-address", "", "URL for example 127.0.0.1:8080")
			f.String("u", "http-proxy-username", "", "Proxy Username")
			f.String("p", "http-proxy-password", "", "Proxy Password")
		},
		Completer: func(prefix string, args []string) []string {
			var suggestions []string

			// 1. Flag completion
			flagSuggestions := []string{
				"--http-proxy-type",
				"--http-proxy-address",
				"--http-proxy-username",
				"--http-proxy-password",
				"--no-console",
			}

			// complete flags
			if strings.HasPrefix(prefix, "-") {
				for _, f := range flagSuggestions {
					if strings.HasPrefix(f, prefix) {
						suggestions = append(suggestions, f)
					}
				}
				return suggestions
			}

			// 2. Positional arguments (your existing logic)
			transportListLinux := []string{"http", "https", "tcp"}
			transportListWindows := []string{
				"http", "http_dll", "http_svc",
				"https", "https_dll", "https_svc",
				"tcp", "tcp_dll", "tcp_svc",
				"smb", "smb_dll", "smb_svc",
			}
			osList := []string{"windows", "linux"}

			var modulesList []string
			if len(args) == 0 {
				modulesList = osList
			}
			if len(args) == 1 {
				if args[0] == "windows" {
					modulesList = transportListWindows
				} else {
					modulesList = transportListLinux
				}
			}

			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName)
				}
			}

			return suggestions
		},
		Run: func(c *grumble.Context) error {
			beaconType := c.Args.String("type")
			console := !c.Flags.Bool("no-console")
			switch beaconType {
			case "http", "http_dll", "http_svc", "https", "https_dll", "https_svc", "tcp", "tcp_dll", "tcp_svc",
				"smb", "smb_dll", "smb_svc":
				// supported type — proceed
			default:
				fmt.Println("unsupported beacon type")
				return nil
			}
			beaconOs := c.Args.String("os")
			if beaconOs != "windows" && beaconOs != "linux" {
				fmt.Println("Os should be windows, linux")
				return nil
			}

			beaconAddress := c.Args.String("address")

			address := strings.Split(beaconAddress, ":")
			if len(address) != 2 {
				fmt.Println("Address should be in the form IP:PORT")
				return nil
			}
			beaconIP := address[0]
			beaconPort := address[1]


			// Get Proxy Stuff
			proxyType := c.Flags.String("http-proxy-type")
			if proxyType != "http" && proxyType != "https" && proxyType != "none" {
				fmt.Println("Proxy should be http or https")
				return nil
			}

			proxyAddress := c.Flags.String("http-proxy-address")
			if proxyAddress != "" {
				proxyAddressSplit := strings.Split(proxyAddress, ":")
				if len(proxyAddressSplit) != 2 {
					fmt.Println("Proxy Address should be in the form IP:PORT")
					return nil
				}
			}

			proxyUsername := c.Flags.String("http-proxy-username")
			proxyPassword := c.Flags.String("http-proxy-password")


			ldflags := fmt.Sprintf("-s -w -X main.Targetip=%s -X main.Targetport=%s -X main.HTTPProxyType=%s -X main.HTTPProxyURL=%s -X main.HTTPProxyUsername=%s -X main.HTTPProxyPassword=%s", beaconIP, beaconPort, proxyType, proxyAddress, proxyUsername, proxyPassword)

			if !console && beaconOs == "windows" {
				ldflags += " -H=windowsgui"

			}
			mainPath := fmt.Sprintf("beacon/main_%s.go", beaconType)
			tags := fmt.Sprintf("-tags=%s", beaconType)

			cmd := exec.Command(
				"go", "build",
				"-ldflags", ldflags,
				"-trimpath",
				tags,
				mainPath,
			)

			if strings.HasSuffix(beaconType, "_dll") {
				cmd = exec.Command(
					"go", "build",
					"-ldflags", ldflags,
					"-trimpath",
					"-buildmode=c-shared",
					"-o", fmt.Sprintf("main_%s.dll", beaconType),
					tags,
					mainPath,
				)
			}

			// Set environment variables like GOOS and GOARCH
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOOS=%s", beaconOs),
				"GOARCH=amd64",
			)

			if strings.HasSuffix(beaconType, "_dll") {
				cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
				cmd.Env = append(cmd.Env, "CC=x86_64-w64-mingw32-gcc")
			}

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Run the command
			fmt.Println(mainPath)
			err := cmd.Run()
			if err != nil {
				fmt.Println("Build failed:", err)
			}

			fmt.Println("Beacon Generated")

			return nil
		},
	}
	generateCmd.AddCommand(generateBeaconCmd)
	app.AddCommand(generateCmd)
}
