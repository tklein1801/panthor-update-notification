package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/robfig/cron"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Interval      string `yaml:"interval"`
		LoadOnStartup bool   `yaml:"load_on_startup"`
	} `yaml:"app"`
	Notification struct {
		Webhooks []string `yaml:"webhooks"`
	} `yaml:"notification"`
}

type Version struct {
	Version string `yaml:"version"`
}

type ChangelogResponse struct {
	Data        []Changelog `json:"data"`
	RequestedAt int         `json:"requested_at"`
}

type Changelog struct {
	ID            int      `json:"id"`
	Version       string   `json:"version"`
	ChangeMission []string `json:"change_mission"`
	ChangeMap     []string `json:"change_map"`
	ChangeMod     []string `json:"change_mod"`
	Note          string   `json:"note"`
	Active        int      `json:"active"`
	Size          string   `json:"size"`
	ReallifeRpg   int      `json:"realliferpg"`
	ReleaseAt     string   `json:"release_at"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

func main() {
	config, err := LoadConfig("config.yml")
	if err != nil {
		log.Fatalln("Failed to load config:", err)
		return
	}

	if config.App.LoadOnStartup || !DoesFileExist("version.yml") {
		changelogs, err := GetChangelogs()
		if err != nil {
			log.Fatalln("Failed to get changelogs:", err)
			return
		}

		if len(*changelogs) == 0 {
			log.Fatalln("no changelogs found")
			return
		}

		err = SaveVersion((*changelogs)[0].Version)
		if err != nil {
			log.Fatalln("Failed to save version:", err)
			return
		}

		log.Println("Version of the first item:", (*changelogs)[0].Version)
	}

	c := cron.New()
	c.AddFunc(config.App.Interval, func() {
		changelogs, err := GetChangelogs()
		if err != nil {
			log.Fatalln("Failed to get changelogs:", err)
			return
		}

		if len(*changelogs) == 0 {
			log.Fatalln("no changelogs found")
			return
		}

		changelog := (*changelogs)[0]
		savedVersion, err := GetSavedVerison()
		if err != nil {
			log.Fatalln("Failed to get saved version:", err)
			return
		}

		if changelog.Version == savedVersion.Version {
			log.Println("Version is the same! No new version avaiable.")
			return
		}

		log.Println("New version is", changelog.Version)

		err = SaveVersion(changelog.Version)
		if err != nil {
			log.Fatalln("Failed to save version:", err)
			return
		}

		// Notifications
		for _, webhook := range config.Notification.Webhooks {
			requestBody, err := json.Marshal(map[string]interface{}{
				"content":      fmt.Sprintf("New version %s is available!", changelog.Version),
				"version":      changelog.Version,
				"size":         changelog.Size,
				"hasModUpdate": strconv.FormatBool(len(changelog.ChangeMod) > 0),
				"releaseAt":    changelog.ReleaseAt,
			})
			if err != nil {
				log.Println("Failed to marshal request body:", err)
				continue
			}

			err = TriggerWebhook(webhook, requestBody)
			if err != nil {
				log.Println("Failed to trigger webhook:", err)
				continue
			}
		}
	})

	c.Start()

	log.Println("Panthor Update Notification started...")

	select {}
}

// TriggerWebhook sends a POST request to the specified webhook URL with the given request body.
// It returns an error if the request fails or if the response status code is not 200 OK.
func TriggerWebhook(webhook string, requestBody []byte) error {
	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error making POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// LoadConfig loads the configuration from the specified file.
// It reads the file, parses the YAML data, and returns a pointer to the Config struct.
// If there is an error reading the file or parsing the YAML data, it returns an error.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading the file: %w", err)
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing the YAML file: %w", err)
	}

	return &config, nil
}

// DoesFileExist checks if a file exists in the given path.
// It returns true if the file exists, and false otherwise.
func DoesFileExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetChangelogs retrieves the changelogs from the Panthor API.
// It sends a GET request to the specified URL and parses the response into a list of Changelog structs.
// If successful, it returns a pointer to the list of Changelogs and nil error.
// If an error occurs during the HTTP request or response parsing, it returns nil and an error.
func GetChangelogs() (*[]Changelog, error) {
	url := "https://api.panthor.de/v1/changelog"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var parsedResponse ChangelogResponse
	err = json.NewDecoder(resp.Body).Decode(&parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	return &parsedResponse.Data, nil
}

// GetSavedVerison reads the version information from the "version.yml" file and returns it as a Version struct.
// If there is an error reading the file or parsing the YAML, an error is returned.
func GetSavedVerison() (*Version, error) {
	data, err := os.ReadFile("version.yml")
	if err != nil {
		return nil, fmt.Errorf("error reading the file: %w", err)
	}

	var version Version

	err = yaml.Unmarshal(data, &version)
	if err != nil {
		return nil, fmt.Errorf("error parsing the YAML file: %w", err)
	}

	return &version, nil
}

// SaveVersion saves the given version string to a YAML file named "version.yml".
// It marshals the version data to YAML format and writes it to the file.
// If any error occurs during the process, it returns an error.
// The file permissions for the created file are set to 0644.
//
// Parameters:
//   - version: The version string to be saved.
//
// Returns:
//   - error: An error if any occurred during the process, otherwise nil.
func SaveVersion(version string) error {
	data := Version{Version: version}
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data to YAML: %w", err)
	}
	err = os.WriteFile("version.yml", yamlData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
