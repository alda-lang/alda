package system

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"github.com/OpenPeeDeeP/xdg"
)

// AldaExecutablePath returns the full path to the `alda` executable that is
// running, i.e. this process.
func AldaExecutablePath() (string, error) {
	aldaPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Symlinks are ubiquitous on Unix/Linux systems, so we should make sure that
	// we're handling them correctly so that we know the _actual_ directory where
	// the `alda` executable lives, not the directory where the symlink is, if
	// it's symlinked.
	//
	// I don't trust that filepath.EvalSymlinks does something sane on Windows. I
	// did some googling and found some GitHub issues that were not encouraging.
	// So, we're just not going to do this on Windows. (I'm not even sure if
	// symlinks are a thing on Windows. If they are, I doubt they're common.)
	if runtime.GOOS != "windows" {
		return filepath.EvalSymlinks(aldaPath)
	}

	return aldaPath, nil
}

// RenamedExecutable returns a renamed version of the provided filepath,
// including a Unix timestamp and a random number to ensure that a unique string
// is returned each time this function is called.
func RenamedExecutable(path string) string {
	return fmt.Sprintf(
		"%s.%d.%d.old",
		path,
		time.Now().Unix(),
		rand.Intn(10000),
	)
}

// IsRenamedExecutable returns true if the provided file basename (e.g.
// "alda.1615061099.2081.old") looks like it's a renamed Alda executable.
func IsRenamedExecutable(basename string) bool {
	return regexp.MustCompile(
		`alda(-player)?(\.exe)?\.\d+\.\d+\.old`,
	).MatchString(basename)
}

// FindRenamedExecutables looks for renamed old versions of the `alda`
// executable. Returns a list of file paths, which is empty if no renamed
// executables are found.
func FindRenamedExecutables() ([]string, error) {
	aldaPath, err := AldaExecutablePath()
	if err != nil {
		return nil, err
	}

	aldaDir := filepath.Dir(aldaPath)

	fileInfos, err := os.ReadDir(aldaDir)
	if err != nil {
		return nil, err
	}

	renamedExecutables := []string{}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		basename := fileInfo.Name()

		if IsRenamedExecutable(basename) {
			renamedExecutable := filepath.Join(aldaDir, basename)
			renamedExecutables = append(renamedExecutables, renamedExecutable)
		}
	}

	return renamedExecutables, nil
}

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

	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	return filepath
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
