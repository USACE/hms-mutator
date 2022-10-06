package transposition

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/HydrologicEngineeringCenter/go-statistics/statistics"
	"github.com/dewberry/gdal"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/wat-go-sdk/plugin"
)

type Model struct {
	//uniform x distribution
	//uniform y distribution
	yDist statistics.UniformDistribution
	xDist statistics.ContinuousDistribution
	//uniform start time distribution
	transpositionRegionDS gdal.DataSource
	watershedBoundaryDS   gdal.DataSource
}
type ModelResult struct {
	X float64
	Y float64
	//time offset?
}

func InitModel(transpositionRegion plugin.ResourceInfo, watershedBoundary plugin.ResourceInfo) (Model, error) {
	wbytes, err := plugin.DownloadObject(watershedBoundary)
	if err != nil {
		return Model{}, err
	}
	localDir := "/app/data/"
	wfileName := "watershedBoundary.gpkg"
	wfilePath := fmt.Sprintf("%v%v", localDir, wfileName)
	err = writeLocalBytes(wbytes, localDir, wfilePath)
	//h, err := gpkg.Open(filePath)
	//ext, err := h.CalculateGeometryExtent("muncie_simple_transposition_region")
	//fmt.Print(ext)
	wds := gdal.OpenDataSource(wfilePath, 0) //defer disposing the datasource and layers.
	//ensure path is local
	bytes, err := plugin.DownloadObject(transpositionRegion)
	if err != nil {
		return Model{}, err
	}
	fileName := "transpositionregion.gpkg"
	filePath := fmt.Sprintf("%v%v", localDir, fileName)
	err = writeLocalBytes(bytes, localDir, filePath)
	ds := gdal.OpenDataSource(filePath, 0) //defer disposing the datasource and layers.
	layer := ds.LayerByIndex(0)
	envelope, err := layer.Extent(true)
	MaxX := envelope.MaxX()
	MinX := envelope.MinX()
	MinY := envelope.MinY()
	MaxY := envelope.MaxY()
	x := statistics.UniformDistribution{Max: MaxX, Min: MinX}
	y := statistics.UniformDistribution{Max: MaxY, Min: MinY}
	return Model{
		yDist:                 y,
		xDist:                 x,
		transpositionRegionDS: ds,
		watershedBoundaryDS:   wds,
	}, nil
}
func (t Model) Transpose(seed int64, pge hms.PrecipGridEvent) (float64, float64, error) {
	r := rand.New(rand.NewSource(seed))
	layer := t.transpositionRegionDS.LayerByIndex(0)
	transpositionRegion := layer.Feature(1)
	if transpositionRegion.IsNull() {
		fmt.Println("im null...")
	}

	wlayer := t.watershedBoundaryDS.LayerByIndex(0)
	fmt.Printf("watershed Layers %v\n", t.watershedBoundaryDS.LayerCount())
	wf := wlayer.Feature(1)
	fcount, ok := wlayer.FeatureCount(true)
	if ok {
		fmt.Printf("watershed layer(1) features %v\n", fcount)
	}

	if wf.IsNull() {
		fmt.Println("im null...")
	}
	ref := layer.SpatialReference()
	xOffset := 0.0
	yOffset := 0.0

	for {
		xrand := rand.New(rand.NewSource(r.Int63()))
		yrand := rand.New(rand.NewSource(r.Int63()))
		xval := t.xDist.InvCDF(xrand.Float64())
		yval := t.yDist.InvCDF(yrand.Float64())

		//validate if in transposition polygon, iterate until it is
		newCenter, err := gdal.CreateFromWKT(fmt.Sprintf("Point (%v %v)\n", xval, yval), ref)
		if err != nil {
			return xval, yval, err
		}
		//fmt.Println(geom.Envelope())
		if transpositionRegion.Geometry().Contains(newCenter) {
			xOffset = xval - pge.CenterX
			yOffset = yval - pge.CenterY
			fmt.Printf("Offset(x,y): (%v,%v)\n", xOffset, yOffset)
			shiftContained := false                           //TODO switch to false and test.
			shiftedWatershedBoundary := wf.Geometry().Clone() //shift watershed boundary
			//shiftedWatershedBoundary = shiftedWatershedBoundary.Geometry(0)
			geometrycount := shiftedWatershedBoundary.GeometryCount()
			//count := shiftedWatershedBoundary.PointCount()
			//origCount := wf.Geometry().Geometry(0).PointCount()

			for g := 0; g < geometrycount; g++ {
				geometry := shiftedWatershedBoundary.Geometry(g)
				fmt.Printf("Geometry Type %v\n", geometry.Type())
				fmt.Printf("Geometry Area %v\n", geometry.Area())
				parts := geometry.GeometryCount()
				fmt.Printf("Geometry parts %v\n", parts)
				geometryisring := geometry.IsRing()
				fmt.Printf("geometry is ring %v\n", geometryisring)
				geometryPointCount := geometry.PointCount()
				fmt.Printf("geometry point count %v\n", geometryPointCount)
				geometryForced := geometry.ForceToPolygon()
				geometryPointCount = geometryForced.PointCount()
				fmt.Printf("forced geometry point count %v\n", geometryPointCount)
				for i := 0; i < geometryPointCount; i++ {
					px, py, pz := geometryForced.Point(i)
					shiftedWatershedBoundary.Geometry(g).SetPoint(i, px-xOffset, py-yOffset, pz) //does this work or does it insert?
				}
			}

			//check shifted watershed boundary is contained in transposition region
			shiftContained = transpositionRegion.Geometry().Contains(shiftedWatershedBoundary)
			if shiftContained {
				return xval, yval, nil
			} else {
				plugin.Log(plugin.Message{
					Status:    plugin.COMPUTING,
					Progress:  50,
					Level:     plugin.INFO,
					Message:   fmt.Sprintf("storm center (%v,%v) rejected due to possible null data\n", xval, yval),
					Sender:    "hms-mutator",
					PayloadId: "unknown payload id",
				})
			}
		}
	}
}
func writeLocalBytes(b []byte, destinationRoot string, destinationPath string) error {
	if _, err := os.Stat(destinationRoot); os.IsNotExist(err) {
		os.MkdirAll(destinationRoot, 0644) //do i need to trim filename?
	}
	err := os.WriteFile(destinationPath, b, 0644)
	if err != nil {
		plugin.Log(plugin.Message{
			Message: fmt.Sprintf("failure to write local file: %v\n\terror:%v", destinationPath, err),
			Level:   plugin.ERROR,
			Sender:  "transposition",
		})
		return err
	}
	return nil
}
