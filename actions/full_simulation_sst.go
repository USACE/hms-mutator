package actions

import (
	"fmt"
	"math/rand"
	"strings"

	"time"

	"github.com/usace/cc-go-sdk"
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
type FullRealizationResult []EventResult

type EventResult struct {
	EventNumber int64   `eventstore:"event_number"`
	StormPath   string  `eventstore:"event_number"`
	X           float64 `eventstore:"x"`
	Y           float64 `eventstore:"y"`
	StormType   string  `eventstore:"storm_type"`
	StormDate   string  `eventstore:"storm_date"`
	BasinPath   string  `eventstore:"basin_path"`
}

func (frsst FullRealizationSST) Compute(realizationNumber int32, pm *cc.PluginManager) error {
	a := frsst.action
	//get parameters
	///get storms
	stormDirectory := a.Attributes.GetStringOrFail("storms_directory")
	stormsStoreKey := a.Attributes.GetStringOrFail("storms_store") //expecting this to be an s3 bucket?
	stormList, err := utils.ListAllPaths(a.IOManager, stormsStoreKey, stormDirectory, "*.dss")
	if err != nil {
		return err
	}
	//if i wanted to bootstrap, i could bootstrap the storm list now...

	///use fishnets to figure out placements - select from list of valid placements.
	fishnetDirectory := a.Attributes.GetStringOrFail("fishnet_directory")
	fishnetStoreKey := a.Attributes.GetStringOrFail("fishnet_store")
	fishnetList, err := utils.ListAllPaths(a.IOManager, fishnetStoreKey, fishnetDirectory, "*.csv")
	if err != nil {
		return err
	}
	fishNetMap, err := utils.ReadFishNets(a.IOManager, fishnetStoreKey, fishnetList)
	if err != nil {
		return err
	}
	//storm type seasonality distributions
	stormTypeSeasonalityDistributionDirectory := a.Attributes.GetStringOrFail("storm_type_seasonality_distibution_directory")
	stormTypeSeasonalityDistributionStoreKey := a.Attributes.GetStringOrFail("storm_type_seasonality_distibution_store")
	stormTypeDistributionList, err := utils.ListAllPaths(a.IOManager, stormTypeSeasonalityDistributionStoreKey, stormTypeSeasonalityDistributionDirectory, "*.csv")
	if err != nil {
		return err
	}
	stormTypeSeasonalityDistributionsMap, err := utils.ReadStormDistributions(a.IOManager, stormTypeSeasonalityDistributionStoreKey, stormTypeDistributionList)
	if err != nil {
		return err
	}
	//time range of POR
	porStartDateString := a.Attributes.GetStringOrFail("por_start_date")
	porStartDate, err := time.Parse("20060102", porStartDateString)
	if err != nil {
		return err
	}
	porEndDateString := a.Attributes.GetStringOrFail("por_end_date")
	porEndDate, err := time.Parse("20060102", porEndDateString)
	if err != nil {
		return err
	}
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
	seeds, err := utils.ReadSeedsFromTiledb(a.IOManager, seedInput.StoreName, seedsKey, "hms-mutator") //hms-mutator is a const in main, but i dont want to create cycles.
	if err != nil {
		return err
	}
	//event/block/simulation relationship
	blocksKey := a.Attributes.GetStringOrFail("blocks_datasource_key")
	blocksInput, err := a.GetInputDataSource(blocksKey) //expecting this to be tiledb
	if err != nil {
		return err
	}
	blocks, err := utils.ReadBlocksFromTiledb(pm, blocksInput.StoreName, blocksInput.Name)
	if err != nil {
		return err
	}
	results, err := compute(realizationNumber, stormList, calibrationEvents, fishNetMap, stormTypeSeasonalityDistributionsMap, porStartDate, porEndDate, seeds, blocks)
	if err != nil {
		return err
	}
	//write results to data stores
	return writeResultsToTileDB(pm, blocksInput.StoreName, results, "storms") //update this to not referenceblock store, and also not hardcode the name to "storms"
}
func compute(realizationNumber int32, stormNames []string, calibrationEventNames []string, fishnets utils.FishNetMap, seasonalDistributions utils.StormTypeSeasonalityDistributionMap, porStart time.Time, porEnd time.Time, seeds []utils.SeedSet, blocks []utils.Block) (FullRealizationResult, error) {
	results := make(FullRealizationResult, 0)
	for _, b := range blocks {
		if b.RealizationIndex == realizationNumber {
			if b.BlockEventCount > 0 {
				for en := b.BlockEventStart; en <= b.BlockEventEnd; en++ {
					//create random number generator for event
					enRng := rand.New(rand.NewSource(seeds[en].EventSeed))
					//sample storm name
					stormName := stormNames[enRng.Intn(len(stormNames))]
					//calculate storm type from storm name
					stormType := strings.Split(stormName, "_")[2] //assuming yyyymmdd_xxhr_data-type_storm-type_storm-rank - if data-type is dropped as i hope this needs to be updated to 2
					//sample calibration event
					calibrationEvent := calibrationEventNames[enRng.Intn(len(calibrationEventNames))]
					//fetch fishnet based on storm name
					fishnet, ok := fishnets[stormName] //storm name is full path not just file name.
					if !ok {
						return results, fmt.Errorf("could not find storm name %v in fishnet map", stormName)
					}
					//sample location
					coordinate := fishnet.Coordinates[enRng.Intn(len(fishnet.Coordinates))]
					//fetch seasonal distribution based on storm type
					seasonalDistribution, ok := seasonalDistributions[stormType]
					if !ok {
						return results, fmt.Errorf("could not find the seasonal distribution for type %v", stormType)
					}
					//fetch day of year
					dayOfYear := seasonalDistribution.Sample(enRng.Float64())
					//determine year.
					yearCount := porEnd.Year() - porStart.Year() //this needs to be checked on both ends for valid dates.
					dayofyearInrange := false
					year := 0
					for !dayofyearInrange {
						initalYearGuess := enRng.Intn(yearCount) + porStart.Year()
						if initalYearGuess == porStart.Year() {
							if dayOfYear >= porStart.YearDay() {
								dayofyearInrange = true
								year = initalYearGuess
							}
						} else if initalYearGuess == porEnd.Year() {
							if dayOfYear <= porEnd.YearDay() {
								dayofyearInrange = true
								year = initalYearGuess
							}
						}
					}
					//create start date from day of year and year
					startDate := time.Date(year, 1, 1, 1, 1, 1, 1, time.Local)
					//convert day of year to duration
					sdur := fmt.Sprintf("%vh", (dayOfYear-1)*24)
					dur, err := time.ParseDuration(sdur)
					if err != nil {
						return results, err
					}
					startDate = startDate.Add(dur)
					event := EventResult{
						EventNumber: en,
						StormPath:   stormName,
						StormType:   stormType,
						X:           coordinate.X,
						Y:           coordinate.Y,
						BasinPath:   fmt.Sprintf("%v_%v", startDate.Format("20060102"), calibrationEvent), //need to add root
					}
					results = append(results, event)
				}

			}

		}
	}
	return results, nil
}
func writeResultsToTileDB(pm *cc.PluginManager, storeKey string, results FullRealizationResult, tableName string) error {
	recordset, err := cc.NewEventStoreRecordset(pm, &results, storeKey, tableName)
	if err != nil {
		return err
	}
	err = recordset.Create()
	if err != nil {
		return err
	}
	return recordset.Write(&results)
}
