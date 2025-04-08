package actions

import (
	"errors"

	"github.com/HydrologicEngineeringCenter/go-statistics/statistics"
	"github.com/usace/cc-go-sdk"
	"github.com/usace/filesapi"
	"github.com/usace/hms-mutator/utils"
)

/*
This action will generate a full realization sized set of storms, placements, and antecedent conditions.
//steps:
1. read in all storm names (from files api just get the contents of the catalog.)
2. select a storm with uniform probability
3. define x and y location (predefine fishnet at 1km or 4km possibly, unique to each storm name.)
4. evaluate storm type (should be in the storm name from the selected storm.)
5. use f(st)=>date (should be a set of emperical distributions of date ranges per st)
6. sample year uniformly (should be based on a start and end date of the por, making sure the date is contained.)
7. sample calibration event id (should be 1-6 options.)
8. get basin file name (should be `44*6*365` combinations.)
9. store in tiledb database or dump to csv
*/
type FullRealizationSST struct {
	action cc.Action
}
type FullRealizationResult struct {
	Events []EventResult
}
type FishNetMap map[string]utils.CoordinateList //storm type coordinate list.
type StormTypeSeasonalDistributions map[string]statistics.EmpiricalDistribution
type EventResult struct {
	EventNumber int64
	StormPath   string
	X           float64
	Y           float64
	StormType   string
	StormDate   string
	BasinPath   string
}

func (frsst FullRealizationSST) Compute(realizationNumber int) error {
	a := frsst.action
	//get parameters
	///get storms
	stormDirectory := a.Attributes.GetStringOrFail("storms_directory")
	stormsStoreKey := a.Attributes.GetStringOrFail("storms_store") //expecting this to be an s3 bucket?
	stormList, err := listAllPaths(a.IOManager, stormsStoreKey, stormDirectory, "*.dss")
	if err != nil {
		return err
	}
	//if i wanted to bootstrap, i could bootstrap the storm list now...

	///use fishnets to figure out placements - select from list of valid placements.
	fishnetDirectory := a.Attributes.GetStringOrFail("fishnet_directory")
	fishnetStoreKey := a.Attributes.GetStringOrFail("fishnet_store")
	fishnetList, err := listAllPaths(a.IOManager, fishnetStoreKey, fishnetDirectory, "*.csv")
	if err != nil {
		return err
	}
	fishNetMap, err := readFishNets(a.IOManager, fishnetStoreKey, fishnetList)
	//storm type seasonality distributions
	stormTypeSeasonalityDistributionDirectory := a.Attributes.GetStringOrFail("storm_type_seasonality_distibution_directory")
	//time range of POR
	porStartDate := a.Attributes.GetStringOrFail("por_start_date")
	porEndDate := a.Attributes.GetStringOrFail("por_end_date")
	//calibration event strings
	calibrationEvents, err := a.Attributes.GetStringSlice("calibration_event_names")
	if err != nil {
		return err
	}
	//seeds
	seedsKey := a.Attributes.GetStringOrFail("seed_datasource_key")
	seedInput, err := a.GetInputDataSource(seedsKey) //expecting this to be a tiledb dense array
	if err != nil {
		return err
	}
	//event/block/simulation relationship
	blocksKey := a.Attributes.GetStringOrFail("blocks_datasource_key")
	blocksInput, err := a.GetInputDataSource(blocksKey) //expecting this to be tiledb
	if err != nil {
		return err
	}
	return nil
}

func listAllPaths(ioManager cc.IOManager, StoreKey string, DirectoryKey string, filter string) ([]string, error) {
	store, err := ioManager.GetStore(StoreKey)
	var pathList []string
	if err != nil {
		return pathList, err
	}
	session, ok := store.Session.(cc.S3DataStore)
	if !ok {
		return pathList, errors.New("storms_store was not an s3datastore type")
	}
	rawSession, ok := session.GetSession().(filesapi.FileStore)
	if !ok {
		return pathList, errors.New("could not convert s3datastore raw session into filestore type")
	}
	pageIdx := 0 //does page index start with 0 or 1?
	input := filesapi.ListDirInput{
		Path:   filesapi.PathConfig{Path: DirectoryKey},
		Page:   pageIdx,
		Size:   filesapi.DEFAULTMAXKEYS,
		Filter: filter,
	}
	for {
		fapiresult, err := rawSession.ListDir(input)
		if err != nil {
			//check if there are files in the list?
			return pathList, err
		}
		list := *fapiresult
		for _, s := range list {
			pathList = append(pathList, s.Path)
		}
		if len(list) < 1000 {
			return pathList, nil
		} else {
			pageIdx++
		}
	}
}
func readFishNets(iomanager cc.IOManager, storeKey string, filePaths []string) (FishNetMap, error) {
	FishNetMap := make(map[string]utils.CoordinateList)
	return FishNetMap, nil
}
