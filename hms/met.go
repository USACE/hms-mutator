package hms

import (
	"errors"
	"fmt"
	"strings"
)

var PrecipStartKeyword string = "Precip Method Parameters:"
var PrecipEndKeyword string = "End:"
var TempStartKeyword string = "Air Temperature Method Parameters:"
var TempEndKeyword string = "End:"
var PrecipGridNameKeyword string = "     Precip Grid Name: "
var TempGridNameKeyword string = "     Temperature Grid Name: "
var StormCenterXKeyword string = "     Storm Center X-coordinate: "
var StormCenterYKeyword string = "     Storm Center Y-coordinate: "
var TimeShiftKeyword string = "     Time Shift: " //in minutes Negative is FORWARD.
var TimeShiftMethodKeyword string = "     Time Shift Method: "

type Met struct {
	metString string
	PrecipMethodParameters
	tempmethod PrecipMethodParameters
}
type PrecipMethodParameters struct {
	lines []string
}

func ReadMet(metResource []byte) (Met, error) {

	metfilestring := string(metResource)
	lines := strings.Split(metfilestring, "\r\n") //maybe rn?
	foundPrecipMethod := false
	foundPrecipEnd := false
	foundTempMethod := false
	foundTempEnd := false
	foundNormalize := false
	metString := ""
	metModel := Met{}
	var precipMethod PrecipMethodParameters
	var tempMethod PrecipMethodParameters
	preciplines := make([]string, 0)
	templines := make([]string, 0)
	for _, l := range lines {

		if strings.Contains(l, PrecipStartKeyword) {
			foundPrecipMethod = true
			precipMethod = PrecipMethodParameters{}
		}
		if strings.Contains(l, TempStartKeyword) {
			foundTempMethod = true
			tempMethod = PrecipMethodParameters{}
		}
		if !foundPrecipMethod {
			if !foundTempMethod {
				if metString == "" {
					metString = l
				} else {
					metString = fmt.Sprintf("%v\r\n%v", metString, l)
				}
			}
		} else if !foundTempMethod {
			if !foundPrecipMethod {
				if metString == "" {
					metString = l
				} else {
					metString = fmt.Sprintf("%v\r\n%v", metString, l)
				}
			}
		}

		if foundPrecipMethod {
			if strings.Contains(l, TimeShiftMethodKeyword) {
				foundNormalize = strings.Contains(l, "NORMALIZE")
			}
			if strings.Contains(l, PrecipEndKeyword) {
				foundPrecipEnd = true
				foundPrecipMethod = false
				if !foundNormalize {
					return metModel, errors.New(`Did not find Normalize as timeshift method.`)
				}
			}
			if !foundPrecipEnd {
				preciplines = append(preciplines, l)
			}
		}
		if foundTempMethod {

			if strings.Contains(l, TempEndKeyword) {
				foundTempEnd = true
				foundTempMethod = false
			}
			if !foundTempEnd {
				templines = append(templines, l)
			}
		}
	}
	metModel.metString = metString
	precipMethod.lines = preciplines
	tempMethod.lines = templines
	metModel.PrecipMethodParameters = precipMethod
	metModel.tempmethod = tempMethod
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
	for idx, l := range m.tempmethod.lines {
		if strings.Contains(l, TempGridNameKeyword) {
			m.tempmethod.lines[idx] = fmt.Sprintf("%v%v", TempGridNameKeyword, stormName)
		}
	}
	return nil
}

/*
	func (m *Met) UpdateTimeShift(timeShift string) error {
		foundTimeShift := false
		foundTimeShiftMethod := false
		for idx, l := range m.PrecipMethodParameters.lines {
			if strings.Contains(l, TimeShiftKeyword) {
				foundTimeShift = true
				m.PrecipMethodParameters.lines[idx] = fmt.Sprintf("%v%v", TimeShiftKeyword, timeShift)
			} else {
				if strings.Contains(l, TimeShiftMethodKeyword) {
					m.PrecipMethodParameters.lines[idx] = fmt.Sprintf("%v%v", TimeShiftMethodKeyword, "NORMALIZE")
					foundTimeShiftMethod = true
				}
			}
		}
		if !foundTimeShift {
			m.PrecipMethodParameters.lines = append(m.PrecipMethodParameters.lines, fmt.Sprintf("%v%v", TimeShiftKeyword, timeShift))
		}
		if !foundTimeShiftMethod {
			m.PrecipMethodParameters.lines = append(m.PrecipMethodParameters.lines, fmt.Sprintf("%v%v", TimeShiftMethodKeyword, "NORMALIZE"))
		}
		foundTimeShift = false
		foundTimeShiftMethod = false
		for idx, l := range m.tempmethod.lines {
			if strings.Contains(l, TimeShiftKeyword) {
				foundTimeShift = true
				m.tempmethod.lines[idx] = fmt.Sprintf("%v%v", TimeShiftKeyword, timeShift)
			} else {
				if strings.Contains(l, TimeShiftMethodKeyword) {
					m.tempmethod.lines[idx] = fmt.Sprintf("%v%v", TimeShiftMethodKeyword, "NORMALIZE")
					foundTimeShiftMethod = true
				}
			}
		}
		if !foundTimeShift {
			m.tempmethod.lines = append(m.tempmethod.lines, fmt.Sprintf("%v%v", TimeShiftKeyword, timeShift))
		}
		if !foundTimeShiftMethod {
			m.tempmethod.lines = append(m.tempmethod.lines, fmt.Sprintf("%v%v", TimeShiftMethodKeyword, "NORMALIZE"))
		}
		return nil
	}
*/
func (m Met) WriteBytes() ([]byte, error) {
	//write a met model.
	filestring := m.metString
	for _, l := range m.PrecipMethodParameters.lines {
		filestring = fmt.Sprintf("%v\r\n%v", filestring, l)
	}
	filestring = fmt.Sprintf("%v\r\n%v\r\n", filestring, "End:")
	if len(m.tempmethod.lines) > 2 {
		for _, l := range m.tempmethod.lines {
			filestring = fmt.Sprintf("%v\r\n%v", filestring, l)
		}
		filestring = fmt.Sprintf("%v\r\n%v\r\n", filestring, "End:")

	}
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
