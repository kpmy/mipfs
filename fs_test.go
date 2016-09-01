package mipfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFS(t *testing.T) {
	NewFS()
	NewLS()
}

func TestTrav(t *testing.T) {
	root := string([]rune{os.PathSeparator})
	var trav func(string) interface{}
	trav = func(name string) interface{} {
		if name == root {
			return ""
		} else {
			_, last := filepath.Split(name)
			trav(filepath.Dir(name))
			t.Log(last)
			return nil
		}
	}

	trav("/user/data/pic.gif")
}
