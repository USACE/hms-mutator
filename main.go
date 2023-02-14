package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/transposition"
	"github.com/usace/wat-go-sdk/plugin"
)

var pluginName string = "hms-mutator"

func main() {

	fmt.Println("hms-mutator!")
	pm, err := cc.InitPluginManager()
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      "could not initiate plugin manager",
		})
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
	if len(pm.GetPayload().Outputs) < 3 {
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      fmt.Sprint("expecting at least 3 outputs to be defined, found", len(payload.Outputs)),
		})
		return errors.New(fmt.Sprint("expecting at least 3 outputs to be defined, found", len(payload.Outputs)))
	}
	if len(payload.Inputs) < 6 {
		err := errors.New(fmt.Sprint("expecting at least 6 inputs to be defined, found ", len(payload.Inputs)))
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      fmt.Sprint("expecting at least 6 inputs to be defined, found ", len(payload.Inputs)),
		})
		return err
	}
	pm.GetFile()
	var mcaRI cc.DataSource
	var eventConfigRI cc.DataSource
	var gridRI cc.DataSource
	var metRI cc.DataSource
	var trgpkgRI cc.DataSource
	var wbgpkgRI cc.DataSource
	var controlRI cc.DataSource
	foundMca := false
	foundEventConfig := false
	foundGrid := false
	foundMet := false
	foundTrGpkg := false
	foundWbGpkg := false
	foundControl := false
	for _, rfd := range payload.Inputs {
		if strings.Contains(rfd.Name, ".mca") {
			foundMca = true
			mcaRI = rfd
		}
		if strings.Contains(rfd.Name, "eventconfiguration.json") {
			foundEventConfig = true
			eventConfigRI = rfd
		}
		if strings.Contains(rfd.Name, ".grid") {
			gridRI = rfd
			foundGrid = true
		}
		if strings.Contains(rfd.Name, ".met") {
			metRI = rfd
			foundMet = true
		}
		if strings.Contains(rfd.Name, "TranspositionRegion.gpkg") {
			trgpkgRI = rfd
			foundTrGpkg = true
		}
		if strings.Contains(rfd.Name, "WatershedBoundary.gpkg") {
			wbgpkgRI = rfd
			foundWbGpkg = true
		}
		if strings.Contains(rfd.Name, ".control") {
			controlRI = rfd
			foundControl = true
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
		msg := "no *.mca file detected, variability is only reflected in storm and storm positioning."
		pm.LogMessage(cc.Message{
			Message: msg,
		})
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
		err := fmt.Errorf("could not find gpkg file for transposition region TranspositionBoundary.gpkg")
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return err
	}
	if !foundWbGpkg {
		err := fmt.Errorf("could not find gpkg file for watershed boundary WatershedBoundary.gpkg")
		pm.LogError(cc.Error{
			ErrorLevel: cc.FATAL,
			Error:      err.Error(),
		})
		return err
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
	seedSet, err := ec.SeedSet(seedSetName) //ec.Seeds[seedSetName]
	if err != nil {
		pm.LogError(cc.Error{
			ErrorLevel: cc.ERROR,
			Error:      fmt.Errorf("no seeds found by name of %v", seedSetName).Error(),
		})
		return err
	}
	//initialize simulation
	sim, err := transposition.InitSimulation(trgpkgRI, wbgpkgRI, metRI, gridRI, controlRI)
	if err != nil {
		logError(err, payload)
		return err
	}
	//compute simulation for given seed set
	m, ge, err := sim.Compute(ss)
	if err != nil {
		logError(err, payload)
		return err
	}
	//update mca file if present
	mca := hms.Mca{}
	if foundMca {
		mca, err := hms.ReadMca(mcaRI)
		if err != nil {
			logError(err, payload)
			return err
		}
		mca.UpdateSeed(seedSet.EventSeed)
	}
	//get met file bytes
	mbytes, err := m.WriteBytes()
	if err != nil {
		logError(err, payload)
		return err
	}
	//find the right resource locations
	var mori plugin.ResourceInfo  //met file output
	var mcori plugin.ResourceInfo //met file output
	var dori plugin.ResourceInfo  //dss file output
	var gori plugin.ResourceInfo  //grid file output
	foundMori := false
	foundMcori := false
	foundDori := false
	foundGori := false
	for _, rfd := range payload.Outputs {
		if strings.Contains(rfd.FileName, ".grid") {
			gori = rfd.ResourceInfo
			foundGori = true
		}
		if strings.Contains(rfd.FileName, ".met") {
			mori = rfd.ResourceInfo
			foundMori = true
		}
		if strings.Contains(rfd.FileName, ".dss") {
			dori = rfd.ResourceInfo
			foundDori = true
		}
		if foundMca {
			if strings.Contains(rfd.FileName, ".mca") {
				mcori = rfd.ResourceInfo
				foundMcori = true
			}
		}
	}
	//upload updated met files.
	if foundMori {
		err = plugin.UpLoadFile(mori, mbytes)
		if err != nil {
			logError(err, payload)
			return err
		}
	} else {
		err := fmt.Errorf("could not find output met file destination")
		logError(err, payload)
		return err
	}
	//if foundMca is true, then foundMcori might be true... upload updated mca file if foundMcori is true.
	if foundMcori { //optional
		err = mca.UploadToS3(mcori)
		logError(err, payload)
		return err
	}

	//upload correct dss file
	if foundDori {
		dssFileName, err := ge.OriginalDSSFile()
		if err != nil {
			logError(err, payload)
			return err
		}
		//leverage the hms project directory structure
		projectPathParts := strings.Split(metRI.Path, "\\")
		dssPath := ""
		for i := 0; i < len(projectPathParts)-1; i++ {
			dssPath = fmt.Sprintf("%v\\%v", dssPath, projectPathParts[i])
		}
		dssPath = fmt.Sprintf("%v\\%v", dssPath, dssFileName)
		projectRI := plugin.ResourceInfo{
			Store: metRI.Store,
			Root:  metRI.Root,
			Path:  dssPath,
		}
		err = ge.DownloadAndUploadDSSFile(projectRI, dori)
		if err != nil {
			logError(err, payload)
			return err
		}
		//update the dss file output to match the ouptut destination
		ge.UpdateDSSFile(dori.Path)
	} else {
		err := fmt.Errorf("could not find output storms.dss file destination")
		logError(err, payload)
		return err
	}

	//upload updated grid files.
	if foundGori {
		err = sim.UploadGridFile(gori, ge)
		if err != nil {
			logError(err, payload)
			return err
		}
	} else {
		err := fmt.Errorf("could not find output grid file destination")
		logError(err, payload)
		return err
	}

	plugin.Log(plugin.Message{
		Status:    plugin.SUCCEEDED,
		Progress:  100,
		Level:     plugin.INFO,
		Message:   "hms mutator complete",
		Sender:    pluginName,
		PayloadId: payload.Id,
	})
	return nil
}
