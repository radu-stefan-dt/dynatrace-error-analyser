package util

import (
	"os"
	"strings"
)

func ReplacePathSeparators(path string) (newPath string) {
	newPath = strings.ReplaceAll(path, "\\", string(os.PathSeparator))
	newPath = strings.ReplaceAll(newPath, "/", string(os.PathSeparator))

	return newPath
}
