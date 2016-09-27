package projection //import "github.com/kpmy/mipfs/dav_ipfs/projection"

import "strings"

func splitPath(filepath string) []string {
	return strings.Split(strings.Trim(filepath, "/"), "/")
}
