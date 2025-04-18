package utils

import (
	"fmt"
	"slices"

	"github.com/usace/cc-go-sdk"
)

// EventConfiguration is a simple structure to support consistency in cc plugins regarding the usage of seeds for natural variability and knowledge uncertainty and realization numbers for indexing
type EventConfiguration struct {
	Seeds map[string]SeedSet `json:"seeds" eventstore:"seeds"`
}
type SeedSet struct {
	EventSeed       int64 `json:"event_seed" eventstore:"event_seed"`
	RealizationSeed int64 `json:"realization_seed" eventstore:"realization_seed"`
}

// seed list is a metadata record called "seed_columns" as a slice of string
// seeds live in a dense array based on "seed_name" with two dimensions, columns being plugins (from the metadata list), rows being events.
// array attributes are "realization_seed" as an int64 and "event_seed" as int64
func ReadSeedsFromTiledb(ioManager cc.IOManager, tileDbStoreName string, datasetName string, seedSetName string) ([]SeedSet, error) {
	seeds := make([]SeedSet, 0)
	//get the metadata
	store, err := ioManager.GetStore(tileDbStoreName)
	if err != nil {
		return seeds, err
	}
	tdbms, ok := store.Session.(cc.MetadataStore)
	if !ok {
		return seeds, fmt.Errorf("the store named %v does not implement metadata store", tileDbStoreName)
	}
	seedNames := make([]string, 0)
	err = tdbms.GetMetadata("seed_columns", &seedNames)
	if err != nil {
		return seeds, err
	}
	columnIndex := slices.Index(seedNames, seedSetName)
	if columnIndex == -1 {
		return seeds, fmt.Errorf("the seed set name %v does not exist in the metadata store under seed_columns", seedSetName)
	}
	//convert store to a dense array store
	tdbmdas, ok := store.Session.(cc.MultiDimensionalArrayStore)
	if !ok {
		return seeds, fmt.Errorf("the store named %v does not implement multidimensional array store", tileDbStoreName)
	}
	getArrayInput := cc.GetArrayInput{
		Attrs:    []string{"realization_seed", "event_seed"}, //does this have to be in the same order as it was written?
		DataPath: datasetName,
		//BufferRange: []int64{0}, //how do i know how big of a buffer to input?
		//SearchOrder: cc.ROWMAJOR,
	}
	result, err := tdbmdas.GetArray(getArrayInput)
	if err != nil {
		return seeds, err
	}
	eventSeeds := make([]int64, 0)
	result.GetColumn(columnIndex, 1, &eventSeeds) //how do i know for certain attribute order?
	realizationSeeds := make([]int64, 0)
	result.GetColumn(columnIndex, 0, &realizationSeeds) //how do i know for certain attribute order?
	for i, es := range eventSeeds {
		seeds = append(seeds, SeedSet{EventSeed: es, RealizationSeed: realizationSeeds[i]})
	}
	return seeds, nil
}
