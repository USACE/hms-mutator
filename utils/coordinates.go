package utils // CoordinateList represents a slice of Coordinates, can be used for many purposes, is used to identify transposition locations spaced thorughout the transposition domain.
import (
	"fmt"
	"strconv"
	"strings"
)

type CoordinateList struct {
	Coordinates []Coordinate
}

// Coordinate represents an x and y location such as a possible transposition location.
type Coordinate struct {
	X float64
	Y float64
}

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
