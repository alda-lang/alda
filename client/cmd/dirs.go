package cmd

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/OpenPeeDeeP/xdg"
)

var cacheDir string

func init() {
	dirs := xdg.New("", "alda")
	cacheDir = dirs.CacheHome()

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
		cacheDir = filepath.Join(cacheDir, "cache")
	}
}

func cachePath(pathSegments ...string) string {
	allPathSegments := append([]string{cacheDir}, pathSegments...)
	return filepath.Join(allPathSegments...)
}

func queryCache(pathSegments ...string) string {
	filepath := cachePath(pathSegments...)

	_, err := os.Stat(filepath)

	if (err != nil && os.IsExist(err)) || err == nil {
		return filepath
	}

	return ""
}
