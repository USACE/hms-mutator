package hms

import (
	"errors"
	"fmt"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var SeedKeyword string = "     Seed Value: "
var RealizationsKeyword string = "     Number Of Realizations: "
var DefaultRealizationsValue int = 1

type Mca struct {
	SeedStringIndex  int
	HasRealizations  bool
	RealizationIndex int
	Lines            []string
}

func ReadMca(mcaResource []byte) (Mca, error) {
	//read bytes
	//loop through and find met and precip blocks
	bytes, err := plugin.DownloadObject(mcaResource)
	if err != nil {
		return Mca{}, err
	}
	mcafilestring := string(bytes)
	lines := strings.Split(mcafilestring, "\r\n") //maybe rn?
	seedFound := false
	hasRealizations := false
	seedStringIndex := 0
	realizationIndex := 0
	for idx, l := range lines {
		if strings.Contains(l, SeedKeyword) {
			seedFound = true
			seedStringIndex = idx
		}
		if strings.Contains(l, RealizationsKeyword) {
			hasRealizations = true
			realizationIndex = idx
		}
	}
	if !hasRealizations { //if the file doesnt specifiy realizations force it to be 1.
		lines = append(lines[:seedStringIndex+1], lines[seedStringIndex:]...)
		lines[seedStringIndex] = fmt.Sprintf("%v%v", RealizationsKeyword, 1)
		realizationIndex = seedStringIndex
		seedStringIndex = seedStringIndex + 1
		hasRealizations = true
	} else {
		//force to 1
		//@TODO allow user to specify they want more than one?
		lines[realizationIndex] = fmt.Sprintf("%v%v", RealizationsKeyword, 1)
	}
	mcaModel := Mca{
		SeedStringIndex:  seedStringIndex,
		HasRealizations:  hasRealizations,
		RealizationIndex: realizationIndex,
		Lines:            lines,
	}
	if seedFound {
		return mcaModel, nil
	} else {
		return mcaModel, errors.New("no seed found in the *.mca file")
	}
}
func (mf *Mca) UpdateSeed(seed int64) error {
	mf.Lines[mf.SeedStringIndex] = fmt.Sprintf("%v%v", SeedKeyword, seed)
	return nil
}

/*
	func (mf *Mca) UpdateRealizations(count int) error {
		mf.Lines[mf.RealizationIndex] = fmt.Sprintf("%v%v", RealizationsKeyword, count)
		return nil
	}
*/
func (mf Mca) UploadToS3(outputResourceInfo plugin.ResourceInfo) error {
	b := make([]byte, 0)
	for _, l := range mf.Lines {
		b = append(b, l...)
		b = append(b, "\r\n"...) //? are we sure i need to add those back in?
	}
	return plugin.UpLoadFile(outputResourceInfo, b)
}
