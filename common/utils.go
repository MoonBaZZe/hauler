package common

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// DefaultDataDir is the default data directory to use for the databases and other persistence requirements.
func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		switch runtime.GOOS {
		case "darwin":
			panic("darwin not supported")
		case "windows":
			panic("windows not supported")
		default:
			return filepath.Join(home, ".hauler")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}
