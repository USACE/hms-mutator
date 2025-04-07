package actions

import (
	"github.com/usace/cc-go-sdk"
)

/*
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

func (fsst FullSimulationSST) Compute() error {
	a := fsst.action
	//get parameters
	stormDirectory := a.Attributes.GetStringOrFail("storms_directory")
	stormsStoreKey := a.Attributes.GetStringOrFail("storms_store") //expecting this to be an s3 bucket?
	//use fishnets to figure out placements - select from list of valid placements.
	fishnetDirectory := a.Attributes.GetStringOrFail("fishnet_directory")
	fishnetStoreKey := a.Attributes.GetStringOrFail("fishnet_store")
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
	seedsKey := a.Attributes.GetStringOrFail("seed_datasource_key")
	seedInput, err := a.GetInputDataSource(seedsKey) //expecting this to be a tiledb dense array
	if err != nil {
		return err
	}

	return nil
}
