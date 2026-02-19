package watch

import (
	"path/filepath"
	"strings"
)

func shouldIgnorePath(path string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(filepath.Separator))

	for _, part := range parts {
		switch part {
		case "node_modules", "dist", "build", "target", ".cache":
			return true
		}
	}

	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == ".git" && (parts[i+1] == "objects" || parts[i+1] == "logs") {
			return true
		}
	}
	return false
}
