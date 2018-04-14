package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

// ProgramName holds the name of the program for building config paths
const ProgramName = "shadowhosts"

// ConfigName holds the expected config file name for building paths
const ConfigName = "config.toml"

// Sep is a shorthand for the system filepath separator
const Sep = string(filepath.Separator)

const DefaultConfig = `# [DANGEROUS] Uncomment to allow redirection entries from remote sources
#allow_redirect = true

# Add additional sources of domains to block here
sources = [
	"https://adaway.org/hosts.txt",
	"https://hosts-file.net/ad_servers.txt",
	"https://pgl.yoyo.org/adservers/serverlist.php?hostformat=hosts&showintro=0&mimetype=plaintext"
]

# Add additional domains to block here
blacklist = []

# Add domains to unblock from online sources here
whitelist = []

# Add redirection rules here, following the example
[redirect]
	# "localhost" = "127.0.0.1"
`

// Test that file exists
func fileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func inUserConfigDir() string {
	// Then check for configuration in the user's config directory
	env := ""
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || strings.Contains(runtime.GOOS, "bsd") || runtime.GOOS == "dragonfly" {
		env = os.Getenv("HOME") + Sep + ".config"
	} else if runtime.GOOS == "windows" {
		env = os.Getenv("APPDATA")
	}

	// If found in user configuration directory, return it
	if env != "" && env != Sep+".config" {
		file := env + Sep + ProgramName + Sep + ConfigName
		return file
	}

	return ""
}

func findConfigFile() (string, error) {
	// If configuration file specified, use it
	if flags.Config != "" {
		if !fileExists(flags.Config) {
			return "", fmt.Errorf("Configuration file %s does not exist", flags.Config)
		}
		return flags.Config, nil
	}

	// Check for a portable config first
	file := "." + Sep + ConfigName
	if fileExists(file) {
		return file, nil
	}

	// Check in user config directory
	file = inUserConfigDir()
	if file != "" && fileExists(file) {
		return file, nil
	}

	// Else if a *nix system, check /etc
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || strings.Contains(runtime.GOOS, "bsd") || runtime.GOOS == "dragonfly" {
		if fileExists("/etc") {
			return "/etc" + Sep + ProgramName + Sep + ConfigName, nil
		}
	}

	// If not found, error
	return "", fmt.Errorf("could not find an existing configuration file. Use --makeconfig to generate one")
}

// GetHostsConfig returns a HostsConfig or an error if something failed
func GetHostsConfig() (HostsConfig, error) {
	// Get configuration file path
	config := NewHostsConfig()
	configFile, err := findConfigFile()
	if err != nil {
		return config, err
	}

	// Parse configuration file
	_, err = toml.DecodeFile(configFile, &config)
	if err != nil {
		return config, err
	}

	// Return configuration
	return config, nil
}

// GenerateConfig creates a default configuration file at the given location
func GenerateConfig(out string) error {
	parent := filepath.Dir(out)
	err := os.MkdirAll(parent, 0755)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(out, ([]byte)(DefaultConfig), 0644)
	return err
}
