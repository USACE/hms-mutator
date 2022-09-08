package hms

var PrecipStartKeyword string = "Precip Method Parameters:"
var PrecipEndKeyword string = "End:"
var PrecipGridNameKeyword string = "Precip Grid Name: "
var StormCenterXKeyword string = "Storm Center X-coordinate: "
var StormCenterYKeyword string = "Storm Center Y-coordinate: "
var TimeShiftKeyword string = "Time Shift: " //in minutes Negative is FORWARD.

type Met struct {
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
