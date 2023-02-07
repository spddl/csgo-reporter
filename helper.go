package main

import (
	"os"
	"strings"
)

func pathJoin(elem ...string) string {
	return strings.Join(elem, "\\")
}
func ExistsFolder(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
