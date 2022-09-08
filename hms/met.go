package hms

import (
	"fmt"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var PrecipStartKeyword string = "Precip Method Parameters:"
var PrecipEndKeyword string = "End:"
var PrecipGridNameKeyword string = "Precip Grid Name: "
var StormCenterXKeyword string = "Storm Center X-coordinate: "
var StormCenterYKeyword string = "Storm Center Y-coordinate: "
var TimeShiftKeyword string = "Time Shift: " //in minutes Negative is FORWARD.

type Met struct {
	metString string
	PrecipMethodParameters
}
type PrecipMethodParameters struct {
	lines []string
}

func ReadMet(metResource plugin.ResourceInfo) (Met, error) {
	//read bytes
	//loop through and find met and precip blocks
	bytes, err := plugin.DownloadObject(metResource)
	if err != nil {
		return Met{}, err
	}
	metfilestring := string(bytes)
	lines := strings.Split(metfilestring, "\n") //maybe rn?
	foundPrecipMethod := false
	metString := ""
	metModel := Met{}
	var precipMethod PrecipMethodParameters
	preciplines := make([]string, 0)
	for _, l := range lines {

		if strings.Contains(l, PrecipStartKeyword) {
			foundPrecipMethod = true
			precipMethod = PrecipMethodParameters{}
			lines = append(preciplines, l)
		}
		if !foundPrecipMethod {
			metString = fmt.Sprintf("%v%v", metString, l)
		} else {
			lines = append(preciplines, l)
		}
	}
	metModel.metString = metString
	precipMethod.lines = lines
	metModel.PrecipMethodParameters = precipMethod
	return metModel, nil
}
func (m Met) UpdateStormCenter(x string, y string) error {
	return nil
}
func (m Met) UpdateTimeShift(x string, y string) error {
	return nil
}
func (m Met) Write(outRI plugin.ResourceInfo) error {
	//write a met model.
	return nil
}

/*
Meteorology: 2011-07-26 event transpose
     Last Modified Date: 12 July 2022
     Last Modified Time: 15:14:18
     Version: 4.11
     Unit System: English
     Set Missing Data to Default: Yes
     Precipitation Method: Gridded Precipitation
     Air Temperature Method: None
     Atmospheric Pressure Method: None
     Dew Point Method: None
     Wind Speed Method: None
     Shortwave Radiation Method: None
     Longwave Radiation Method: None
     Snowmelt Method: None
     Evapotranspiration Method: No Evapotranspiration
End:

Precip Method Parameters: Gridded Precipitation
     Last Modified Date: 12 August 2022
     Last Modified Time: 18:48:30
     Precip Grid Name: 2011-07-26 event
     Storm Center X-coordinate: 361986
     Storm Center Y-coordinate: 2016990
	 Time Shift: -60
End:
*/
