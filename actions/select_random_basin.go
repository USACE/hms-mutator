package actions

import (
	"fmt"
	"math/rand"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/cc-go-sdk/plugin"
)

//the objective of this action is to randomize the basin utilized in an HMS compute
//this allows the basin parameterization to be randomized and the anticedent conditions to be randomized.
//for simplicty, the process is based on an indexed list of basin files, that are selected randomly
//downloaded to the container, and then uploaded with a new name to the event ouptut destination.

type SelectBasinAction struct {
	action   cc.Action
	seedSet  plugin.SeedSet
	inputDS  cc.DataSource
	outputDS cc.DataSource
}

func InitSelectBasinAction(action cc.Action, seedSet plugin.SeedSet, inputDs cc.DataSource, outputDS cc.DataSource) *SelectBasinAction {
	sba := SelectBasinAction{
		action:   action,
		seedSet:  seedSet,
		inputDS:  inputDs,
		outputDS: outputDS,
	}
	return &sba
}
func (sba SelectBasinAction) Compute() error {
	//get range of basin scenarios (ints between 0 and n?)
	maxbasinid := sba.action.Parameters.GetIntOrFail("maxBasinId")
	basinExtension := sba.action.Parameters.GetStringOrFail("basinExtension")
	targetBasinFileName := sba.action.Parameters.GetStringOrFail("targetBasinFileName")
	controlExtension := sba.action.Parameters.GetStringOrFail("controlExtension")
	targetControlFileName := sba.action.Parameters.GetStringOrFail("targetControlFileName")
	//generate a natural variabiilty seed generator
	rng := rand.New(rand.NewSource(sba.seedSet.EventSeed))

	//sample an int in the range of basin scenarios
	sampledBasinId := rng.Int31n(int32(maxbasinid) + 1) //0 to exclusive upper bound
	//download the file from filesapi
	pm, err := cc.InitPluginManager()
	if err != nil {
		return err
	}
	inDS := sba.inputDS
	inDSRoot := inDS.DataPaths[0]
	inDS.DataPaths[0] = fmt.Sprintf("%v/%v.%v", inDSRoot, string(sampledBasinId), basinExtension)
	basinbytes, err := pm.GetFile(sba.inputDS, 0)
	if err != nil {
		return err
	}
	//upload the file to filesapi with the appropriate new name.
	outDS := sba.outputDS
	outDSRoot := outDS.DataPaths[0]
	outDS.DataPaths[0] = fmt.Sprintf("%v/%v.%v", outDSRoot, targetBasinFileName, basinExtension)
	err = pm.PutFile(basinbytes, sba.outputDS, 0)
	if err != nil {
		return err
	}

	inDS.DataPaths[0] = fmt.Sprintf("%v/%v.%v", inDSRoot, string(sampledBasinId), controlExtension)
	controlbytes, err := pm.GetFile(sba.inputDS, 0)
	if err != nil {
		return err
	}
	//upload the file to filesapi with the appropriate new name.
	outDS.DataPaths[0] = fmt.Sprintf("%v/%v.%v", outDSRoot, targetControlFileName, controlExtension)
	err = pm.PutFile(controlbytes, sba.outputDS, 0)
	if err != nil {
		return err
	}
	return nil
}
