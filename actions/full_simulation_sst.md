# full-simulation-sst
The full simulation sst action takes an input of a set of storm typed storms, fishnets, basin files, and seasonality distributions (by storm type) and generates a recordset output of the selected storm, new x and y location, selected storm start date, and selected basin file dataset.
This output dataset is used by the hms-runner to govern the storms simulated by hms.

# implementation details
This implementation is based on an input of a repository of fishnet files which store a list of coordinates named by storm name. This fishnet by storm name represents all viable placment locations for each storm. 
This implementation is based on an input of a repository of storms with their storm name and type.
This implementation is based on an input of a repository of basin files representeing the antecedent conditions
This implementation is based on an input of seasonality distributions for each storm type defining valid weighted day of year values for each storm as a cdf. 
# process flow
This action will generate a full simulation sized set of storms, placements, and antecedent conditions.

1. read in all storm names (from files api just get the contents of the catalog.)
2. select a storm with uniform probability
3. define x and y location (predefine fishnet at 1km or 4km possibly, unique to each storm name.)
4. evaluate storm type (should be in the storm name from the selected storm.)
5. use f(st)=>date (should be a set of emperical distributions of date ranges per st)
6. sample year uniformly (should be based on a start and end date of the por, making sure the date is contained.)
7. sample calibration event id (should be 1-6 options.)
8. get basin file name (should be `44*6*365` combinations.)
9. store in tiledb database or dump to csv

# configuration
The configuration for this action is fully defined in the action itself, no global attributes or inputs/outputs are necessary.
## action attributes:
```
		"attributes": {
			"output_data_source": "storms",
			"storms_directory": "model-library/ffrd-trinity/conformance/storm-catalog/storms/",
			"storms_store": "FFRD",
			"fishnet_directory": "model-library/ffrd-trinity/conformance/storm-catalog/fishnets/",
			"fishnet_store": "FFRD",
			"fishnet_type_or_name": "name",
			"storm_type_seasonality_distribution_directory": "model-library/ffrd-trinity/conformance/storm-catalog/seasonality_distributions/",
			"storm_type_seasonality_distribution_store":"FFRD",
			"basin_root_directory": "data/basinmodels",
			"basin_name": "trinity",
			"por_start_date": "19791001",
			"por_end_date": "20220930",
			"calibration_event_names": [
				"apr_may_1990", 
				"aug_sep_2017", 
				"dec_1991", 
				"may_jun_2015", 
				"nov_dec_2015", 
				"oct_nov_2015"],
			"seed_datasource_key": "seeds",
			"blocks_datasource_key": "blocks",
			"sampling_method": "best_estimate"
		}
```
-  ouptut_data_source: the name of the output datasource where the storm recordset data will be stored.
-  storms_direcory: the directory as a path where the storms (in .dss format) are stored.
-  storms_store: the store name for the storms directory
-  fishnet_directory: the directory as a path where the fishnets (in .csv format) are stored.
-  fishnet_store: the store name for the fishnet directory
-  fishnet_type_or_name: allows user the ability to choose fishnets to be defined by storm type or by storm name. 
-  storm_type_seasonality_distribution_directory: the directory of seasonality distributions defined by storm type in csv format
-  storm_type_seasonality_distribution_store: the store name for the storm type seasonality distributions
-  basin_root_directory: the directory relative to the .hms file where basin files are stored
-  basin_name: the name of the hms basin
-  por_start_date: defines the date range where valid storm start dates can begin
-  por_end_date: defines the date range where valid storm start dates can end
-  calibration_event_names: defines the string name of calibration datasets used, will be used in conjunction with the basin root directory, por start and end dates to construct a fully qualified path to a basin for each storm.
-  seed_datasource_key: the name of the seed datasource
-  blocks_datasource_key: the name of the blocks datasource
-  sampling_method: the sampling type bootstrap, jackknife or best_estimate (default)

## inputs
No environment variables are needed
No global attributes are required
A store with the name defined in the storms_store, fishnet_store, and storm_type_seasonality_distribution_store must exist.
An input datasource for seeds and blocks are required and their names must match the attributes in the action.
## outputs
An output datasource named consistently with the output_data_source msut be defined in the action outputs. 

## stores
The inputs are largely expected to be in the format of dss or csv files, seeds can be a tile db dataset. If the output_datasource_name refers to a datasource that is associated with a TileDB store it will store in that store. If S3 is used, a csv will be generated.
