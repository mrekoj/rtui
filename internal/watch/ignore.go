package watch

import "strings"

func shouldIgnorePath(path string) bool {
	ignoreContains := []string{
		"/node_modules/",
		"/.git/objects/",
		"/.git/logs/",
		"/dist/",
		"/build/",
		"/.cache/",
	}
	for _, frag := range ignoreContains {
		if strings.Contains(path, frag) {
			return true
		}
	}
	return false
}
