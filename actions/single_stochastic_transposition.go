package actions

import (
	"time"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/transposition"
	"github.com/usace/hms-mutator/utils"
)

var pluginName string = "hms-mutator"

type SingleStochasticTransposition struct {
	pm                       *cc.PluginManager
	gridFile                 hms.GridFile
	metFile                  hms.Met
	seedSet                  utils.SeedSet
	foundMCA                 bool
	mcaFile                  hms.Mca
	transpositionDomainBytes []byte
	watershedBytes           []byte
}
type StochasticTranspositionResult struct {
	MetBytes  []byte
	McaBytes  []byte
	GridBytes []byte
	StormName string
}

func InitSingleStochasticTransposition(pm *cc.PluginManager, gridFile hms.GridFile, metFile hms.Met, foundMCA bool, mcaFile hms.Mca, seedSet utils.SeedSet, tbytes []byte, wbytes []byte) SingleStochasticTransposition {
	return SingleStochasticTransposition{
		pm:                       pm,
		gridFile:                 gridFile,
		metFile:                  metFile,
		seedSet:                  seedSet,
		foundMCA:                 foundMCA,
		mcaFile:                  mcaFile,
		transpositionDomainBytes: tbytes,
		watershedBytes:           wbytes,
	}
}
func (sst SingleStochasticTransposition) Compute(bootstrapCatalog bool, bootstrapCatalogLength int, normalize bool, controlStartTime time.Time, userSpecifiedOffset int) (StochasticTranspositionResult, error) {
	//initialize simulation
	var ge hms.PrecipGridEvent
	var te hms.TempGridEvent
	var m hms.Met
	var mca hms.Mca
	var gfbytes []byte
	var originalDssPath string
	sim, err := transposition.InitTranspositionSimulation(sst.transpositionDomainBytes, sst.watershedBytes, sst.metFile, sst.gridFile)
	if err != nil {
		sst.pm.Logger.Error(err.Error())
		return StochasticTranspositionResult{}, err
	}
	//compute simulation for given seed set
	m, ge, te, err = sim.Compute(sst.seedSet.EventSeed, sst.seedSet.RealizationSeed, bootstrapCatalog, bootstrapCatalogLength)
	if err != nil {
		sst.pm.Logger.Error(err.Error())
		return StochasticTranspositionResult{}, err
	}
	//update mca file if present
	if sst.foundMCA {
		sst.mcaFile.UpdateSeed(sst.seedSet.EventSeed)
	}
	originalDssPath, _ = ge.OriginalDSSFile()
	//update the dss file output to match the agreed upon convention /data/Storm.dss
	ge.UpdateDSSFile("Storm")
	te.UpdateDSSFile("Storm")
	gfbytes = sim.GetGridFileBytes(ge, te)
	geStartTime, err := time.Parse("02Jan2006:1504", ge.StartTime)
	if err != nil {
		sst.pm.Logger.Error(err.Error())
		return StochasticTranspositionResult{}, err
	}
	//get met file bytes
	m.UpdatePrecipTimeShift(normalize, controlStartTime, geStartTime, userSpecifiedOffset)
	mbytes, err := m.WriteBytes()
	if err != nil {
		sst.pm.Logger.Error(err.Error())
		return StochasticTranspositionResult{}, err
	}
	mcaBytes := make([]byte, 0)
	if sst.foundMCA {
		mcaBytes = mca.ToBytes()
	}
	// prepare result
	result := StochasticTranspositionResult{
		MetBytes:  mbytes,
		McaBytes:  mcaBytes,
		GridBytes: gfbytes,
		StormName: originalDssPath,
	}
	return result, nil
	//find the right resource locations
}
