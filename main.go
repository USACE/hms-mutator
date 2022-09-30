package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/transposition"
	"github.com/usace/wat-go-sdk/plugin"
)

var pluginName string = "hms-mutator"

func main() {

	fmt.Println("hms-mutator!")
	var payloadPath string
	flag.StringVar(&payloadPath, "payload", "pathtopayload.yml", "please specify an input file using `-payload pathtopayload.yml`")
	flag.Parse()
	if payloadPath == "" {
		plugin.Log(plugin.Message{
			Status:    plugin.FAILED,
			Progress:  0,
			Level:     plugin.ERROR,
			Message:   "given a blank path...\n\tplease specify an input file using `payload pathtopayload.yml`",
			Sender:    pluginName,
			PayloadId: "unknown payloadid because the plugin package could not be properly initalized",
		})
		return
	}
	err := plugin.InitConfigFromEnv()
	if err != nil {
		logError(err, plugin.ModelPayload{Id: "unknownpayloadid"})
		return
	}
	payload, err := plugin.LoadPayload(payloadPath)
	if err != nil {
		logError(err, plugin.ModelPayload{Id: "unknownpayloadid"})
		return
	}
	err = computePayload(payload)
	if err != nil {
		logError(err, payload)
		return
	}
}
func computePayload(payload plugin.ModelPayload) error {

	if len(payload.Outputs) < 3 {
		err := errors.New(fmt.Sprint("expecting at least 3 outputs to be defined, found", len(payload.Outputs)))
		logError(err, payload)
		return err
	}
	if len(payload.Inputs) < 5 {
		err := errors.New(fmt.Sprint("expecting at least 5 inputs to be defined, found ", len(payload.Inputs)))
		logError(err, payload)
		return err
	}
	var mcaRI plugin.ResourceInfo
	var eventConfigRI plugin.ResourceInfo
	var gridRI plugin.ResourceInfo
	var metRI plugin.ResourceInfo
	var gpkgRI plugin.ResourceInfo
	var controlRI plugin.ResourceInfo
	foundMca := false
	foundEventConfig := false
	foundGrid := false
	foundMet := false
	foundGpkg := false
	foundControl := false
	for _, rfd := range payload.Inputs {
		if strings.Contains(rfd.FileName, ".mca") {
			mcaRI = rfd.ResourceInfo
			foundMca = true
		}
		if strings.Contains(rfd.FileName, "eventconfiguration.json") {
			eventConfigRI = rfd.ResourceInfo
			foundEventConfig = true
		}
		if strings.Contains(rfd.FileName, ".grid") {
			gridRI = rfd.ResourceInfo
			foundGrid = true
		}
		if strings.Contains(rfd.FileName, ".met") {
			metRI = rfd.ResourceInfo
			foundMet = true
		}
		if strings.Contains(rfd.FileName, ".gpkg") {
			gpkgRI = rfd.ResourceInfo
			foundGpkg = true
		}
		if strings.Contains(rfd.FileName, ".control") {
			controlRI = rfd.ResourceInfo
			foundControl = true
		}
	}
	if !foundEventConfig {
		err := fmt.Errorf("could not find event configuration to find the proper seeds to run sst")
		logError(err, payload)
		return err
	}
	if !foundMca {
		msg := "no *.mca file detected, variability is only reflected in storm and storm positioning."
		plugin.Log(plugin.Message{
			Status:    plugin.COMPUTING,
			Progress:  0,
			Level:     plugin.INFO,
			Message:   msg,
			Sender:    pluginName,
			PayloadId: payload.Id,
		})
	}
	if !foundGrid {
		err := fmt.Errorf("could not find grid file for storm definitions")
		logError(err, payload)
		return err
	}
	if !foundMet {
		err := fmt.Errorf("could not find met file for meterologic conditions")
		logError(err, payload)
		return err
	}
	if !foundGpkg {
		err := fmt.Errorf("could not find gpkg file for transposition region")
		logError(err, payload)
		return err
	}
	if !foundControl {
		err := fmt.Errorf("could not find control file for timewindow specifications")
		logError(err, payload)
		return err
	}
	//read event configuration
	ec, err := plugin.LoadEventConfiguration(eventConfigRI)
	if err != nil {
		logError(err, payload)
		return err
	}
	//obtain seed set
	ss, err := ec.SeedSet(payload.Model.Alternative)
	if err != nil {
		logError(err, payload)
		return err
	}
	//initialize simulation
	sim, err := transposition.InitSimulation(gpkgRI, metRI, gridRI, controlRI)
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
		mca.UpdateSeed(ss.EventSeed)
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
func logError(err error, payload plugin.ModelPayload) {
	plugin.Log(plugin.Message{
		Status:    plugin.FAILED,
		Progress:  0,
		Level:     plugin.ERROR,
		Message:   err.Error(),
		Sender:    pluginName,
		PayloadId: payload.Id,
	})
}
