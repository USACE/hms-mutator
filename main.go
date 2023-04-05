package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/usace/cc-go-sdk"
	"github.com/usace/cc-go-sdk/plugin"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/transposition"
)

var pluginName string = "hms-mutator"

func main() {

	fmt.Println("hms-mutator!")
	pm, err := cc.InitPluginManager()
	if err != nil {
		fmt.Println("could not initiate plugin manager")
		return
	}
	err = computePayload(pm)
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
func computePayload(pm *cc.PluginManager) error {
	payload := pm.GetPayload()
	useActualStormName := false
	stormNameAttribute, ok := payload.Attributes["use_actual_storm_name"]
	if ok {
		useActualStormName = stormNameAttribute.(bool)
	}
	updateRealizationNumbers := false
	updateRealizationAttribute, rok := payload.Attributes["update_realization_numbers"]
	if rok {
		updateRealizationNumbers = updateRealizationAttribute.(bool)
	}
	walkGrids := false
	walkGridsAttribute, gok := payload.Attributes["walk_grids"]
	if gok {
		walkGrids = walkGridsAttribute.(bool)
	}
	expectedOutputs := 3
	expectedInputs := 4 //hms model (grid, met, control), watershed boundary, transposition region, seeds
	if updateRealizationNumbers && walkGrids {
		//need the mca file too
		//need the realization numbers file (in the hms model)
		//dont need transposition nor watershed boundary
		expectedInputs = 3
	}
	if len(pm.GetPayload().Outputs) < expectedOutputs {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      fmt.Sprintf("expecting at least %v outputs to be defined, found %v", expectedOutputs, len(payload.Outputs)),
		})
		return errors.New(fmt.Sprintf("expecting at least %v outputs to be defined, found %v", expectedOutputs, len(payload.Outputs)))
	}
	if len(payload.Inputs) < expectedInputs {
		err := errors.New(fmt.Sprintf("expecting at least %v inputs to be defined, found %v", expectedInputs, len(payload.Inputs)))
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      fmt.Sprintf("expecting at least %v inputs to be defined, found %v", expectedInputs, len(payload.Inputs)),
		})
		return err
	}

	var mcaRI cc.DataSource
	mcaIdx := 0
	var eventConfigRI cc.DataSource
	var gridRI cc.DataSource
	gridIdx := 0
	var metRI cc.DataSource
	metIdx := 0
	var trgpkgRI cc.DataSource
	var wbgpkgRI cc.DataSource
	var controlRI cc.DataSource
	controlIdx := 0
	foundMca := false
	foundEventConfig := false
	foundGrid := false
	foundMet := false
	foundTrGpkg := false
	foundWbGpkg := false
	foundControl := false
	csvIdx := 0
	var csvRI cc.DataSource
	foundCsv := false
	for _, rfd := range payload.Inputs {
		if strings.Contains(rfd.Name, "HMS Model") {
			for idx, path := range rfd.Paths {
				if strings.Contains(path, ".grid") {
					gridIdx = idx
					gridRI = rfd
					foundGrid = true
				}
				if strings.Contains(path, ".met") {
					metIdx = idx
					metRI = rfd
					foundMet = true
				}
				if strings.Contains(path, ".control") {
					controlIdx = idx
					controlRI = rfd
					foundControl = true
				}
				if strings.Contains(path, ".mca") {
					mcaIdx = idx
					mcaRI = rfd
					foundMca = true
				}
				if strings.Contains(path, ".csv") {
					csvIdx = idx
					csvRI = rfd
					foundCsv = true
				}
			}

		}

		if strings.Contains(rfd.Name, "seeds") {
			foundEventConfig = true
			eventConfigRI = rfd
		}
		if strings.Contains(rfd.Name, "TranspositionRegion") {
			trgpkgRI = rfd
			foundTrGpkg = true
		}
		if strings.Contains(rfd.Name, "WatershedBoundary") {
			wbgpkgRI = rfd
			foundWbGpkg = true
		}

	}
	if !foundEventConfig {
		err := fmt.Errorf("could not find event configuration to find the proper seeds to run sst")
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return err
	}
	if !foundMca {
		if updateRealizationNumbers {
			msg := "no *.mca file detected, cannot update relaization numbers, aborting."
			pm.LogMessage(cc.Message{
				Message: msg,
			})
			return errors.New("requested to update realization numbers but no mca file was found")
		}
		msg := "no *.mca file detected, variability is only reflected in storm selection and storm positioning in space and time."
		pm.LogMessage(cc.Message{
			Message: msg,
		})
	}
	if updateRealizationNumbers {
		if !foundCsv {
			err := fmt.Errorf("could not find csv files with the storms and realization numbers")
			pm.LogError(cc.Error{
				ErrorLevel: cc.FATAL,
				Error:      err.Error(),
			})
			return err
		}
	}
	if !foundGrid {
		err := fmt.Errorf("could not find grid file for storm definitions")
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return err
	}
	if !foundMet {
		err := fmt.Errorf("could not find met file for meterologic conditions")
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return err
	}
	if !foundTrGpkg {
		if updateRealizationNumbers && walkGrids {
			msg := "no transposition region found, but the attributes for walk grids and update realization numbers are set to true, attempting to continue"
			pm.LogMessage(cc.Message{
				Message: msg,
			})
		} else {
			err := fmt.Errorf("could not find gpkg file for transposition region TranspositionBoundary.gpkg")
			pm.LogError(cc.Error{
				ErrorLevel: cc.FATAL,
				Error:      err.Error(),
			})
			return err
		}

	}
	if !foundWbGpkg {
		if updateRealizationNumbers && walkGrids {
			msg := "no watershed boundary found, but the attributes for walk grids and update realization numbers are set to true, attempting to continue"
			pm.LogMessage(cc.Message{
				Message: msg,
			})
		} else {
			err := fmt.Errorf("could not find gpkg file for watershed boundary WatershedBoundary.gpkg")
			pm.LogError(cc.Error{
				ErrorLevel: cc.FATAL,
				Error:      err.Error(),
			})
			return err
		}
	}
	if !foundControl {
		err := fmt.Errorf("could not find control file for timewindow specifications")
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return err
	}
	//read event configuration
	var ec plugin.EventConfiguration
	eventConfigurationReader, err := pm.FileReader(eventConfigRI, 0)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      err.Error(),
		})
		return err
	}
	defer eventConfigurationReader.Close()
	err = json.NewDecoder(eventConfigurationReader).Decode(&ec)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      err.Error(),
		})
		return err
	}

	seedSetName := pluginName
	seedSet, ssok := ec.Seeds[seedSetName]
	if !ssok {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("no seeds found by name of %v", seedSetName).Error(),
		})
		return err
	}
	//initialize simulation
	trbytes, err := pm.GetFile(trgpkgRI, 0)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("could not get bytes for transposition region").Error(),
		})
		return err
	}
	wbgpkgbytes, err := pm.GetFile(wbgpkgRI, 0)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("could not get bytes for watershed boundary").Error(),
		})
		return err
	}
	metbytes, err := pm.GetFile(metRI, metIdx)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("could not get bytes for met file").Error(),
		})
		return err
	}
	gridbytes, err := pm.GetFile(gridRI, gridIdx)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("could not get bytes for grid file").Error(),
		})
		return err
	}
	controlbytes, err := pm.GetFile(controlRI, controlIdx)
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("could not get bytes for control file").Error(),
		})
		return err
	}
	var ge hms.PrecipGridEvent
	var m hms.Met
	var mca hms.Mca
	var csvbytes []byte
	var gfbytes []byte
	if walkGrids && updateRealizationNumbers {
		if foundMca {
			mcabytes, err := pm.GetFile(mcaRI, mcaIdx)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.ERROR,
					Error:      err.Error(),
				})
				return err
			}
			if foundCsv {
				csvbytes, err = pm.GetFile(csvRI, csvIdx)
				if err != nil {
					pm.LogError(cc.Error{
						ErrorLevel: cc.ERROR,
						Error:      err.Error(),
					})
					return err
				}
			}
			sim, err := transposition.InitWalkSimulation(metbytes, gridbytes, controlbytes, mcabytes, csvbytes)
			m, ge, err = sim.Walk(seedSet.EventSeed, int64(pm.EventNumber()))
			gfbytes = sim.GetGridFileBytes(ge)
		}
	} else {
		sim, err := transposition.InitTranspositionSimulation(trbytes, wbgpkgbytes, metbytes, gridbytes, controlbytes)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
		//compute simulation for given seed set
		m, ge, err = sim.Compute(seedSet.EventSeed, seedSet.RealizationSeed)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
		//update mca file if present
		if foundMca {
			mcabytes, err := pm.GetFile(mcaRI, mcaIdx)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.ERROR,
					Error:      err.Error(),
				})
				return err
			}
			mca, err := hms.ReadMca(mcabytes)
			if err != nil {
				pm.LogError(cc.Error{
					ErrorLevel: cc.ERROR,
					Error:      err.Error(),
				})
				return err
			}
			mca.UpdateSeed(seedSet.EventSeed)
		}
		gfbytes = sim.GetGridFileBytes(ge)
	}

	//get met file bytes
	mbytes, err := m.WriteBytes()
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      err.Error(),
		})
		return err
	}
	//find the right resource locations
	var mori cc.DataSource  //met file output
	var mcori cc.DataSource //met file output
	var dori cc.DataSource  //dss file output
	var gori cc.DataSource  //grid file output
	foundMori := false
	foundMcori := false
	foundDori := false
	foundGori := false
	for _, rfd := range payload.Outputs {
		if strings.Contains(rfd.Name, "Grid File") {
			gori = rfd
			foundGori = true
		}
		if strings.Contains(rfd.Name, "Met File") {
			mori = rfd
			foundMori = true
		}
		if strings.Contains(rfd.Name, "Storm DSS File") {
			dori = rfd
			foundDori = true
		}
		if foundMca {
			if strings.Contains(rfd.Name, "MCA File") {
				mcori = rfd
				foundMcori = true
			}
		}
	}
	//upload updated met files.
	if foundMori {
		err = pm.PutFile(mbytes, mori, 0)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
	} else {
		err := fmt.Errorf("could not find output met file destination")
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
		return err
	}
	//if foundMca is true, then foundMcori might be true... upload updated mca file if foundMcori is true.
	if foundMcori { //optional
		bytes := mca.ToBytes()
		err = pm.PutFile(bytes, mcori, 0)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
		return err
	}

	//upload correct dss file
	if foundDori {
		dssFileName, err := ge.OriginalDSSFile()
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
		//leverage the hms project directory structure
		//danger zone!!!! this is risky because the data could be missing
		//this is to reduce the input data soruce specification for each event in the directory
		//there could be hundreds of storms in the hms model and we dont want to copy them all
		//just the one that is selected.
		projectPathParts := strings.Split(gridRI.Paths[gridIdx], "/")
		dssPath := ""
		for i := 0; i < len(projectPathParts)-1; i++ {
			dssPath = fmt.Sprintf("%v/%v", dssPath, projectPathParts[i])
		}
		dssPath = fmt.Sprintf("%v/%v", dssPath, dssFileName)
		dssPath = strings.Replace(dssPath, "\\", "/", -1)
		ds := cc.DataSource{
			Name:      dssFileName,
			ID:        &uuid.NameSpaceDNS,
			Paths:     []string{dssPath},
			StoreName: metRI.StoreName,
		}
		dssBytes, err := pm.GetFile(ds, 0)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
		err = pm.PutFile(dssBytes, dori, 0)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}

		if useActualStormName {
			ge.UpdateDSSFile(ge.Name)
		} else {
			//update the dss file output to match the agreed upon convention /data/Storm.dss
			ge.UpdateDSSFile("Storm")
		}

	} else {
		err := fmt.Errorf("could not find output storms.dss file destination")
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
	}
	//upload updated grid files.
	if foundGori {

		pm.PutFile(gfbytes, gori, 0)
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
	} else {
		err := fmt.Errorf("could not find output grid file destination")
		if err != nil {
			pm.LogError(cc.Error{
				ErrorLevel: cc.ERROR,
				Error:      err.Error(),
			})
			return err
		}
	}

	return nil
}
