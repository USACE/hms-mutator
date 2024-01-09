package utils // CoordinateList represents a slice of Coordinates, can be used for many purposes, is used to identify transposition locations spaced thorughout the transposition domain.
import "fmt"

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
