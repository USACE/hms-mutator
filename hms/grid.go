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
	lines := strings.Split(gridstring, "\n") //maybe rn?
	grids := make([]PrecipGridEvent, 0)
	var precipGrid PrecipGridEvent
	var isPrecipGrid = false
	for _, l := range lines {
		if strings.Contains(l, GridStartKeyword) {
			isPrecipGrid = false
			name := strings.TrimLeft(l, GridStartKeyword)
			precipGrid = PrecipGridEvent{Name: name}
		}
		if strings.Contains(l, GridTypeKeyword) {
			gridType := strings.TrimLeft(l, GridTypeKeyword)
			if gridType == PrecipEndKeyword {
				isPrecipGrid = true
			}
		}
		if strings.Contains(l, DssPathNameKeyword) {
			pathName := strings.TrimLeft(l, DssPathNameKeyword)
			parts := strings.Split(pathName, "/")
			startTime := parts[3] //parse DDMMMYYYY:HHMM //24 hour clocktime
			precipGrid.StartTime = startTime
		}
		if strings.Contains(l, GridEndKeyword) {
			if isPrecipGrid {
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
