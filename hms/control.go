package hms

import (
	"fmt"
	"strings"
	"time"

	"github.com/usace/wat-go-sdk/plugin"
)

var ControlKeyword string = "Control: "
var StartDateKeyword string = "     Start Date: " //DD FULLMONTHNAME YYYY
var StartTimeKeyword string = "     Start Time: " //HH:MM hours in 24 hour clock

type Control struct {
	Name      string
	StartDate string
	StartTime string
}

func ReadControl(controlRI []byte) (Control, error) {
	//read bytes
	//loop through and find startdate and start time
	bytes, err := plugin.DownloadObject(controlRI)
	if err != nil {
		return Control{}, err
	}
	controlstring := string(bytes)
	lines := strings.Split(controlstring, "\r\n") //maybe rn?
	control := Control{}
	for _, l := range lines {
		if strings.Contains(l, ControlKeyword) {
			control.Name = strings.TrimLeft(l, ControlKeyword)
		}
		if strings.Contains(l, StartDateKeyword) {
			control.StartDate = strings.TrimLeft(l, StartDateKeyword)
		}
		if strings.Contains(l, StartTimeKeyword) {
			control.StartTime = strings.TrimLeft(l, StartTimeKeyword)
			if control.StartTime == "24:00" {
				control.StartTime = "23:59"
			}
		}
	}
	return control, nil
}

func (c Control) ComputeOffset(gridStartDateTime string) int {
	//parse input as DDMMMYYYY:HHMM //24 hour clocktime
	//	Jan 2 15:04:05 2006 MST - reference
	gsdt, err := time.Parse("02Jan2006:1504", gridStartDateTime)
	if err != nil {
		fmt.Println(err)
	}
	//parse control start date and time.
	//DD FULLMONTHNAME YYYY
	//HH:MM hours in 24 hour clock
	fulltime := fmt.Sprint(c.StartDate, " ", c.StartTime)
	csdt, err := time.Parse("02 January 2006 15:04", fulltime)
	if err != nil {
		fmt.Println(err)
	}
	//compute offset set negative for pushing grid startDateTime into the future set positive to bring grid set into the past
	detailOffset := gsdt.Sub(csdt)
	minOffset := detailOffset.Minutes()
	return int(minOffset)
}

//find the start date and time and compare it to the start date and time of the grid. calculate an offset to input into the met file
/*
Control: 2014 Event
     Last Modified Date: 13 April 2022
     Last Modified Time: 19:37:31
     Version:ResourceInfo
     Time Zone ID: America/Chicago
     Time Zone GMT Offset: -21600000
     Start Date: 27 June 2014
     Start Time: 24:00
     End Date: 6 July 2014
     End Time: 24:00
     Time Interval: 60
End:
*/
