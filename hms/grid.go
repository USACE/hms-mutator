package hms

import (
	"math/rand"

	"github.com/usace/wat-go-sdk/plugin"
)

var string StartKeyword = "Grid:"
var string EndKeyword = "End:"

type GridEvent struct {
	Name string
	data string
}

type GridFile struct {
	Events []GridEvent
}

func ReadGrid(gridResource plugin.ResourceInfo) (GridFile, error) {
	//ensure it is local
	//read bytes
	//loop through and find grids
	return GridFile{}, nil
}

func (gf GridFile) SelectEvent(seed int64) (GridEvent, error) {
	//randomly select one event from the list of events
	length := len(gf.Events)
	r := rand.New(rand.NewSource(seed))
	idx := r.Int31n(int32(length))
	return gf.Events[idx], nil
}
