package grumblecli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/desertbit/grumble"
	"golang.org/x/term"
	"google.golang.org/grpc"
)

type Gadget struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Type string `json:"type"`
	Hash string `json:"hash"`
}

type BatCave struct {
	Gadgets []Gadget              `json:"gadgets"`
	Bundles []map[string][]string `json:"bundles"`
}

var BatcaveVar BatCave
var AuthToken string

func GetLatestURL(url string) string {
	// Parse the GitHub URL to extract owner and repo
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 3 {
		return ""
	}

	owner := matches[1]
	repo := strings.TrimSuffix(matches[2], ".git")

	// Build the GitHub API URL for latest release
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	// Make request to GitHub API
	resp, err := http.Get(apiURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	// Parse the JSON response
	var release struct {
		Assets []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
			Name               string `json:"name"`
		} `json:"assets"`
		ZipballURL string `json:"zipball_url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return ""
	}

	// Return the first asset download URL if available
	if len(release.Assets) > 0 {
		return release.Assets[0].BrowserDownloadURL
	}

	// Otherwise return the zipball URL as fallback
	return release.ZipballURL
}

func DownloadFileToPath(url, directory, fileName string) error {
	// Create the HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add the Authorization header
	if AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+AuthToken)
	}


	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	path := filepath.Join(directory, fileName)

	fmt.Println(fmt.Sprintf("Downloading %s to %s", url, path))

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the destination file
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func UnzipFile(zipPath, destDir string) error {
	// Open the zip file
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// Create destination directory if it doesn't exist
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract each file from the archive
	for _, f := range r.File {
		err := extractFile(f, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractFile extracts a single file from the zip archive
func extractFile(f *zip.File, destDir string) error {
	// Create the full path for the file
	filePath := filepath.Join(destDir, f.Name)

	// Check for ZipSlip vulnerability
	if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// Create directories if the file is a directory
	if f.FileInfo().IsDir() {
		return os.MkdirAll(filePath, f.Mode())
	}

	// Create parent directories for the file
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}

	// Create the destination file
	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()

	// Open the file from the zip archive
	zipFile, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file from zip: %w", err)
	}
	defer zipFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, zipFile)
	if err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}

// INitialize batcave by installing it and parsing the struct
func InitBatcave() error {
	batcaveURL := Conf.BatcaveURL
	batcaveJsonPath := Conf.BatcaveJsonPath

	batcaveName := "batcave.json"
	releaseURL := GetLatestURL(batcaveURL)
	err := DownloadFileToPath(releaseURL, batcaveJsonPath, batcaveName)

	batcaveLocation := filepath.Join(batcaveJsonPath, batcaveName)
	data, err := os.ReadFile(batcaveLocation)
	err = json.Unmarshal(data, &BatcaveVar)
	return err

}

func CheckBatcaveExists() bool {
	batcaveJsonPath := Conf.BatcaveJsonPath

	batcaveName := "batcave.json"
	batcaveLocation := filepath.Join(batcaveJsonPath, batcaveName)
	_, err := os.Stat(batcaveLocation)
	return err == nil
}

func SetBatcaveCommands(conn grpc.ClientConnInterface) {
	batcaveCmd := &grumble.Command{
		Name: "batcave",
		Help: "Commands related to the batcave",
	}

	authCmd := &grumble.Command{
		Name: "authenticate",
		Help: "authenticate to the API. Token is stored in client process memory.",
		Flags: func(f *grumble.Flags) {
			f.String("t", "token", "", "Token to authenticate to github API. If not specified will ask securely")
		},
		Run: func(c *grumble.Context) error {
			enteredTok := c.Flags.String("token")
			if enteredTok != "" {
				AuthToken = enteredTok
				fmt.Println("Token saved in client process memory")
				return nil 
			}
			fmt.Print("Enter Github API Token: ")
			password, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				fmt.Println("Failed to read password from terminal")
			}
			fmt.Println() // Print newline after password input

			// password is []byte, convert to string if needed
			AuthToken = string(password)
			fmt.Println("Token saved in client process memory")
			return nil
		},
	}
	initCmd := &grumble.Command{
		Name: "update",
		Help: "Repoll the batcave with any potential new batgadget or config",
		Run: func(c *grumble.Context) error {
			err := InitBatcave()
			if err != nil {
				fmt.Println("Error Updating Batcave ", err)
			}
			return nil
		},
	}

	searchCmd := &grumble.Command{
		Name: "search",
		Help: "Search for batgadget and batbundle",
		Args: func(f *grumble.Args) {
			f.String("param", "String to Search for")
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			searchParam := c.Args.String("param")
			if !CheckBatcaveExists() {
				err := InitBatcave()
				if err != nil {
					fmt.Println("Error Updating Batcave ", err)
					return nil
				}
			}

			var GadgetsNameList []string
			for _, g := range BatcaveVar.Gadgets {
				if strings.Contains(strings.ToLower(g.Name), strings.ToLower(searchParam)) {
					GadgetsNameList = append(GadgetsNameList, g.Name)
				}

			}

			var BundleNameList []string
			for _, bundle := range BatcaveVar.Bundles {
				for key := range bundle {
					if strings.Contains(strings.ToLower(key), strings.ToLower(searchParam)) {
						BundleNameList = append(BundleNameList, key)
					}
				}

			}
			fmt.Println("BatGadget ===> ")
			for _, g := range GadgetsNameList {
				fmt.Println(g)

			}
			fmt.Println(" ================================= ")
			for _, g := range BundleNameList {
				fmt.Println(g)

			}
			return nil
		},
	}

	installCmd := &grumble.Command{
		Name: "install",
		Help: "list current listener",
		Args: func(f *grumble.Args) {
			f.String("batType", "batgadget or batbundle to install")
			f.String("name", "Name of the batgadget or batbundle")
		},
		Completer: func(prefix string, args []string) []string {
			transportList := []string{"batgadget", "batbundle"}
			var suggestions []string

			var modulesList []string
			if len(args) == 0 {
				modulesList = transportList
			}

			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName)
				}
			}
			return suggestions
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			return nil
		},
	}

	batcaveCmd.AddCommand(searchCmd)
	batcaveCmd.AddCommand(authCmd)
	batcaveCmd.AddCommand(initCmd)
	batcaveCmd.AddCommand(installCmd)
	app.AddCommand(batcaveCmd)
}
