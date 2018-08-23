package raftnode

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

func DefaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "raft")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "raft")
		} else {
			return filepath.Join(home, ".raft")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
