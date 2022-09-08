package hms

import (
	"fmt"
	"strings"

	"github.com/usace/wat-go-sdk/plugin"
)

var PrecipStartKeyword string = "Precip Method Parameters:"
var PrecipEndKeyword string = "End:"
var PrecipGridNameKeyword string = "     Precip Grid Name: "
var StormCenterXKeyword string = "     Storm Center X-coordinate: "
var StormCenterYKeyword string = "     Storm Center Y-coordinate: "
var TimeShiftKeyword string = "     Time Shift: " //in minutes Negative is FORWARD.

type Met struct {
	metString string
	PrecipMethodParameters
	restOfFile string
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
	lines := strings.Split(metfilestring, "\r\n") //maybe rn?
	foundPrecipMethod := false
	foundPrecipEnd := false
	metString := ""
	restOfFileString := ""
	metModel := Met{}
	var precipMethod PrecipMethodParameters
	preciplines := make([]string, 0)
	for _, l := range lines {

		if strings.Contains(l, PrecipStartKeyword) {
			foundPrecipMethod = true
			precipMethod = PrecipMethodParameters{}
		}

		if !foundPrecipMethod {
			if metString == "" {
				metString = l
			} else {
				metString = fmt.Sprintf("%v\r\n%v", metString, l)
			}
		} else {
			if strings.Contains(l, PrecipEndKeyword) {
				foundPrecipEnd = true
			}
			if !foundPrecipEnd {
				preciplines = append(preciplines, l)
			} else {
				if restOfFileString == "" {
					restOfFileString = l
				} else {
					restOfFileString = fmt.Sprintf("%v\r\n%v", restOfFileString, l)
				}

			}

		}
	}
	metModel.metString = metString
	metModel.restOfFile = restOfFileString
	precipMethod.lines = preciplines
	metModel.PrecipMethodParameters = precipMethod
	return metModel, nil
}
func (m *Met) UpdateStormCenter(x string, y string) error {
	foundX := false
	foundY := false
	for idx, l := range m.PrecipMethodParameters.lines {
		if strings.Contains(l, StormCenterXKeyword) {
			foundX = true
			m.PrecipMethodParameters.lines[idx] = fmt.Sprintf("%v%v", StormCenterXKeyword, x)
		}
		if strings.Contains(l, StormCenterYKeyword) {
			foundY = true
			m.PrecipMethodParameters.lines[idx] = fmt.Sprintf("%v%v", StormCenterYKeyword, y)
		}
	}
	if !foundX {
		m.PrecipMethodParameters.lines = append(m.PrecipMethodParameters.lines, fmt.Sprintf("%v%v", StormCenterXKeyword, x))
	}
	if !foundY {
		m.PrecipMethodParameters.lines = append(m.PrecipMethodParameters.lines, fmt.Sprintf("%v%v", StormCenterYKeyword, y))
	}
	return nil
}
func (m *Met) UpdateStormName(stormName string) error {
	for idx, l := range m.PrecipMethodParameters.lines {
		if strings.Contains(l, PrecipGridNameKeyword) {
			m.PrecipMethodParameters.lines[idx] = fmt.Sprintf("%v%v", PrecipGridNameKeyword, stormName)
		}
	}
	return nil
}
func (m *Met) UpdateTimeShift(timeShift string) error {
	foundTimeShift := false
	for idx, l := range m.PrecipMethodParameters.lines {
		if strings.Contains(l, TimeShiftKeyword) {
			foundTimeShift = true
			m.PrecipMethodParameters.lines[idx] = fmt.Sprintf("%v%v", TimeShiftKeyword, timeShift)
		}
	}
	if !foundTimeShift {
		m.PrecipMethodParameters.lines = append(m.PrecipMethodParameters.lines, fmt.Sprintf("%v%v", TimeShiftKeyword, timeShift))
	}
	return nil
}
func (m Met) WriteBytes() ([]byte, error) {
	//write a met model.
	filestring := m.metString
	for _, l := range m.PrecipMethodParameters.lines {
		filestring = fmt.Sprintf("%v\r\n%v", filestring, l)
	}
	filestring = fmt.Sprintf("%v\r\n%v", filestring, m.restOfFile)
	bytes := []byte(filestring)
	return bytes, nil
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
