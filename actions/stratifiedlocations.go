package actions

import (
	"fmt"
	"math"
	"path"
	"strings"

	"github.com/dewberry/gdal"
	"github.com/usace/cc-go-sdk"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/utils"
)

//this action is designed to create a set of uniformly distributed points within a bounding box that are within a polygon
//it will also prepare grid files for each storm in the storm catalog and store them in distinct output locations by storm name

type StratifiedCompute struct {
	Spacing                  float64
	GridFile                 hms.GridFile
	TranspositionPolygon     gdal.DataSource //ideally this would be the buffered transposition domain to represent valid transposition locations.
	StudyAreaPolygon         gdal.DataSource
	AcceptanceDepthThreshold float64
}
type StratifiedComputeResult struct {
	CandiateLocations utils.CoordinateList
	GridFiles         map[string][]byte
}
type ValidLocationsComputeResult struct {
	AllStormsAllLocations []LocationInfo
	StormMap              map[string]utils.CoordinateList
}
type LocationInfo struct {
	StormName  string
	Coordinate utils.Coordinate
	IsValid    bool
}

const LOCALDIR = "/app/data/"
const COORDFILE = "locations.csv"

func InitStratifiedCompute(a cc.Action, gridfile hms.GridFile, polygonBytes []byte, watershedbytes []byte, outputDataSource cc.DataSource) (StratifiedCompute, error) {

	//ensure path is local
	fileName := "transpositionpolygon.gpkg"
	filePath := fmt.Sprintf("%v%v", LOCALDIR, fileName)
	err := utils.WriteLocalBytes(polygonBytes, LOCALDIR, filePath)
	if err != nil {
		return StratifiedCompute{}, err
	}
	wfileName := "watershedpolygon.gpkg"
	wfilePath := fmt.Sprintf("%v%v", LOCALDIR, wfileName)
	err = utils.WriteLocalBytes(watershedbytes, LOCALDIR, wfilePath)
	if err != nil {
		return StratifiedCompute{}, err
	}
	tds := gdal.OpenDataSource(filePath, 0)  //defer disposing the datasource and layers.
	wds := gdal.OpenDataSource(wfilePath, 0) //defer disposing the datasource and layers.
	spacing := a.Attributes.GetFloatOrFail("spacing")
	acceptance_threshold := a.Attributes.GetFloatOrFail("acceptance_threshold")
	return StratifiedCompute{Spacing: spacing, GridFile: gridfile, TranspositionPolygon: tds, StudyAreaPolygon: wds, AcceptanceDepthThreshold: acceptance_threshold}, nil
}
func (sc StratifiedCompute) Compute() (StratifiedComputeResult, error) {
	centers, err := sc.generateStormCenters() //still need to upload storm centers to the proper output location specified by the plugin manager.
	if err != nil {
		return StratifiedComputeResult{}, err
	}
	centers.Write(LOCALDIR, COORDFILE)
	//generate grid files?
	gridFileMap, err := sc.generateGridFiles()
	if err != nil {
		return StratifiedComputeResult{}, err
	}
	result := StratifiedComputeResult{
		CandiateLocations: centers,
		GridFiles:         gridFileMap,
	}
	return result, nil
}
func (sc StratifiedCompute) DetermineValidLocations(inputRoot cc.DataSource) (ValidLocationsComputeResult, error) {
	var computeResult ValidLocationsComputeResult
	allStormsAllLocations := make([]LocationInfo, 0)
	validLocationMap := make(map[string]utils.CoordinateList, 0)
	//generate of candidate storm centers.
	candidateStormCenters, err := sc.generateStormCenters()
	if err != nil {
		return computeResult, err
	}
	//take list of cell centers for the study area
	studyAreaCellCenters, err := generateUniformPointList(sc.StudyAreaPolygon, sc.Spacing)
	if err != nil {
		return computeResult, err
	}
	ref := gdal.CreateSpatialReference("")
	ref.FromEPSG(5070)
	outref := gdal.CreateSpatialReference("")
	outref.FromEPSG(4326)
	root := path.Dir(inputRoot.Paths["default"])
	//could be a go routine at this level
	//loop through the storms in the grid file(in order for simplicity)
	stormcenterbytes := make([]byte, 0)
	for _, storm := range sc.GridFile.Events {
		//create a validlocation coordinate list.
		validLocations := utils.CoordinateList{Coordinates: make([]utils.Coordinate, 0)}
		//determine the center of the storm.

		stormCenter, err := gdal.CreateFromWKT(fmt.Sprintf("Point (%v %v)\n", storm.CenterX, storm.CenterY), ref)
		if err != nil {
			return computeResult, err
		}
		err = stormCenter.TransformTo(outref)
		if err != nil {
			return computeResult, err
		}
		stormCoord := utils.Coordinate{X: stormCenter.Y(0), Y: stormCenter.X(0)}
		stormcenterbytes = append(stormcenterbytes, fmt.Sprintf("%v,%v,%v\n", storm.Name, stormCoord.X, stormCoord.Y)...)
		//determine the start date of the storm
		startDate := strings.Split(storm.Name, " ")[1]
		//create a vsis3 path to that tif
		tr, err := utils.InitTifReader(fmt.Sprintf("%v/%v.tif", root, startDate)) //get root path from one of the input data sources?
		if err != nil {
			return computeResult, err
		}
		defer tr.Close()

		//fmt.Println(time.Now())
		//loop through each point in the candidate storm centers
		for _, candidate := range candidateStormCenters.Coordinates {
			locationInfo := LocationInfo{
				StormName:  storm.Name,
				Coordinate: candidate,
				IsValid:    false,
			}
			//calculate an offset from the center to the new destination location
			offset := candidate.DetermineXandYOffset(stormCoord)
			//invert that offset
			offset.X = -offset.X
			offset.Y = -offset.Y
			//loop through each point in the cell centers for the study area
			hasPrecipitation := false
			hasNull := false
			for _, cellCenter := range studyAreaCellCenters.Coordinates {
				//offset the point by the inverted offset
				cellCenter.ShiftPoint(offset)
				//query the vsis3 tiff
				value, err := tr.Query(cellCenter)
				if err != nil {
					//null or out of tif range, reject
					hasNull = true
					break //if data is null reject location
				}
				if value > sc.AcceptanceDepthThreshold { //if data is greater than 0 in any cell accept location
					hasPrecipitation = true
				}
			}
			if hasPrecipitation && !hasNull {
				locationInfo.IsValid = true
				validLocations.Coordinates = append(validLocations.Coordinates, candidate)
			}
			allStormsAllLocations = append(allStormsAllLocations, locationInfo)
			//next cell center
		} //next transposition location
		validLocationMap[fmt.Sprintf("%v.csv", startDate)] = validLocations
	} //next storm
	computeResult.StormMap = validLocationMap
	computeResult.AllStormsAllLocations = allStormsAllLocations
	fmt.Println(string(stormcenterbytes))
	return computeResult, nil
}
func (sc StratifiedCompute) generateStormCenters() (utils.CoordinateList, error) {
	return generateUniformPointList(sc.TranspositionPolygon, sc.Spacing)

}
func generateUniformPointList(ds gdal.DataSource, spacing float64) (utils.CoordinateList, error) {
	coordinates := utils.CoordinateList{Coordinates: make([]utils.Coordinate, 0)}
	layer := ds.LayerByIndex(0)
	ref := layer.SpatialReference()
	//fmt.Println("features:")
	//fmt.Println(layer.FeatureCount(true))
	polygon := layer.NextFeature()
	defer polygon.Destroy()
	envelope, err := layer.Extent(true)
	if err != nil {
		return coordinates, err
	}
	MaxX := envelope.MaxX()
	MinX := envelope.MinX()
	MinY := envelope.MinY()
	MaxY := envelope.MaxY()
	y := 0
	x := 0
	//get distance in x domain
	xdist := MaxX - MinX

	//get distance in y domain
	ydist := MaxY - MinY
	//get total number of x and y steps
	xSteps := int(math.Floor(math.Abs(xdist) / spacing))
	ySteps := int(math.Floor(math.Abs(ydist) / spacing))
	//offset by half in each direction
	currentYval := MaxY + (spacing / 2)
	var currentXval float64
	//generate a full row, incriment y and start the next row.
	for y < ySteps { //iterate across all rows
		x = 0
		currentXval = MinX + (spacing / 2)
		for x < xSteps { // Iterate across all x values in a row
			x++
			currentXval += spacing
			//determine if polygon contains the point.
			location, err := gdal.CreateFromWKT(fmt.Sprintf("Point (%v %v)\n", currentXval, currentYval), ref)
			if err != nil {
				return coordinates, err
			}
			if polygon.Geometry().Contains(location) {
				//record the location.
				coordinates.Coordinates = append(coordinates.Coordinates, utils.Coordinate{X: currentXval, Y: currentYval})
			}
		}
		y++ //step to next row
		currentYval -= spacing
	}
	return coordinates, err
}
func wrieStormCenters(coordinates utils.CoordinateList) error {
	//write out coordinates.
	return coordinates.Write(LOCALDIR, COORDFILE)
}

func (sc StratifiedCompute) generateGridFiles() (map[string][]byte, error) {
	gf := sc.GridFile
	outputMap := make(map[string][]byte, 0)
	//trim root to remove
	for _, pe := range gf.Events {
		for _, te := range gf.Temps {
			if strings.Contains(pe.Name, te.Name) {
				b := gf.ToBytes(pe, te)
				outputMap[fmt.Sprintf("%vGridFile.grid", pe.Name)] = b
			}
		}
	}
	return outputMap, nil
}
