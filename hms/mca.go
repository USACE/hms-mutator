package hms

import (
	"errors"
	"fmt"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var SeedKeyword string = "     Seed Value: "

type Mca struct {
	SeedStringIndex int
	Lines           []string
}

func ReadMca(mcaResource plugin.ResourceInfo) (Mca, error) {
	//read bytes
	//loop through and find met and precip blocks
	bytes, err := plugin.DownloadObject(mcaResource)
	if err != nil {
		return Mca{}, err
	}
	mcafilestring := string(bytes)
	lines := strings.Split(mcafilestring, "\r\n") //maybe rn?
	seedFound := false
	seedStringIndex := 0
	for idx, l := range lines {
		if strings.Contains(l, SeedKeyword) {
			seedFound = true
			seedStringIndex = idx
		}
	}
	mcaModel := Mca{
		SeedStringIndex: seedStringIndex,
		Lines:           lines,
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
func (mf Mca) UploadToS3(outputResourceInfo plugin.ResourceInfo) error {
	b := make([]byte, 0)
	for _, l := range mf.Lines {
		b = append(b, l...)
		b = append(b, "\r\n"...) //? are we sure i need to add those back in?
	}
	return plugin.UpLoadFile(outputResourceInfo, b)
}
