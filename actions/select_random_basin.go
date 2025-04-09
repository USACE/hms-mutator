package actions

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/cc-go-sdk/plugin"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/utils"
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
func (sba SelectBasinAction) Compute() (time.Time, error) {
	//get range of basin scenarios (ints between 0 and n?)
	maxbasinid := sba.action.Attributes.GetIntOrFail("maxBasinId")
	basinExtension := sba.action.Attributes.GetStringOrFail("basinExtension")
	targetBasinFileName := sba.action.Attributes.GetStringOrFail("targetBasinFileName")
	controlExtension := sba.action.Attributes.GetStringOrFail("controlExtension")
	targetControlFileName := sba.action.Attributes.GetStringOrFail("targetControlFileName")
	//allowing user specified start date to accommodate the inclusion of a setback period.
	updateStartDateAndTime, err := strconv.ParseBool(sba.action.Attributes.GetStringOrFail("updateStartDateAndTime"))
	if err != nil {
		return time.Now(), err
	}
	hoursOffset := sba.action.Attributes.GetIntOrDefault("startDateAndTimeOffset", 0)

	//generate a natural variabiilty seed generator
	rng := rand.New(rand.NewSource(sba.seedSet.EventSeed))

	//sample an int in the range of basin scenarios
	sampledBasinId := rng.Int31n(int32(maxbasinid)) //0 to exclusive upper bound
	//download the file from filesapi
	pm, err := cc.InitPluginManager()
	if err != nil {
		return time.Now(), err
	}
	inDS := sba.inputDS
	inDSRoot := inDS.Paths["default"]
	inDS.Paths["default"] = fmt.Sprintf("%v/%v.%v", inDSRoot, fmt.Sprint(sampledBasinId), basinExtension)
	//fmt.Println(inDS.Paths["default"])
	basinbytes, err := utils.GetFile(*pm, sba.inputDS, "default") //pm.GetFile(sba.inputDS, 0)
	if err != nil {
		return time.Now(), err
	}
	//upload the file to filesapi with the appropriate new name.
	outDS := sba.outputDS
	outDSRoot := outDS.Paths["default"]
	outDS.Paths["default"] = fmt.Sprintf("%v/%v.%v", outDSRoot, targetBasinFileName, basinExtension)
	//fmt.Println(outDS.Paths["default"])
	err = utils.PutFile(basinbytes, *pm, sba.outputDS, "default")
	if err != nil {
		return time.Now(), err
	}

	inDS.Paths["default"] = fmt.Sprintf("%v/%v.%v", inDSRoot, fmt.Sprint(sampledBasinId), controlExtension)
	//fmt.Println(inDS.Paths["default"])
	controlbytes, err := utils.GetFile(*pm, sba.inputDS, "default")
	if err != nil {
		return time.Now(), err
	}
	control, err := hms.ReadControl(controlbytes)
	if err != nil {
		return time.Now(), err
	}
	controltime, err := control.StartDateAndTime()
	if err != nil {
		return controltime, err
	}
	if updateStartDateAndTime {
		controltime, err = control.AddHoursToStart(hoursOffset)
		if err != nil {
			return time.Now(), err
		}
		controlbytes = control.ToBytes()
	}

	//upload the file to filesapi with the appropriate new name.
	outDS.Paths["default"] = fmt.Sprintf("%v/%v.%v", outDSRoot, targetControlFileName, controlExtension)
	fmt.Println(outDS.Paths["default"])
	err = utils.PutFile(controlbytes, *pm, sba.outputDS, "default")
	if err != nil {
		return time.Now(), err
	}
	return controltime, nil
}
