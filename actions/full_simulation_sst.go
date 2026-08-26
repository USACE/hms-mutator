package actions

import (
	"bytes"
	"errors"
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
type FullSimulationSST struct {
	action cc.Action
}
type FullSimulationResult []EventResult

type EventResult struct {
	EventNumber int64   `eventstore:"event_number"`
	StormPath   string  `eventstore:"storm_path"`
	X           float64 `eventstore:"x"`
	Y           float64 `eventstore:"y"`
	StormType   string  `eventstore:"storm_type"`
	StormDate   string  `eventstore:"storm_date"`
	BasinPath   string  `eventstore:"basin_path"`
}

func InitFullRealizationSST(a cc.Action) *FullSimulationSST {
	return &FullSimulationSST{action: a}
}
func (frsst *FullSimulationSST) Compute(pm *cc.PluginManager) error {
	a := frsst.action
	//get parameters
	//get output datasource
	outputDataSourceKey := a.Attributes.GetStringOrFail("output_data_source")
	outputDataSource, err := a.GetOutputDataSource(outputDataSourceKey)
	if err != nil {
		return err
	}
	///get storms
	stormDirectory := a.Attributes.GetStringOrFail("storms_directory")
	stormsStoreKey := a.Attributes.GetStringOrFail("storms_store") //expecting this to be an s3 bucket?
	stormList, err := utils.ListAllPaths(a.IOManager, stormsStoreKey, stormDirectory, "*.dss")
	if err != nil {
		return err
	}

	samplingMethod := a.Attributes.GetStringOrDefault("sampling_method", "best_estimate")
	var sampler utils.StormSampler
	switch samplingMethod {
	case "best_estimate":
		sampler, err = utils.InitBestEstimateSampler(stormList)
		if err != nil {
			return err
		}
	case "bootstrap":
		sampler, err = utils.InitBootstrapSampler(stormList)
		if err != nil {
			return err
		}
	case "jackknife":
		sampler, err = utils.InitJackknifeSampler(stormList)
		if err != nil {
			return err
		}
	default:

		return errors.New("samlper type not defined " + samplingMethod)
	}

	///use fishnets to figure out placements - select from list of valid placements. fishnets are currently expected to be unique to each storm... could be converted to be unique to each storm type.
	fishnetDirectory := a.Attributes.GetStringOrFail("fishnet_directory")
	fishnetStoreKey := a.Attributes.GetStringOrFail("fishnet_store")
	fishnettypeorname := a.Attributes.GetStringOrFail("fishnet_type_or_name")
	fishnetList, err := utils.ListAllPaths(a.IOManager, fishnetStoreKey, fishnetDirectory, "*.csv")
	if err != nil {
		return err
	}
	fishNetMap, err := utils.ReadFishNets(a.IOManager, fishnetStoreKey, fishnetList, fishnetDirectory)
	if err != nil {
		return err
	}
	//storm type seasonality distributions
	stormTypeSeasonalityDistributionDirectory := a.Attributes.GetStringOrFail("storm_type_seasonality_distribution_directory")
	stormTypeSeasonalityDistributionStoreKey := a.Attributes.GetStringOrFail("storm_type_seasonality_distribution_store")
	stormTypeDistributionList, err := utils.ListAllPaths(a.IOManager, stormTypeSeasonalityDistributionStoreKey, stormTypeSeasonalityDistributionDirectory, "*.csv")
	if err != nil {
		return err
	}
	stormTypeSeasonalityDistributionsMap, err := utils.ReadStormDistributions(a.IOManager, stormTypeSeasonalityDistributionStoreKey, stormTypeDistributionList, stormTypeSeasonalityDistributionDirectory)
	if err != nil {
		return err
	}
	//basin root directory
	basinRootDir := a.Attributes.GetStringOrFail("basin_root_directory")
	basinName := a.Attributes.GetStringOrFail("basin_name")
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
	seeds, err := utils.ReadSeedsFromTiledb(a.IOManager, seedInput.StoreName, "seeds", "hms-mutator") //hms-mutator is a const in main, but i dont want to create cycles.
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
	results, err := compute(stormList, calibrationEvents, basinRootDir, basinName, fishNetMap, fishnettypeorname, stormTypeSeasonalityDistributionsMap, porStartDate, porEndDate, seeds, blocks, sampler)
	if err != nil {
		return err
	}
	//write results to data stores
	if outputDataSource.StoreName == "store" {
		return writeResultsToTileDB(pm, outputDataSource.StoreName, results, outputDataSource.Name) //update this to not referenceblock store, and also not hardcode the name to "storms"
	} else {
		return writeResultsToCSV(a.IOManager, outputDataSource, results)
	}

}
func compute(stormNames []string, calibrationEventNames []string, basinRootDir string, basinName string, fishnets utils.FishNetMap, fishnettypeorname string, seasonalDistributions utils.StormTypeSeasonalityDistributionMap, porStart time.Time, porEnd time.Time, seeds []utils.SeedSet, blocks []utils.Block, sampler utils.StormSampler) (FullSimulationResult, error) {
	results := make(FullSimulationResult, 0)
	//realizationIndex := -1
	for _, b := range blocks {
		//right here i would have logic to determine if the sampler needs to be updated for the list of storms at either the realization or block level
		//sampler.SampleNames(b.BlockEventStart,seeds)//find the right event number at the start of a block
		//or
		//if realizationIndex != int(b.RealizationIndex){
		//	sampler.SampleNames(b.BlockEventStart,seeds) //find the right event number at the start of a block
		//	realizationIndex = int(b.RealizationIndex)
		//}

		if b.BlockEventCount > 0 {
			for en := b.BlockEventStart; en <= b.BlockEventEnd; en++ {
				//create random number generator for event
				if int(en) <= len(seeds) {
					enRng := rand.New(rand.NewSource(seeds[en-1].EventSeed))
					//right here i would have logic to determine if the sampler needs to be updated for the list of storms at the event level
					//sampler.SampleNames(en,seeds)
					//sample storm name
					stormName := sampler.SampleName(enRng) //stormNames[enRng.Intn(len(stormNames))]
					//calculate storm type from storm name
					stormType := strings.Split(stormName, "_")[2] //assuming yyyymmdd_xxhr_storm-type_storm-rank
					//sample calibration event
					calibrationEvent := calibrationEventNames[enRng.Intn(len(calibrationEventNames))]
					//fetch fishnet based on storm name -
					sname := strings.Split(stormName, ".")[0]
					sname = strings.Replace(sname, "st", "ST", -1) //how did this happen?//storm name just file name no extension.
					if fishnettypeorname == "type" {
						sname = strings.Replace(stormType, "st", "ST", -1)
					} else if fishnettypeorname != "name" {
						sname = fishnettypeorname //if not type or name, just use whatever they give directly.
					}
					fishnet, ok := fishnets[sname]
					if !ok {

						return results, fmt.Errorf("could not find fishnet %v in fishnet map", sname)
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
						initalYearGuess := enRng.Intn(yearCount+1) + porStart.Year() //+1 is due to [0,n)
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
						} else if porStart.Year() < initalYearGuess && initalYearGuess < porEnd.Year() {
							dayofyearInrange = true
							year = initalYearGuess
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
						StormDate:   startDate.Format("20060102"),
						BasinPath:   fmt.Sprintf("%v/%v_%v_%v", basinRootDir, startDate.Format("2006-01-02"), basinName, calibrationEvent),
					}
					results = append(results, event)
				}
			}

		}
	}
	return results, nil
}
func writeResultsToTileDB(pm *cc.PluginManager, storeKey string, results FullSimulationResult, tableName string) error {
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
func writeResultsToCSV(iomanager cc.IOManager, ds cc.DataSource, results FullSimulationResult) error {
	//create a header
	data := "event_number,storm_path,x,y,storm_type,storm_date,basin_path"
	for _, r := range results {
		data = fmt.Sprintf("%v\n%v,%v,%v,%v,%v,%v,%v", data, r.EventNumber, r.StormPath, r.X, r.Y, r.StormType, r.StormDate, r.BasinPath)
	}
	bytedata := []byte(data)
	writer := bytes.NewReader(bytedata)
	_, err := iomanager.Put(cc.PutOpInput{
		SrcReader:         writer,
		DataSourceOpInput: cc.DataSourceOpInput{DataSourceName: ds.Name, PathKey: "default"},
	})
	return err
}
