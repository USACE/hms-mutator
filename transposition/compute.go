package transposition

import (
	"fmt"
	"math/rand"

	"github.com/usace/hms-mutator/hms"
)

type TranspositionSimulation struct {
	transpositionModel Model
	metModel           hms.Met
	gridFile           hms.GridFile
	control            hms.Control
}
type WalkSimulation struct {
	metModel hms.Met
	mcaFile  hms.Mca
	gridFile hms.GridFile
	control  hms.Control
	csvFile  []byte
}

func InitWalkSimulation(metRI []byte, gridRI []byte, controlRI []byte, mcaRI []byte, csvRI []byte) (WalkSimulation, error) {
	s := WalkSimulation{}
	//read grid file
	gf, err := hms.ReadGrid(gridRI)
	if err != nil {
		return s, err
	}
	s.gridFile = gf
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
	//read mca file
	mca, err := hms.ReadMca(mcaRI)
	if err != nil {
		return s, err
	}
	s.mcaFile = mca
	s.csvFile = csvRI
	return s, nil
}
func InitTranspositionSimulation(trgpkgRI []byte, wbgpkgRI []byte, metRI []byte, gridRI []byte, controlRI []byte) (TranspositionSimulation, error) {
	s := TranspositionSimulation{}
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
func (s *TranspositionSimulation) Compute(eventSeed int64, realizationSeed int64) (hms.Met, hms.PrecipGridEvent, error) {
	nvrng := rand.New(rand.NewSource(eventSeed))
	stormSeed := nvrng.Int63()
	transpositionSeed := nvrng.Int63()
	kurng := rand.New(rand.NewSource(realizationSeed))
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
	fmt.Printf("%v,%f,%f,%v\n", ge.Name, x, y, offset)
	//update met storm name
	err = s.metModel.UpdateStormName(ge.Name)
	if err != nil {
		return s.metModel, ge, err
	}
	//update storm center
	err = s.metModel.UpdateStormCenter(fmt.Sprintf("%f", x), fmt.Sprintf("%f", y))
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
func (s *WalkSimulation) Walk(eventSeed int64, eventNumber int64) (hms.Met, hms.PrecipGridEvent, error) {
	nvrng := rand.New(rand.NewSource(eventSeed))
	stormSeed := nvrng.Int63()
	//select event
	ge, err := s.gridFile.SelectEventByIndex(eventNumber)
	if err != nil {
		return s.metModel, ge, err
	}
	if err != nil {
		return s.metModel, ge, err
	}
	//update met storm name
	err = s.metModel.UpdateStormName(ge.Name)
	if err != nil {
		return s.metModel, ge, err
	}
	//update timeshift
	//compute offset from control specification
	offset := s.control.ComputeOffset(ge.StartTime)
	fmt.Printf("%v,%v\n", ge.Name, offset)
	err = s.metModel.UpdateTimeShift(fmt.Sprintf("%v", offset))
	if err != nil {
		return s.metModel, ge, err
	}
	err = s.mcaFile.UpdateSeed(stormSeed) //using event seed because we need to use a different seed for each storm
	if err != nil {
		return s.metModel, ge, err
	}
	//read csv bytes into a map.
	rmap, err := hms.ReadCsv(s.csvFile)
	err = s.mcaFile.UpdateRealizations(rmap.Query[ge.Name])
	if err != nil {
		return s.metModel, ge, err
	}
	return s.metModel, ge, nil
}
func (s TranspositionSimulation) GetGridFileBytes(precipevent hms.PrecipGridEvent) []byte {
	return s.gridFile.ToBytes(precipevent)
}
func (s WalkSimulation) GetGridFileBytes(precipevent hms.PrecipGridEvent) []byte {
	return s.gridFile.ToBytes(precipevent)
}
