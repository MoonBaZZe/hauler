package common

import (
	"github.com/syndtr/goleveldb/leveldb"
	lerrors "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// NextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two.  This is a helper function used during the
// calculation of a merkle tree.
func NextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
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

func CreateOrOpenLevelDb(name string) (*leveldb.DB, error) {
	opts := &opt.Options{OpenFilesCacheCapacity: 200}
	evDir := filepath.Join(DefaultDataDir(), DefaultHeaderChainDir)
	if _, err := os.Stat(evDir); os.IsNotExist(err) {
		if err = os.MkdirAll(evDir, 0700); err != nil {
			return nil, err
		}
	}
	dbDir := filepath.Join(evDir, name)
	ldb, err := leveldb.OpenFile(dbDir, opts)
	if _, isCorrupted := err.(*lerrors.ErrCorrupted); isCorrupted {
		ldb, err = leveldb.RecoverFile(dbDir, nil)
		if err != nil {
			return nil, err
		}
	}

	return ldb, nil
}

func DeleteLvlDb(name string) error {
	dbPath := filepath.Join(DefaultDataDir(), DefaultHeaderChainDir, name)
	return os.RemoveAll(dbPath)
}
