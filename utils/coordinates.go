package utils // CoordinateList represents a slice of Coordinates, can be used for many purposes, is used to identify transposition locations spaced thorughout the transposition domain.
import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/usace/cc-go-sdk"
)

type CoordinateList struct {
	Coordinates []Coordinate
}

// Coordinate represents an x and y location such as a possible transposition location.
type Coordinate struct {
	X float64
	Y float64
}
type FishNetMap map[string]CoordinateList //storm type coordinate list.
func (o Coordinate) DetermineXandYOffset(c Coordinate) Coordinate {
	xdifference := o.X - c.X
	ydifference := o.Y - c.Y
	return Coordinate{X: xdifference, Y: ydifference}
}
func (o *Coordinate) ShiftPoint(offset Coordinate) {
	o.X += offset.X
	o.Y += offset.Y
}
func (c Coordinate) ToString() string {
	return fmt.Sprintf("%v,%v\r\n", c.X, c.Y)
}
func (cl CoordinateList) Write(root string, path string) error {
	b := cl.ToBytes()
	err := WriteLocalBytes(b, root, fmt.Sprintf("%v%v", root, path))
	return err
}
func (cl CoordinateList) ToBytes() []byte {
	b := make([]byte, 0)
	b = append(b, "x,y\r\n"...)
	for _, c := range cl.Coordinates {
		b = append(b, c.ToString()...)
	}
	return b
}
func BytesToCoordinateList(bytes []byte) (CoordinateList, error) {
	coords := make([]Coordinate, 0)
	list := CoordinateList{Coordinates: coords}
	bytestring := string(bytes)
	stringlist := strings.Split(bytestring, "\r\n")

	for i, c := range stringlist {
		if i != 0 { //skip header
			if len(c) > 0 {
				points := strings.Split(c, ",")
				x, err := strconv.ParseFloat(points[0], 64)
				if err != nil {
					return list, err
				}
				y, err := strconv.ParseFloat(points[1], 64)
				if err != nil {
					return list, err
				}
				coord := Coordinate{
					X: x,
					Y: y,
				}
				list.Coordinates = append(list.Coordinates, coord)
			}

		}

	}
	return list, nil
}
func ReadFishNets(iomanager cc.IOManager, storeKey string, filePaths []string) (FishNetMap, error) {
	FishNetMap := make(map[string]CoordinateList)
	store, err := iomanager.GetStore(storeKey)
	if err != nil {
		return FishNetMap, err
	}
	session, ok := store.Session.(cc.S3DataStore)
	if !ok {
		return FishNetMap, errors.New(fmt.Sprintf("%v was not an s3datastore type", storeKey))
	}
	root := store.Parameters.GetStringOrFail("root")
	for _, path := range filePaths {
		pathpart := strings.Replace(path, fmt.Sprintf("%v/", root), "", -1)
		reader, err := session.Get(pathpart, "")
		if err != nil {
			return FishNetMap, err
		}
		bytes, err := io.ReadAll(reader)
		if err != nil {
			return FishNetMap, err
		}
		coordlist, err := BytesToCoordinateList(bytes)
		if err != nil {
			return FishNetMap, err
		}
		FishNetMap[path] = coordlist
	}
	return FishNetMap, nil
}
