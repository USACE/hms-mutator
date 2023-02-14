package transposition

import (
	"fmt"
	"math/rand"

	"github.com/usace/hms-mutator/hms"
	"github.com/usace/wat-go-sdk/plugin"
)

type Simulation struct {
	transpositionModel Model
	metModel           hms.Met
	gridFile           hms.GridFile
	control            hms.Control
}

func InitSimulation(trgpkgRI []byte, wbgpkgRI []byte, metRI []byte, gridRI []byte, controlRI []byte) (Simulation, error) {
	s := Simulation{}
	//read grid file
	gf, err := hms.ReadGrid(gridRI)
	if err != nil {
		return s, err
	}
	s.gridFile = gf
	//initialize transposition region
	t, err := InitModel(trgpkgRI, wbgpkgRI) //TODO fix this.
	if err != nil {
		return s, err
	}
	s.transpositionModel = t
	//read control
	c, err := hms.ReadControl(controlRI)
	if err != nil {
		return s, err
	}
	s.control = c
	//read met file
	m, err := hms.ReadMet(metRI)
	if err != nil {
		return s, err
	}
	s.metModel = m
	return s, nil

}
func (s *Simulation) Compute(seeds plugin.SeedSet) (hms.Met, hms.PrecipGridEvent, error) {
	nvrng := rand.New(rand.NewSource(seeds.EventSeed))
	stormSeed := nvrng.Int63()
	transpositionSeed := nvrng.Int63()
	kurng := rand.New(rand.NewSource(seeds.RealizationSeed))
	bootstrapSeed := kurng.Int63()
	//bootstrap events
	s.gridFile.Bootstrap(bootstrapSeed)
	//select event
	ge, err := s.gridFile.SelectEvent(stormSeed)
	if err != nil {
		return s.metModel, ge, err
	}
	//transpose
	x, y, err := s.transpositionModel.Transpose(transpositionSeed, ge)
	if err != nil {
		return s.metModel, ge, err
	}
	//compute offset from control specification
	offset := s.control.ComputeOffset(ge.StartTime)
	//update met storm name
	err = s.metModel.UpdateStormName(ge.Name)
	if err != nil {
		return s.metModel, ge, err
	}
	//update storm center
	err = s.metModel.UpdateStormCenter(fmt.Sprintf("%v", x), fmt.Sprintf("%v", y))
	if err != nil {
		return s.metModel, ge, err
	}
	//update timeshift
	err = s.metModel.UpdateTimeShift(fmt.Sprintf("%v", offset))
	if err != nil {
		return s.metModel, ge, err
	}
	return s.metModel, ge, nil
}
func (s Simulation) UploadGridFile(gori plugin.ResourceInfo, pge hms.PrecipGridEvent) error {
	return s.gridFile.Write(gori, pge)
}
