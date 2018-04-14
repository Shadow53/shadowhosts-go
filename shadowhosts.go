package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var flags struct {
	Config    string
	Output    string
	MkDir     bool
	GenConfig bool
}
var flagsAreInit = false

func initFlags() {
	if !flag.Parsed() {
		flag.StringVar(&flags.Config, "config", "", "Path to a configuration file for shadowhosts to use")
		flag.StringVar(&flags.Output, "out", "", "Path to file to output hosts file or default configuration to. File will be truncated if it exists")
		flag.BoolVar(&flags.MkDir, "mkdir", false, "Set to true to make any missing parent directories for the file specified in --out")
		flag.BoolVar(&flags.GenConfig, "genconfig", false, "Generate a default configuration and exit")
		flag.Parse()
	}
}

func getHostsFile() string {
	// Windows is special
	if runtime.GOOS == "windows" {
		return os.Getenv("SystemRoot") + Sep + "System32" + Sep + "drivers" + Sep + "etc" + Sep + "hosts"
	}

	// Everyone else makes it available here
	return "/etc/hosts"
}

func main() {
	initFlags()

	if flags.GenConfig {
		userConfig := inUserConfigDir()
		if flags.Output != "" {
			GenerateConfig(flags.Output)
		} else if userConfig != "" {
			GenerateConfig(userConfig)
		} else if !fileExists("." + Sep + ConfigName) {
			GenerateConfig("." + Sep + ConfigName)
		} else {
			fmt.Println("Could not determine where to put the configuration file. Try again using the --out flag.")
			os.Exit(1)
		}
		os.Exit(0)
	}

	config, err := GetHostsConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Figure out which file to output to
	var out string
	if flags.Output != "" {
		out = flags.Output
	} else {
		out = getHostsFile()
	}

	// Deal with parent directories
	parent := filepath.Dir(out)
	if !fileExists(parent) {
		if flags.MkDir {
			os.MkdirAll(parent, 0755)
		} else {
			fmt.Printf("Parent folders of %s don't exist. Create them yourself or pass the --mkdir flag to shadowhosts.\n", out)
			os.Exit(1)
		}
	}

	err = config.DownloadSources()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(out, config.GenerateHosts(), 0644)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Could not write to file %s - you may need to run this program with admin privileges\n", out)
			os.Exit(1)
		} else {
			fmt.Printf("Could not write to file %s: %s\n", out, err)
			os.Exit(1)
		}
	}
}
