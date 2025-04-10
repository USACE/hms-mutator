package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Test_Main(t *testing.T) {
	main()
}
func Test_RenameStorms(t *testing.T) {
	dir := "/workspaces/hms-mutator/exampledata/trinity/storms"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fail()
	}
	for _, e := range entries {
		name := e.Name()
		nameParts := strings.Split(name, "_")
		storm_type := strings.Split(nameParts[2], ".")[0]
		newName := fmt.Sprintf("%v/%v_72hr_%v_%v.dss", dir, nameParts[0], storm_type, nameParts[1])
		oldName := fmt.Sprintf("%v/%v", dir, name)
		os.Rename(oldName, newName)
	}
}
