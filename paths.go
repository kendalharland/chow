package chow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Converts the input path to an absolute path for the current platform.
//
// The path is expected to have a Unix-style syntax, using '/' as the path separator.  The
// caller may refer to the "root" of the current task using '//', and the current working
// directory as './'.
func (r *prodRunner) convertAnyPaths(paths []string) error {
	for i, p := range paths {
		if strings.HasPrefix(p, "//CWD/") {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get cwd: %v", err)
			}

			suffix := strings.SplitN(p, "//CWD/", 2)[1]
			paths[i] = filepath.FromSlash(wd + "/" + suffix)
			continue
		}

		if strings.HasPrefix(p, "//") {
			suffix := strings.SplitN(p, "//", 2)[1]
			r.startDir = strings.TrimRight(r.startDir, "/")
			paths[i] = filepath.FromSlash(r.startDir + "/" + suffix)
			continue
		}

		if strings.HasPrefix(p, "/") {
			logWarning("unsafe use of absolute path " + p)
			paths[i] = filepath.FromSlash(p)
			continue
		}

		// Ignore relative paths and non-path arguments
	}

	return nil
}
