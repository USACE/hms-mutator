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
	xDist statistics.UniformDistribution
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
	x := statistics.UniformDistribution{Max: envelope.MaxX(), Min: envelope.MinX()}
	y := statistics.UniformDistribution{Max: envelope.MaxY(), Min: envelope.MinY()}
	return Model{
		yDist: y,
		xDist: x,
		ds:    ds,
	}, nil
}
func (t Model) Transpose(seed int64) (float64, float64, error) {
	r := rand.New(rand.NewSource(seed))
	layer := t.ds.LayerByIndex(0)
	f := layer.Feature(0)
	jsonStartString := "{\"type\": \"Feature\",\"geometry\": {\"type\": \"Point\",\"coordinates\": ["
	jsonEndString := "]},\"properties\": {\"name\": \"Storm Center\"}}"
	for {
		xrand := rand.New(rand.NewSource(r.Int63()))
		yrand := rand.New(rand.NewSource(r.Int63()))
		xval := t.xDist.InvCDF(xrand.Float64())
		yval := t.xDist.InvCDF(yrand.Float64())
		//validate if in transposition polygon, iterate until it is
		jsonString := fmt.Sprintf("%v%v,%v%v", jsonStartString, xval, yval, jsonEndString)
		geom := gdal.CreateFromJson(jsonString)

		if f.Geometry().Intersects(geom) {
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
