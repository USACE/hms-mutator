package transposition

type Model struct {
	//uniform x distribution
	//uniform y distribution
	//uniform start time distribution
	//shapefilePath
}
type ModelResult struct {
	X float64
	Y float64
	//time offset?
}

func InitModel(transpositionRegion string) (Model, error) {
	return Model{}, nil
}
