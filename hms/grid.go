package hms

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var GridManagerKeyword string = "Grid Manager: "
var GridStartKeyword string = "Grid: "
var GridEndKeyword string = "End:"
var GridTypeKeyword string = "     Grid Type: "
var PrecipitationKeyword string = "Precipitation"
var TemperatureKeyword string = "Temperature"
var DssPathNameKeyword string = "       DSS Pathname: "
var DssFileNameKeyword string = "       DSS File Name: "
var GridStormCenterXKeyword string = "     Storm Center X: "
var GridStormCenterYKeyword string = "     Storm Center Y: "

type PrecipGridEvent struct {
	Name      string
	StartTime string  //parse DDMMMYYYY:HHMM //24 hour clocktime
	CenterX   float64 //maybe float64?
	CenterY   float64 //maybe float64?
	Lines     []string
}
type TempGridEvent struct {
	Name  string
	Lines []string
}
type GridFileInfo struct {
	Lines []string
}
type GridFile struct {
	GridFileInfo
	Events []PrecipGridEvent
	Temps  []TempGridEvent
}

func ReadGrid(gridResource []byte) (GridFile, error) {
	//read bytes
	//loop through and find grids
	gridstring := string(gridResource)
	lines := strings.Split(gridstring, "\n") //maybe rn?
	precipgrids := make([]PrecipGridEvent, 0)
	tempgrids := make([]TempGridEvent, 0)
	var precipGrid PrecipGridEvent
	var tempGrid TempGridEvent
	var gridFileInfo GridFileInfo
	var precipGridLines = make([]string, 0)
	var tempGridLines = make([]string, 0)
	var gridLines = make([]string, 0)
	var gridFound = false
	var isPrecipGrid = false
	var isTempGrid = false
	var foundX = false
	var foundY = false
	for _, l := range lines {
		l = strings.Replace(l, "\r", "", -1) //remove returns if they exist in the line.
		if l == "" {
			continue
		}
		if strings.Contains(l, GridStartKeyword) {
			gridFound = true
			precipGridLines = make([]string, 0)
			tempGridLines = make([]string, 0)
			name := strings.TrimLeft(l, GridStartKeyword)
			//wont know it is precip for one more line...
			//so get the name just in case.
			//add the first line just in case.
			isPrecipGrid = false
			isTempGrid = false
			precipGrid = PrecipGridEvent{Name: name}
			tempGrid = TempGridEvent{Name: name}
			precipGridLines = append(precipGridLines, l)
			tempGridLines = append(tempGridLines, l)
		}
		if strings.Contains(l, GridTypeKeyword) {
			gridType := strings.TrimLeft(l, GridTypeKeyword)
			if gridType == PrecipitationKeyword {
				isPrecipGrid = true
				foundX = false
				foundY = false
				//pop the last line off of the gridLines because it is a precip grid not a different grid type.
				gridLines = gridLines[:len(gridLines)-1]
			}
			if gridType == TemperatureKeyword {
				isTempGrid = true
				//pop the last line off of the gridLines because it is a temp grid not a different grid type.
				gridLines = gridLines[:len(gridLines)-1]
			}
		}
		if gridFound {
			if isPrecipGrid {
				if strings.Contains(l, GridStormCenterXKeyword) {
					centerxstring := strings.TrimLeft(l, GridStormCenterXKeyword)
					x, err := strconv.ParseFloat(centerxstring, 64)
					if err != nil {
						foundX = false
					} else {
						foundX = true
						precipGrid.CenterX = x
					}
				}
				if strings.Contains(l, GridStormCenterYKeyword) {
					centerystring := strings.TrimLeft(l, GridStormCenterYKeyword)
					y, err := strconv.ParseFloat(centerystring, 64)
					if err != nil {
						foundY = false
					} else {
						foundY = true
						precipGrid.CenterY = y
					}
				}
				precipGridLines = append(precipGridLines, l)
			} else if isTempGrid {
				tempGridLines = append(tempGridLines, l)
			} else {
				gridLines = append(gridLines, l) //adding everythign that is a grid.
			}

		} else {
			gridLines = append(gridLines, l) //adding everything that isnt a precip grid too!
		}
		//check after adding to include grid end in the data and to set the precip grids into the grid list.
		if strings.Contains(l, GridEndKeyword) {
			if gridFound {
				gridFound = false
				if isPrecipGrid {
					if foundX && foundY {
						precipGrid.Lines = precipGridLines
						precipgrids = append(precipgrids, precipGrid)
					} else {
						/*plugin.Log(plugin.Message{
							Status:    plugin.COMPUTING,
							Progress:  10,
							Level:     plugin.INFO,
							Message:   fmt.Sprintf("found grid %v but found no x and y center not adding grid to grid list\r\n", precipGrid.Name),
							Sender:    "hms-mutator",
							PayloadId: "unknown payload id",
						})*/
					}

				} else if isTempGrid {
					tempGrid.Lines = tempGridLines
					tempgrids = append(tempgrids, tempGrid)
				}
			}
		}

		if strings.Contains(l, DssPathNameKeyword) {
			pathName := strings.TrimLeft(l, DssPathNameKeyword)
			parts := strings.Split(pathName, "/")
			startTime := parts[4] //parse DDMMMYYYY:HHMM //24 hour clocktime
			precipGrid.StartTime = startTime
			if strings.Contains(precipGrid.StartTime, "2400") {
				precipGrid.StartTime = strings.Replace(precipGrid.StartTime, "2400", "2359", 1)
			}
		}

	}
	gridFileInfo.Lines = gridLines
	if len(precipgrids) == 0 {
		return GridFile{GridFileInfo: gridFileInfo, Events: precipgrids}, errors.New("found no grids with x and y centers specified, please specify storm centers for transposition")
	}
	return GridFile{GridFileInfo: gridFileInfo, Events: precipgrids, Temps: tempgrids}, nil
}
func (gf *GridFile) Bootstrap(knowledgeUncertaintySeed int64) error {
	length := len(gf.Events)
	r := rand.New(rand.NewSource(knowledgeUncertaintySeed))
	updatedList := make([]PrecipGridEvent, length)
	for i := 0; i < length; i++ {
		idx := r.Int31n(int32(length))
		updatedList[i] = gf.Events[idx] //sample with replacement.
	}
	gf.Events = updatedList //replace dataset with bootstrap.
	return nil
}
func (gf GridFile) SelectEvent(naturalVariabilitySeed int64) (PrecipGridEvent, TempGridEvent, error) {
	//randomly select one event from the list of events
	length := len(gf.Events)
	r := rand.New(rand.NewSource(naturalVariabilitySeed))
	idx := r.Int31n(int32(length))
	pge := gf.Events[idx]
	for _, tempEvent := range gf.Temps {
		if strings.Contains(pge.Name, tempEvent.Name) {
			return pge, tempEvent, nil
		}
	}
	return pge, TempGridEvent{}, nil
}
func (gf GridFile) SelectEventByIndex(idx int64) (PrecipGridEvent, error) {
	//provide the indexed event
	return gf.Events[idx-1], nil
}
func (pge *PrecipGridEvent) OriginalDSSFile() (string, error) {
	for _, l := range pge.Lines {
		if strings.Contains(l, DssFileNameKeyword) {
			output := strings.Split(l, DssFileNameKeyword)[1]
			return output, nil
		}
	}
	return "", errors.New("did not find the dss file name keyword")
}
func (pge *PrecipGridEvent) UpdateDSSFile(stormName string) error {
	//force the name to be constant in the file. "/data/Storm.dss"
	path := fmt.Sprintf("data/%v.dss", stormName)
	for idx, l := range pge.Lines {
		if strings.Contains(l, DssFileNameKeyword) {
			pge.Lines[idx] = fmt.Sprintf("%v%v", DssFileNameKeyword, path)
		}
	}
	return nil
}
func (pge *TempGridEvent) UpdateDSSFile(stormName string) error {
	//force the name to be constant in the file. "/data/Storm.dss"
	path := fmt.Sprintf("data/%v.dss", stormName)
	for idx, l := range pge.Lines {
		if strings.Contains(l, DssFileNameKeyword) {
			pge.Lines[idx] = fmt.Sprintf("%v%v", DssFileNameKeyword, path)
		}
	}
	return nil
}
func (gf GridFile) ToBytes(precipEvent PrecipGridEvent, tempEvent TempGridEvent) []byte {
	b := make([]byte, 0)
	for _, l := range gf.GridFileInfo.Lines {
		b = append(b, l...)
		if l == GridEndKeyword {
			b = append(b, "\r\n"...)
		}
		b = append(b, "\r\n"...)
	}

	for _, l := range precipEvent.Lines {
		b = append(b, l...)
		b = append(b, "\r\n"...)
	}
	if tempEvent.Name == "" {
		gridName := precipEvent.Name
		gridNameSansAORC := strings.Trim(gridName, "AORC ")
		gridNameSansAORC = strings.ReplaceAll(gridNameSansAORC, " ", "_")
		gridNameSansAORC = strings.ReplaceAll(gridNameSansAORC, "-", "")
		tempMonkeyPatch := fmt.Sprintf("Grid: %v\n     Grid Type: Temperature\n     Last Modified Date: 9 August 2022\n     Last Modified Time: 17:14:25\n     Reference Height Units: Meters\n     Reference Height: 10.0\n     Data Source Type: External DSS\n     Variant: Variant-1\n       Last Variant Modified Date: 9 August 2022\n       Last Variant Modified Time: 17:14:25\n       Default Variant: Yes\n       DSS File Name: data/Storm.dss\n       DSS Pathname: /SHG2K/KANAWHA/TEMPERATURE///AORC/\n     End Variant: Variant-1\n     Use Lookup Table: No\nEnd:\n", gridName)
		b = append(b, tempMonkeyPatch...)
	} else {
		for _, l := range tempEvent.Lines {
			b = append(b, l...)
			b = append(b, "\r\n"...)
		}
	}

	return b
}
