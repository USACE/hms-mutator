package hms

type Control struct {
	Name string
}

//find the start date and time and compare it to the start date and time of the grid. calculate an offset to input into the met file
/*
Control: 2014 Event
     Last Modified Date: 13 April 2022
     Last Modified Time: 19:37:31
     Version: 4.11
     Time Zone ID: America/Chicago
     Time Zone GMT Offset: -21600000
     Start Date: 27 June 2014
     Start Time: 24:00
     End Date: 6 July 2014
     End Time: 24:00
     Time Interval: 60
End:
*/
func (c Control) ComputeOffset(gridStartDateTime string) int {
	//parse input as DDMMMYYYY:HHMM //24 hour clocktime
	//parse control start date and time.
	//compute offset set negative for pushing grid startDateTime into the future set positive to bring grid set into the past
	return 0
}
