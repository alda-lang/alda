package system

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/OpenPeeDeeP/xdg"
)

// CacheDir is the system-dependent directory where we store temporary files to
// do things like write logs and keep track of the states of Alda player
// processes.
var CacheDir string

// ConfigDir is the system-dependent directory where we store user-specific
// config files.
var ConfigDir string

func init() {
	dirs := xdg.New("", "alda")
	CacheDir = dirs.CacheHome()

	// There is a discrepancy in the Windows cache path between the "standard
	// directories" libraries we're using in the Go client vs. the JVM player.
	//
	//   According to the Go library (OpenPeeDeeP/xdg):
	//     C:\Users\johndoe\AppData\Local\alda
	//
	//   According to the JVM library (soc/directories-jvm):
	//     C:\Users\johndoe\AppData\Local\alda\cache
	//
	// To compensate, we add a final "\cache" path segment if the OS is Windows.
	if runtime.GOOS == "windows" {
		CacheDir = filepath.Join(CacheDir, "cache")
	}

	ConfigDir = dirs.ConfigHome()
}

func pathImpl(baseDir string, pathSegments []string) string {
	allPathSegments := append([]string{baseDir}, pathSegments...)
	return filepath.Join(allPathSegments...)
}

func queryImpl(baseDir string, pathSegments []string) string {
	filepath := pathImpl(baseDir, pathSegments)

	_, err := os.Stat(filepath)

	if (err != nil && os.IsExist(err)) || err == nil {
		return filepath
	}

	return ""
}

// CachePath returns the full path to a cache file consisting of the provided
// segments. This is done in a cross-platform way according to XDG conventions.
func CachePath(pathSegments ...string) string {
	return pathImpl(CacheDir, pathSegments)
}

// QueryCache returns the full path to a cache file consisting of the provided
// segments, if-and-only-if that file currently exists.
//
// Returns "" if the file doesn't exist.
func QueryCache(pathSegments ...string) string {
	return queryImpl(CacheDir, pathSegments)
}

// ConfigPath returns the full path to a cache file consisting of the provided
// segments. This is done in a cross-platform way according to XDG conventions.
func ConfigPath(pathSegments ...string) string {
	return pathImpl(ConfigDir, pathSegments)
}

// QueryConfig returns the full path to a cache file consisting of the provided
// segments, if-and-only-if that file currently exists.
//
// Returns "" if the file doesn't exist.
func QueryConfig(pathSegments ...string) string {
	return queryImpl(ConfigDir, pathSegments)
}
