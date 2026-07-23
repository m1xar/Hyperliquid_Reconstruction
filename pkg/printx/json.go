package printx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tidwall/pretty"
)

func PrintObject(object any) {
	callerFile, callerLine := callerLocation(2)

	objectBytes, err := json.Marshal(object)
	if err != nil {
		fmt.Printf("[printx] %s:%d\njson.Marshal: %v", callerFile, callerLine, err)
		return
	}

	result := pretty.Pretty(objectBytes)

	fmt.Printf("[printx] %s:%d\n%s", callerFile, callerLine, string(result))
}

func callerLocation(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", 0
	}

	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil && isLocalPath(rel) {
			file = rel
		}
	}

	return filepath.Clean(file), line
}

func isLocalPath(path string) bool {
	return path != ".." &&
		!filepath.IsAbs(path) &&
		!strings.HasPrefix(path, ".."+string(filepath.Separator))
}
