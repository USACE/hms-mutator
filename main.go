package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/usace/cc-go-sdk"
	tiledb "github.com/usace/cc-go-sdk/tiledb-store"
	"github.com/usace/hms-mutator/actions"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/utils"
)

var pluginName string = "hms-mutator"

const WORKING_DIRECTORY string = "/data"

func main() {

	fmt.Println("hms-mutator!")
	//register tiledb
	cc.DataStoreTypeRegistry.Register("TILEDB", tiledb.TileDbEventStore{})
	pm, err := cc.InitPluginManager()
	if err != nil {
		fmt.Println("could not initiate plugin manager")
		return
	}
	// get the payload.
	payload := pm.Payload
	//validate the payload
	err = validatePayload(payload)
	if err != nil {
		pm.Logger.Error(err.Error())
	}
	//download all required files as bytes
	gridFileBytes, err := getInputBytes("HMS Model", ".grid", payload, pm)
	if err != nil {
		pm.Logger.Error(err.Error())
		return
	}
	metFileBytes, err := getInputBytes("HMS Model", ".met", payload, pm)
	if err != nil {
		pm.Logger.Error(err.Error())
		return
	}
	foundMCA := false
	mcaFileBytes, err := getInputBytes("HMS Model", ".mca", payload, pm)
	if err != nil {
		err = nil
		pm.Logger.Info("no *.mca file detected, variability is only reflected in storm selection and storm positioning in space and time.")
	} else {
		foundMCA = true
	}
	seedFileBytes, err := getInputBytes("seeds", "", payload, pm)
	if err != nil {
		pm.Logger.Error(err.Error())
		return
	}
	seedSet, err := readSeedFile(seedFileBytes)
	if err != nil {
		pm.Logger.Error(err.Error())
		return
	}
	transpositionDomainBytes, err := getInputBytes("TranspositionRegion", "", payload, pm)
	if err != nil {
		pm.Logger.Error(err.Error())
		return
	}
	watershedDomainBytes, err := getInputBytes("WatershedBoundary", "", payload, pm)
	if err != nil {
		pm.Logger.Error(err.Error())
		return
	}
	gridFile, err := hms.ReadGrid(gridFileBytes)
	metFile, err := hms.ReadMet(metFileBytes)
	//controlFile, err := hms.ReadControl(controlFileBytes)
	mcaFile := hms.Mca{}
	if foundMCA {
		mcaFile, err = hms.ReadMca(mcaFileBytes)
	}
	controlStartTime := time.Now()
	for _, a := range payload.Actions {
		switch a.Type {
		case "select_random_basin":
			basinDS, err := pm.GetInputDataSource("Input_Basin_Directory")
			if err != nil {
				pm.Logger.Error(err.Error())
				return
			}
			outBasinDS, err := pm.GetOutputDataSource("Output_Basin_Directory")
			if err != nil {
				pm.Logger.Error(err.Error())
				return
			}
			srb := actions.InitSelectBasinAction(a, seedSet, basinDS, outBasinDS)
			controlStartTime, err = srb.Compute()
			if err != nil {
				return
			}

		case "single_stochastic_transposition":
			sst := actions.InitSingleStochasticTransposition(pm, gridFile, metFile, foundMCA, mcaFile, seedSet, transpositionDomainBytes, watershedDomainBytes)
			bootstrapCatalogString := a.Attributes.GetStringOrDefault("bootstrap_catalog", "false")
			bootstrapCatalog, err := strconv.ParseBool(bootstrapCatalogString)
			if err != nil {
				pm.Logger.Error("could not parse bootstrap_catalog parameter")
				return
			}
			bootstrapCatalogLength := a.Attributes.GetIntOrDefault("bootstrap_catalog_length", len(gridFile.Events))
			if len(gridFile.Events) < bootstrapCatalogLength {
				pm.Logger.Error("cannot allow bootstrap_catalog_length to be greater than the catalog length")
				return
			}
			normalizeTimeShiftString := a.Attributes.GetStringOrDefault("normalize", "true")
			normalizeTimeShift, err := strconv.ParseBool(normalizeTimeShiftString)
			userSpecifiedOffset := a.Attributes.GetIntOrDefault("start_time_offset", 0)
			if err != nil {
				pm.Logger.Error("could not parse normalize parameter")
				return
			}
			output, err := sst.Compute(bootstrapCatalog, bootstrapCatalogLength, normalizeTimeShift, controlStartTime, userSpecifiedOffset)
			if err != nil {
				pm.Logger.Error("could not compute payload")
				return
			}
			dssGridCacheDataSource, err := pm.GetInputDataSource("DSS Grid Cache")
			if err != nil {
				pm.Logger.Error("could not find DSS Grid Cache datasource")
				return
			}
			root := dssGridCacheDataSource.Paths["default"]
			stormName := strings.Replace(output.StormName, "\\", "/", -1)
			stormDataSource := cc.DataSource{
				Name:      "DssFile",
				ID:        &uuid.NameSpaceDNS,
				Paths:     map[string]string{"default": fmt.Sprintf("%v%v", root, stormName)},
				StoreName: dssGridCacheDataSource.StoreName,
			}
			dssBytes, err := utils.GetFile(*pm, stormDataSource, "default")
			if err != nil {
				pm.Logger.Error("could not find storm")
				return
			}
			err = putOutputBytes(dssBytes, "Storm DSS File", payload, pm)
			if err != nil {
				pm.Logger.Error("could not put storm")
				return
			}
			err = putOutputBytes(output.GridBytes, "Grid File", payload, pm)
			if err != nil {
				pm.Logger.Error("could not put grid file")
				return
			}
			err = putOutputBytes(output.MetBytes, "Met File", payload, pm)
			if err != nil {
				pm.Logger.Error("could not put grid file")
				return
			}
			if foundMCA {
				err = putOutputBytes(output.McaBytes, "MCA File", payload, pm)
				if err != nil {
					pm.Logger.Error("could not put MCA file")
					return
				}
			}
		case "stratified_locations":
			sla, err := actions.InitStratifiedCompute(a, gridFile, transpositionDomainBytes, watershedDomainBytes) //, payload.Outputs[0])
			if err != nil {
				pm.Logger.Error("could not initalize stratified locations for this payload")
				return
			}
			output, err := sla.Compute()
			//put the output

			if err != nil {
				pm.Logger.Error("could not compute stratified locations for this payload")
				return
			}
			locations, err := pm.GetOutputDataSource("Locations")
			if err != nil {
				pm.Logger.Error("could not put stratified locations for this payload")
				return
			}
			utils.PutFile(output.CandiateLocations.ToBytes(), pm.IOManager, locations, "default")
			gridFileOutput, err := pm.GetOutputDataSource("GridFile")
			if err != nil {
				pm.Logger.Error("could not put gridfiles for this payload")
				return
			}
			root := path.Dir(gridFileOutput.Paths["default"])
			for k, v := range output.GridFiles {
				gridFileOutput.Paths["default"] = fmt.Sprintf("%v/%v.grid", root, k)
				utils.PutFile(v, pm.IOManager, gridFileOutput, "default")
			}
		case "valid_stratified_locations":
			sla, err := actions.InitStratifiedCompute(a, gridFile, transpositionDomainBytes, watershedDomainBytes) //, payload.Outputs[0])
			if err != nil {
				pm.Logger.Error("could not initalize valid stratified locations for this payload")
				return
			}
			//inputSource, err := pm.GetInputDataSource("Cumulative Grids")
			outputDataSource, err := a.GetOutputDataSource("ValidLocations")
			if err != nil {
				pm.Logger.Error("could not put valid stratified locations for this payload")
			}
			root := outputDataSource.Paths["default"]
			output, err := sla.DetermineValidLocationsQuickly(pm.IOManager) //sla.DetermineValidLocations(inputSource) //update to be based on output location?
			if err != nil {
				pm.Logger.Error("could not compute valid stratified locations for this payload")
				return
			}

			outputDataSource.Paths["default"] = fmt.Sprintf("%v/%v.csv", root, "AllStormsAllLocations")
			outbytes := make([]byte, 0)
			outbytes = append(outbytes, "StormName,X,Y,IsValid"...)
			//create random list of ints
			indexes := make([]int, len(output.AllStormsAllLocations))
			rand := rand.New(rand.NewSource(945631))
			for i := 0; i < len(indexes); i++ {
				j := rand.Intn(i + 1)
				if i != j {
					indexes[i] = indexes[j]
				}
				indexes[j] = i
			}
			for i, _ := range output.AllStormsAllLocations {
				outbytes = append(outbytes, fmt.Sprintf("%v,%v,%v,%v\n", output.AllStormsAllLocations[indexes[i]].StormName, output.AllStormsAllLocations[indexes[i]].Coordinate.X, output.AllStormsAllLocations[indexes[i]].Coordinate.Y, output.AllStormsAllLocations[indexes[i]].IsValid)...)
			}
			utils.PutFile(outbytes, pm.IOManager, outputDataSource, "default")
		}
	}
	if err != nil {
		fmt.Println(err.Error())
		pm.Logger.Error("could not compute payload")
		return
	} else {
		pm.Logger.Info("complete 100 percent")
	}
}
func validatePayload(payload cc.Payload) error {
	expectedOutputs := 3
	expectedInputs := 4 //hms model (grid, met, control), watershed boundary, transposition region, seeds
	if len(payload.Outputs) < expectedOutputs {
		return errors.New(fmt.Sprintf("expecting at least %v outputs to be defined, found %v", expectedOutputs, len(payload.Outputs)))
	}
	if len(payload.Inputs) < expectedInputs {
		err := errors.New(fmt.Sprintf("expecting at least %v inputs to be defined, found %v", expectedInputs, len(payload.Inputs)))
		return err
	}
	return nil
}
func getInputBytes(keyword string, extension string, payload cc.Payload, pm *cc.PluginManager) ([]byte, error) {
	returnBytes := make([]byte, 0)
	for _, input := range payload.Inputs {
		if strings.Contains(input.Name, keyword) {
			index := "default"
			has := false
			if extension != "" {
				for i, Path := range input.Paths {
					//index, _ := strconv.Atoi(i)
					if strings.Contains(Path, extension) {
						index = i
						has = true
					}
				}
			} else {
				has = true
			}
			if has {
				return utils.GetFile(*pm, input, index)
			} else {
				return returnBytes, errors.New("could not find extension " + extension)
			}

		}
	}
	return returnBytes, errors.New("could not find keyword " + keyword)
}
func putOutputBytes(data []byte, keyword string, payload cc.Payload, pm *cc.PluginManager) error {
	output, err := pm.GetOutputDataSource(keyword)
	if err != nil {
		return err
	}
	err = utils.PutFile(data, pm.IOManager, output, "default")
	if err != nil {
		return err
	}
	return nil
}

func readSeedFile(seedFileBytes []byte) (utils.SeedSet, error) {
	//read event configuration
	var ec []utils.EventConfiguration
	var seedSet utils.SeedSet
	err := json.Unmarshal(seedFileBytes, &ec)
	if err != nil {
		return seedSet, err
	}
	seedSetName := pluginName
	seedinstance := ec[0] //[seedSetName]
	seeds, ssok := seedinstance.Seeds[seedSetName]
	if !ssok {
		return seedSet, errors.New("could not find seed set for seedset name")
	}
	return seeds, nil
}
