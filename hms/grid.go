package hms

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var GridManagerKeyword string = "Grid Manager: "
var GridStartKeyword string = "Grid: "
var GridEndKeyword string = "End:"
var GridTypeKeyword string = "     Grid Type: "
var PrecipitationKeyword string = "Precipitation"
var DssPathNameKeyword string = "       DSS Pathname: "
var DssFileNameKeyword string = "       DSS File Name: "
var GridStormCenterXKeyword string = "     Storm Center X: "
var GridStormCenterYKeyword string = "     Storm Center Y: "

type PrecipGridEvent struct {
	Name      string
	StartTime string //parse DDMMMYYYY:HHMM //24 hour clocktime
	Lines     []string
}
type GridFileInfo struct {
	Lines []string
}
type GridFile struct {
	GridFileInfo
	Events []PrecipGridEvent
}

func ReadGrid(gridResource plugin.ResourceInfo) (GridFile, error) {
	//read bytes
	//loop through and find grids
	bytes, err := plugin.DownloadObject(gridResource)
	if err != nil {
		return GridFile{}, err
	}
	gridstring := string(bytes)
	lines := strings.Split(gridstring, "\n") //maybe rn?
	grids := make([]PrecipGridEvent, 0)
	var precipGrid PrecipGridEvent
	var gridFileInfo GridFileInfo
	var precipGridLines = make([]string, 0)
	var gridLines = make([]string, 0)
	var gridFound = false
	var isPrecipGrid = false
	for _, l := range lines {
		l = strings.Replace(l, "\r", "", -1) //remove returns if they exist in the line.
		if l == "" {
			continue
		}
		if strings.Contains(l, GridStartKeyword) {
			gridFound = true
			precipGridLines = make([]string, 0)

			name := strings.TrimLeft(l, GridStartKeyword)
			//wont know it is precip for one more line...
			//so get the name just in case.
			//add the first line just in case.
			isPrecipGrid = false
			precipGrid = PrecipGridEvent{Name: name}
			precipGridLines = append(precipGridLines, l)
		}
		if strings.Contains(l, GridTypeKeyword) {
			gridType := strings.TrimLeft(l, GridTypeKeyword)
			if gridType == PrecipitationKeyword {
				isPrecipGrid = true
				//pop the last line off of the gridLines because it is a precip grid not a different grid type.
				gridLines = gridLines[:len(gridLines)-1]
			}
		}
		if gridFound {
			if isPrecipGrid {
				precipGridLines = append(precipGridLines, l)
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
					precipGrid.Lines = precipGridLines
					grids = append(grids, precipGrid)
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
	return GridFile{GridFileInfo: gridFileInfo, Events: grids}, nil
}

func (gf GridFile) SelectEvent(seed int64) (PrecipGridEvent, error) {
	//randomly select one event from the list of events
	length := len(gf.Events)
	r := rand.New(rand.NewSource(seed))
	idx := r.Int31n(int32(length))
	return gf.Events[idx], nil
}
func (pge PrecipGridEvent) DownloadAndUploadDSSFile(dssResourceInfo plugin.ResourceInfo, outputResourceInfo plugin.ResourceInfo) error {
	bytes, _ := plugin.DownloadObject(dssResourceInfo)
	//output destination should be "/data/Storm.dss"
	plugin.UpLoadFile(outputResourceInfo, bytes)
	return nil
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
func (pge *PrecipGridEvent) UpdateDSSFile(path string) error {
	for idx, l := range pge.Lines {
		if strings.Contains(l, DssFileNameKeyword) {
			pge.Lines[idx] = fmt.Sprintf("%v%v", DssFileNameKeyword, path)
		}
	}
	return nil
}
func (gf GridFile) toBytes(precipEvent PrecipGridEvent) []byte {
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
	return b
}
func (gf GridFile) Write(outputResourceInfo plugin.ResourceInfo, precipEvent PrecipGridEvent) error {
	b := gf.toBytes(precipEvent)
	return plugin.UpLoadFile(outputResourceInfo, b)
}
