package mipfs

import (
	"github.com/peterbourgon/diskv"
)

var memo *diskv.Diskv

func init() {
	memo = diskv.New(diskv.Options{
		BasePath: ".memo",
		Transform: func(s string) []string {
			return []string{}
		},
	})
}
