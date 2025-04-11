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
func Test_UpdateGridFileStormNames(t *testing.T) {
	gridFilePath := "/workspaces/hms-mutator/exampledata/trinity/catalog.grid"
	data, err := os.ReadFile(gridFilePath)
	if err != nil {
		t.Fail()
	}
	stringdata := string(data)
	lines := strings.Split(stringdata, "\n")
	//search for Grid:
	//search for "     Grid: "
	//search for "     Filename: " /these should all contain the same string. replace the parts to match the new convention.
	firstGrid := ""
	secondGrid := ""
	//filename := ""
	firstGridLine := 0
	secondGridLine := 0
	filenameGridLine := 0
	isPrecip := false
	for i, line := range lines {

		if strings.Contains(line, "Grid: ") {
			isPrecip = false
			if strings.Contains(line, "     Grid: ") {
				secondGrid = strings.Split(line, ": ")[1]
				secondGridLine = i
			} else {
				firstGrid = strings.Split(line, ": ")[1]
				firstGridLine = i
			}
		}
		if strings.Contains(line, "     Filename: ") {
			//filename = strings.Split(line, ": ")[1]
			filenameGridLine = i
			if isPrecip {
				if firstGrid == secondGrid {
					//parse the gridname
					nameParts := strings.Split(firstGrid, " ")
					date := nameParts[1]
					rank := nameParts[2]
					rank = strings.Replace(rank, "T", "r", -1)
					stormType := nameParts[3]
					dateParts := strings.Split(date, "-")
					year := dateParts[0]
					month := dateParts[1]
					day := dateParts[2]
					newName := fmt.Sprintf("%v%v%v_72hr_%v_%v", year, month, day, stormType, rank)
					lines[firstGridLine] = fmt.Sprintf("Grid: %v", newName)
					lines[secondGridLine] = fmt.Sprintf("     Grid: %v", newName)
					lines[filenameGridLine] = fmt.Sprintf("     Filename: data\\%v.dss", newName)
				}
			}

		}
		if strings.Contains(line, "     Grid Type: Precipitation") {
			isPrecip = true
		}
	}
	newString := ""
	for _, line := range lines {
		newString = fmt.Sprintf("%v%v\n", newString, line)
	}
	newdata := make([]byte, 0)
	newdata = append(newdata, newString...)
	os.WriteFile(gridFilePath, newdata, 0600)
}
