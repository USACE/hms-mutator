package hms

import (
	"math/rand"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var GridStartKeyword string = "Grid: "
var GridEndKeyword string = "End:"
var GridTypeKeyword string = "     Grid Type: "
var PrecipitationKeyword string = "Precipitation"
var DssPathNameKeyword string = "       DSS Pathname: "

type PrecipGridEvent struct {
	Name      string
	StartTime string //parse DDMMMYYYY:HHMM //24 hour clocktime
	Lines     []string
}

type GridFile struct {
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
	lines := strings.Split(gridstring, "\r\n") //maybe rn?
	grids := make([]PrecipGridEvent, 0)
	var precipGrid PrecipGridEvent
	var gridLines = make([]string, 0)
	var gridFound = false
	var isPrecipGrid = false
	for _, l := range lines {
		if gridFound {
			gridLines = append(gridLines, l)
		}
		if strings.Contains(l, GridStartKeyword) {
			gridFound = true
			gridLines = make([]string, 0)
			gridLines = append(gridLines, l)
			isPrecipGrid = false
			name := strings.TrimLeft(l, GridStartKeyword)
			precipGrid = PrecipGridEvent{Name: name}
		}
		if strings.Contains(l, GridTypeKeyword) {
			gridType := strings.TrimLeft(l, GridTypeKeyword)
			if gridType == PrecipitationKeyword {
				isPrecipGrid = true
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
		if strings.Contains(l, GridEndKeyword) {
			gridFound = false
			if isPrecipGrid {
				precipGrid.Lines = gridLines
				grids = append(grids, precipGrid)
			}
		}
	}
	return GridFile{Events: grids}, nil
}

func (gf GridFile) SelectEvent(seed int64) (PrecipGridEvent, error) {
	//randomly select one event from the list of events
	length := len(gf.Events)
	r := rand.New(rand.NewSource(seed))
	idx := r.Int31n(int32(length))
	return gf.Events[idx], nil
}
