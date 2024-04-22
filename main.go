package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/usace/cc-go-sdk"
	"github.com/usace/cc-go-sdk/plugin"
	"github.com/usace/hms-mutator/actions"
	"github.com/usace/hms-mutator/hms"
)

var pluginName string = "hms-mutator"

func main() {

	fmt.Println("hms-mutator!")
	pm, err := cc.InitPluginManager()
	if err != nil {
		fmt.Println("could not initiate plugin manager")
		return
	}
	// get the payload.
	payload := pm.GetPayload()
	//validate the payload
	err = validatePayload(payload)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
	}
	//download all required files as bytes
	gridFileBytes, err := getInputBytes("HMS Model", ".grid", payload, pm)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	metFileBytes, err := getInputBytes("HMS Model", ".met", payload, pm)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	controlFileBytes, err := getInputBytes("HMS Model", ".control", payload, pm)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	foundMCA := false
	mcaFileBytes, err := getInputBytes("HMS Model", ".control", payload, pm)
	if err != nil {
		pm.LogMessage(cc.Message{
			Message: "no *.mca file detected, variability is only reflected in storm selection and storm positioning in space and time.",
		})
	} else {
		foundMCA = true
	}
	seedFileBytes, err := getInputBytes("seeds", "", payload, pm)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	seedSet, err := readSeedFile(seedFileBytes)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	transpositionDomainBytes, err := getInputBytes("TranspositionRegion", "", payload, pm)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	watershedDomainBytes, err := getInputBytes("WatershedBoundary", "", payload, pm)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return
	}
	gridFile, err := hms.ReadGrid(gridFileBytes)
	metFile, err := hms.ReadMet(metFileBytes)
	controlFile, err := hms.ReadControl(controlFileBytes)
	mcaFile := hms.Mca{}
	if foundMCA {
		mcaFile, err = hms.ReadMca(mcaFileBytes)
	}

	for _, a := range payload.Actions {
		switch a.Name {
		case "select_random_basin":
			basinDS, err := pm.GetInputDataSource("Input_Basin_Directory")
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      err.Error(),
				})
				return
			}
			outBasinDS, err := pm.GetOutputDataSource("Output_Basin_Directory")
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      err.Error(),
				})
				return
			}
			srb := actions.InitSelectBasinAction(a, seedSet, basinDS, outBasinDS)
			srb.Compute()

		case "single_stochastic_transposition":
			sst := actions.InitSingleStochasticTransposition(pm, gridFile, metFile, controlFile, foundMCA, mcaFile, seedSet, transpositionDomainBytes, watershedDomainBytes)

			output, err := sst.Compute()
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not compute payload",
				})
				return
			}
			dssGridCacheDataSource, err := pm.GetInputDataSource("DSS Grid Cache")
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not find DSS Grid Cache datasource",
				})
				return
			}
			root := path.Dir(dssGridCacheDataSource.DataPaths[0])
			stormDataSource := cc.DataSource{
				Name:      "DssFile",
				ID:        &uuid.NameSpaceDNS,
				Paths:     []string{fmt.Sprintf("%v/%v", root, output.StormName)},
				DataPaths: []string{},
				StoreName: dssGridCacheDataSource.StoreName,
			}
			dssBytes, err := pm.GetFile(stormDataSource, 0)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not find storm",
				})
				return
			}
			err = putOutputBytes(dssBytes, "Storm DSS File", payload, pm)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not put storm",
				})
				return
			}
			err = putOutputBytes(output.GridBytes, "Grid File", payload, pm)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not put grid file",
				})
				return
			}
			err = putOutputBytes(output.MetBytes, "Met File", payload, pm)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not put grid file",
				})
				return
			}
			if foundMCA {
				err = putOutputBytes(output.McaBytes, "MCA File", payload, pm)
				if err != nil {
					pm.LogError(cc.Error{
						ErrorLevel: cc.FATAL,
						Error:      "could not put grid file",
					})
					return
				}
			}
		case "stratified_locations":
			sla, err := actions.InitStratifiedCompute(a, gridFile, transpositionDomainBytes, watershedDomainBytes, payload.Outputs[0])
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not initalize stratified locations for this payload",
				})
				return
			}
			output, err := sla.Compute()
			//put the output

			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not compute stratified locations for this payload",
				})
				return
			}
			locations, err := pm.GetOutputDataSource("Locations")
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not put stratified locations for this payload",
				})
				return
			}
			pm.PutFile(output.CandiateLocations.ToBytes(), locations, 0)
			gridFileOutput, err := pm.GetOutputDataSource("GridFile")
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not put gridfiles for this payload",
				})
				return
			}
			root := path.Dir(gridFileOutput.Paths[0])
			for k, v := range output.GridFiles {
				gridFileOutput.Paths[0] = fmt.Sprintf("%v/%v.grid", root, k)
				pm.PutFile(v, gridFileOutput, 0)
			}
		case "valid_stratified_locations":
			sla, err := actions.InitStratifiedCompute(a, gridFile, transpositionDomainBytes, watershedDomainBytes, payload.Outputs[0])
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not initalize valid stratified locations for this payload",
				})
				return
			}
			inputSource, err := pm.GetInputDataSource("Cumulative Grids")
			output, err := sla.DetermineValidLocations(inputSource) //update to be based on output location?
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not compute valid stratified locations for this payload",
				})
				return
			}
			outputDataSource, err := pm.GetOutputDataSource("ValidLocations")
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.FATAL,
					Error:      "could not put valid stratified locations for this payload",
				})
				return
			}
			root := path.Dir(outputDataSource.DataPaths[0])
			for k, v := range output.StormMap {
				outputDataSource.Paths[0] = fmt.Sprintf("%v/%v.csv", root, k)
				pm.PutFile(v.ToBytes(), outputDataSource, 0)
			}
			outputDataSource.Paths[0] = fmt.Sprintf("%v/%v.csv", root, "AllStormsAllLocations")
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
			pm.PutFile(outbytes, outputDataSource, 0)
		}
	}
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      "could not compute payload",
		})
		return
	} else {
		pm.ReportProgress(cc.StatusReport{
			Status:   "complete",
			Progress: 100,
		})
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
			index := 0
			if extension != "" {
				for i, Path := range input.Paths {
					if strings.Contains(Path, extension) {
						index = i
					}
				}

			}
			return pm.GetFile(input, index)
		}
	}
	return returnBytes, nil
}
func putOutputBytes(data []byte, keyword string, payload cc.Payload, pm *cc.PluginManager) error {
	output, err := pm.GetOutputDataSource(keyword)
	if err != nil {
		return err
	}
	err = pm.PutFile(data, output, 0)
	if err != nil {
		return err
	}
	return nil
}

func readSeedFile(seedFileBytes []byte) (plugin.SeedSet, error) {
	//read event configuration
	var ec plugin.EventConfiguration
	var seedSet plugin.SeedSet
	err := json.Unmarshal(seedFileBytes, &ec)
	if err != nil {
		return seedSet, err
	}
	seedSetName := pluginName
	seedSet, ssok := ec.Seeds[seedSetName]
	if !ssok {
		return seedSet, errors.New("could not find seed set for seedset name")
	}
	return seedSet, nil
}
