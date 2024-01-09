package utils

import (
	"errors"

	"github.com/dewberry/gdal"
)

type TifReader struct {
	FilePath         string
	ds               *gdal.Dataset
	nodata           float64
	verticalIsMeters bool //default false
}

// init creates and produces an unexported cogReader
func InitTifReader(fp string) (TifReader, error) {
	//read the file path
	//make sure it is a tif
	//fmt.Println("Connecting to: " + fp)
	ds, err := gdal.Open(fp, gdal.ReadOnly)
	if err != nil {
		return TifReader{}, errors.New("Cannot connect to tif at path " + fp + err.Error())
	}
	v, valid := ds.RasterBand(1).NoDataValue()
	cr := TifReader{FilePath: fp, ds: &ds, verticalIsMeters: false}
	if valid {
		cr.nodata = v
	}
	return cr, nil
}
func (cr *TifReader) Close() {
	cr.ds.Close()
}
func (cr *TifReader) Query(c Coordinate) (float64, error) {
	rb := cr.ds.RasterBand(1)

	igt := cr.ds.InvGeoTransform()
	px := int(igt[0] + c.X*igt[1] + c.Y*igt[2])
	py := int(igt[3] + c.X*igt[4] + c.Y*igt[5])
	buffer := make([]float32, 1*1)
	if px < 0 || px > rb.XSize() {
		return cr.nodata, errors.New("X is out of range")
	}
	if py < 0 || py > rb.YSize() {
		return cr.nodata, errors.New("Y is out of range")
	}
	err := rb.IO(gdal.Read, px, py, 1, 1, buffer, 1, 1, 0, 0)
	if err != nil {
		return cr.nodata, err
	}
	depth := buffer[0]
	d := float64(depth)
	if d == cr.nodata {
		return cr.nodata, errors.New("COG reader had the no data value observed, setting to %v")
	}
	return d, nil
}
