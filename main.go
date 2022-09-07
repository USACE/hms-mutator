package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var string pluginName = "hms-mutator"

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

	if len(payload.Outputs) != 1 {
		err := errors.New(fmt.Sprint("expecting one output to be defined found", len(payload.Outputs)))
		logError(err, payload)
		return err
	}
	if len(payload.Inputs) != 3 {
		err := errors.New(fmt.Sprint("expecting 2 inputs to be defined found ", len(payload.Inputs)))
		logError(err, payload)
		return err
	}
	var gpkgRI plugin.ResourceInfo
	var depthGridRI plugin.ResourceInfo
	foundDepthGrid := false
	foundGPKG := false
	for _, rfd := range payload.Inputs {
		if strings.Contains(rfd.FileName, ".tif") {
			depthGridRI = rfd.ResourceInfo
			foundDepthGrid = true
		}
		if strings.Contains(rfd.FileName, ".gpkg") {
			gpkgRI = rfd.ResourceInfo
			foundGPKG = true
		}
	}

	//output read all bytes
	bytes, err := ioutil.ReadFile(outfp)
	if err != nil {
		logError(err, payload)
		return err
	}
	err = plugin.UpLoadFile(payload.Outputs[0].ResourceInfo, bytes)
	if err != nil {
		logError(err, payload)
		return err
	}
	plugin.Log(plugin.Message{
		Status:    plugin.SUCCEEDED,
		Progress:  100,
		Level:     plugin.INFO,
		Message:   "consequences complete",
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
func writeLocalBytes(b []byte, destinationRoot string, destinationPath string) error {
	if _, err := os.Stat(destinationRoot); os.IsNotExist(err) {
		os.MkdirAll(destinationRoot, 0644) //do i need to trim filename?
	}
	err := os.WriteFile(destinationPath, b, 0644)
	if err != nil {
		plugin.Log(plugin.Message{
			Message: fmt.Sprintf("failure to write local file: %v\n\terror:%v", destinationPath, err),
			Level:   plugin.ERROR,
			Sender:  pluginName,
		})
		return err
	}
	return nil
}
