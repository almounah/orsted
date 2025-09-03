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
		Completer: func(prefix string, args []string) []string {
			transportListLinux := []string{"http", "https", "tcp"}
			transportListWindows := []string{"http", "http_dll", "http_svc", 
				                             "https", "https_dll", "https_svc", 
				                             "tcp", "tcp_dll", "tcp_svc",
				                             "smb", "smb_dll", "smb_svc"}
			osList := []string{"windows", "linux"}
			var suggestions []string

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
			switch beaconType {
			case "http", "http_dll", "http_svc", "https", "https_dll", "https_svc", "tcp", "tcp_dll", "tcp_svc",
				 "smb", "smv_dll", "smb_svc":
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

			ldflags := fmt.Sprintf("-s -w -X main.Targetip=%s -X main.Targetport=%s", beaconIP, beaconPort)
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
