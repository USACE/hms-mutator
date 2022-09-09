package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

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

	if len(payload.Outputs) != 5 {
		err := errors.New(fmt.Sprint("expecting 5 outputs to be defined, found", len(payload.Outputs)))
		logError(err, payload)
		return err
	}
	if len(payload.Inputs) != 1 {
		err := errors.New(fmt.Sprint("expecting 1 input to be defined, found ", len(payload.Inputs)))
		logError(err, payload)
		return err
	}
	var eventConfigRI plugin.ResourceInfo
	var gridRI plugin.ResourceInfo
	var metRI plugin.ResourceInfo
	var gpkgRI plugin.ResourceInfo
	var controlRI plugin.ResourceInfo
	foundEventConfig := false
	foundGrid := false
	foundMet := false
	foundGpkg := false
	foundControl := false
	for _, rfd := range payload.Inputs {
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
	m, err := sim.Compute(ss)
	if err != nil {
		logError(err, payload)
		return err
	}
	//get met file bytes
	mbytes, err := m.WriteBytes()
	if err != nil {
		logError(err, payload)
		return err
	}
	//upload updated files.
	err = plugin.UpLoadFile(payload.Outputs[0].ResourceInfo, mbytes)
	if err != nil {
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
