package transposition

import (
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/HydrologicEngineeringCenter/go-statistics/statistics"
	"github.com/dewberry/gdal"

	"github.com/usace/hms-mutator/hms"
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

func InitModel(transpositionRegion []byte, watershedBoundary []byte) (Model, error) {
	model := Model{}
	localDir := "/app/data/"
	wfileName := "watershedBoundary.gpkg"
	wfilePath := fmt.Sprintf("%v%v", localDir, wfileName)
	err := writeLocalBytes(watershedBoundary, localDir, wfilePath)
	if err != nil {
		return model, err
	}
	//h, err := gpkg.Open(filePath)
	//ext, err := h.CalculateGeometryExtent("muncie_simple_transposition_region")
	//fmt.Print(ext)
	wds := gdal.OpenDataSource(wfilePath, 0) //defer disposing the datasource and layers.
	//ensure path is local
	fileName := "transpositionregion.gpkg"
	filePath := fmt.Sprintf("%v%v", localDir, fileName)
	err = writeLocalBytes(transpositionRegion, localDir, filePath)
	if err != nil {
		return model, err
	}
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
	//fmt.Printf("transposition Layers %v\n", t.watershedBoundaryDS.LayerCount())
	//tfc, _ := layer.FeatureCount(true)
	//fmt.Printf("watershed features in layer 0 %v\n", tfc)
	//defer layer.Definition().Destroy()
	transpositionRegion := layer.NextFeature()
	defer transpositionRegion.Destroy()
	if transpositionRegion.IsNull() {
		fmt.Println("im null...")
	} else {
		fmt.Println("transposition region is not null")

	}

	wlayer := t.watershedBoundaryDS.LayerByIndex(0)
	//fmt.Printf("watershed Layers %v\n", t.watershedBoundaryDS.LayerCount())
	//fc, _ := wlayer.FeatureCount(true)
	//fmt.Printf("watershed features in layer 0 %v\n", fc)
	wf := wlayer.NextFeature()
	defer wf.Destroy()
	if wf.Geometry().Type() != 3 {
		return 0, 0, errors.New("watershed boundary geometry not a simple polygon")
	}
	if wf.IsNull() {
		fmt.Println("im null...")
	}
	ref := layer.SpatialReference()
	//defer ref.Destroy()
	xOffset := 0.0
	yOffset := 0.0
	//fmt.Printf("Original Center (%v,%v)\n", pge.CenterX, pge.CenterY)
	for {
		xrand := rand.New(rand.NewSource(r.Int63()))
		yrand := rand.New(rand.NewSource(r.Int63()))
		xval := t.xDist.InvCDF(xrand.Float64())
		yval := t.yDist.InvCDF(yrand.Float64())

		//validate if in transposition polygon, iterate until it is
		newCenter, err := gdal.CreateFromWKT(fmt.Sprintf("Point (%v %v)\n", xval, yval), ref)
		//defer newCenter.Destroy()
		if err != nil {
			return xval, yval, err
		}

		//fmt.Println(geom.Envelope())
		if transpositionRegion.Geometry().Contains(newCenter) {
			xOffset = xval - pge.CenterX
			yOffset = yval - pge.CenterY

			//fmt.Printf("Offset(x,y): (%v,%v)\n", xOffset, yOffset)
			shiftContained := false                           //TODO switch to false and test.
			shiftedWatershedBoundary := wf.Geometry().Clone() //shift watershed boundary
			//defer shiftedWatershedBoundary.Destroy()
			//shiftedWatershedBoundary = shiftedWatershedBoundary.Geometry(0)
			geometrycount := shiftedWatershedBoundary.GeometryCount()
			//count := shiftedWatershedBoundary.PointCount()
			//origCount := wf.Geometry().Geometry(0).PointCount()

			for g := 0; g < geometrycount; g++ {
				geometry := shiftedWatershedBoundary.Geometry(g)
				defer geometry.Destroy()
				geometryPointCount := geometry.PointCount()
				//fmt.Printf("geometry point count %v\n", geometryPointCount)
				for i := 0; i < geometryPointCount; i++ {
					px, py, pz := geometry.Point(i)
					shiftedWatershedBoundary.Geometry(g).SetPoint(i, px-xOffset, py-yOffset, pz) //does this work or does it insert?
				}
			}

			//check shifted watershed boundary is contained in transposition region
			shiftContained = transpositionRegion.Geometry().Contains(shiftedWatershedBoundary)
			if shiftContained {
				//s, _ := shiftedWatershedBoundary.ToWKT()
				//fmt.Println(s)
				//return pge.CenterX, pge.CenterY, nil //for debugging issues with time offsets and to avoid confusion created by different centerings.
				return xval, yval, nil
			}
		}
	}
}
func writeLocalBytes(b []byte, destinationRoot string, destinationPath string) error {
	if _, err := os.Stat(destinationRoot); os.IsNotExist(err) {
		os.MkdirAll(destinationRoot, 0644) //do i need to trim filename?
	}
	return os.WriteFile(destinationPath, b, 0644)
}
