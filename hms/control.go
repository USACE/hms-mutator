package hms

import (
	"fmt"
	"strings"
	"time"
)

var ControlKeyword string = "Control: "
var StartDateKeyword string = "     Start Date: " //DD FULLMONTHNAME YYYY
var StartTimeKeyword string = "     Start Time: " //HH:MM hours in 24 hour clock

type Control struct {
	Name      string
	StartDate string
	StartTime string
	bytes     []byte
}

func ReadControl(controlRI []byte) (Control, error) {
	//read bytes
	//loop through and find startdate and start time
	controlstring := string(controlRI)
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

		}
	}
	if control.StartTime == "24:00" {
		fulltime := fmt.Sprint(control.StartDate, " 00:00")
		csdt, err := time.Parse("2 January 2006 15:04", fulltime)
		if err != nil {
			fmt.Println(err)
			return Control{}, err
		}
		csdt = csdt.Add(time.Hour * 24)
		control.StartTime = csdt.Format("15:04") //fmt.Sprintf("%v:%v",csdt.Hour(),csdt.Minute())
		control.StartDate = csdt.Format("2 January 2006")
	}
	control.bytes = controlRI
	return control, nil
}
func (c *Control) StartDateAndTime() (time.Time, error) {
	fulltime := fmt.Sprint(c.StartDate, " ", c.StartTime)
	return time.Parse("2 January 2006 15:04", fulltime)
}
func (c *Control) AddHoursToStart(timeWindowModifier int) (time.Time, error) {
	fmt.Printf("adding %v hours to start\n", timeWindowModifier)
	//fmt.Println(c)
	//parse control start date and time.
	//DD FULLMONTHNAME YYYY
	//HH:MM hours in 24 hour clock
	fulltime := fmt.Sprint(c.StartDate, " ", c.StartTime)
	csdt, err := time.Parse("2 January 2006 15:04", fulltime)
	if err != nil {
		fmt.Println(err)
		return time.Now(), err
	}
	hours, err := time.ParseDuration(fmt.Sprintf("%vh", timeWindowModifier))
	if err != nil {
		fmt.Println(err)
		return time.Now(), err
	}
	csdt = csdt.Add(hours)
	c.StartTime = csdt.Format("15:04") //fmt.Sprintf("%v:%v",csdt.Hour(),csdt.Minute())
	c.StartDate = csdt.Format("2 January 2006")
	//fmt.Println(c)
	return csdt, nil
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
	csdt, err := time.Parse("2 January 2006 15:04", fulltime)
	if err != nil {
		fmt.Println(err)
	}
	//compute offset set negative for pushing grid startDateTime into the future set positive to bring grid set into the past
	detailOffset := gsdt.Sub(csdt)
	minOffset := detailOffset.Minutes()
	return int(minOffset)
}
func (c Control) ToBytes() []byte {
	controlstring := string(c.bytes)
	inlines := strings.Split(controlstring, "\r\n") //maybe rn?
	outlines := ""
	for _, l := range inlines {
		line := l
		if strings.Contains(l, StartDateKeyword) {
			line = fmt.Sprintf("%v%v", StartDateKeyword, c.StartDate)
		}
		if strings.Contains(l, StartTimeKeyword) {
			line = fmt.Sprintf("%v%v", StartTimeKeyword, c.StartTime)
		}
		outlines = fmt.Sprint(outlines, line, "\r\n")
	}
	return []byte(outlines)
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
