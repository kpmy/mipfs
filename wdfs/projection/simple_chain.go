package projection

import "strings"

func splitPath(filepath string) []string {
	return strings.Split(strings.Trim(filepath, "/"), "/")
}
