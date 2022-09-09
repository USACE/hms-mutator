package transposition

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/HydrologicEngineeringCenter/go-statistics/statistics"
	"github.com/dewberry/gdal"
	"github.com/usace/wat-go-sdk/plugin"
)

type Model struct {
	//uniform x distribution
	//uniform y distribution
	yDist statistics.UniformDistribution
	xDist statistics.ContinuousDistribution
	//uniform start time distribution
	ds gdal.DataSource
}
type ModelResult struct {
	X float64
	Y float64
	//time offset?
}

func InitModel(transpositionRegion plugin.ResourceInfo) (Model, error) {
	//ensure path is local
	bytes, err := plugin.DownloadObject(transpositionRegion)
	if err != nil {
		return Model{}, err
	}
	localDir := "/app/data/"
	fileName := "transpositionregion.gpkg"
	filePath := fmt.Sprintf("%v%v", localDir, fileName)
	err = writeLocalBytes(bytes, localDir, filePath)
	ds := gdal.OpenDataSource(filePath, 0) //defer disposing the datasource and layers.
	layer := ds.LayerByIndex(0)
	envelope, err := layer.Extent(true)
	MaxX := envelope.MaxX()
	MinX := envelope.MinX()
	//MidX := (MinX + MaxX) / 2.0
	MinY := envelope.MinY()
	MaxY := envelope.MaxY()
	//fmt.Println(envelope)
	x := statistics.UniformDistribution{Max: MaxX, Min: MinX}
	y := statistics.UniformDistribution{Max: MaxY, Min: MinY}
	return Model{
		yDist: y,
		xDist: x,
		ds:    ds,
	}, nil
}
func (t Model) Transpose(seed int64) (float64, float64, error) {
	r := rand.New(rand.NewSource(seed))
	layer := t.ds.LayerByIndex(0)
	f := layer.Feature(1)
	if f.IsNull() {
		fmt.Println("im null...")
	}
	//fmt.Println(f.Geometry().Envelope())
	ref := layer.SpatialReference()

	for {
		xrand := rand.New(rand.NewSource(r.Int63()))
		yrand := rand.New(rand.NewSource(r.Int63()))
		xval := t.xDist.InvCDF(xrand.Float64())
		yval := t.yDist.InvCDF(yrand.Float64())
		//validate if in transposition polygon, iterate until it is
		geom, err := gdal.CreateFromWKT(fmt.Sprintf("Point (%v %v)", xval, yval), ref)
		//fmt.Println(geom.ToWKT())
		//fmt.Printf("%v,%v,%v\n", 1, xval, yval)
		if err != nil {
			return xval, yval, err
		}
		//fmt.Println(geom.Envelope())
		if f.Geometry().Contains(geom) {
			return xval, yval, nil
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
