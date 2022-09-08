package transposition

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/usace/wat-go-sdk/plugin"
)

func TestInitTransposition(t *testing.T) {
	path := "../exampledata/muncie_simple_transpostion_region.gpkg"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	tr, err := InitModel(ri)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("X Dist: %v \nY Dist: %v", tr.xDist, tr.yDist)
}
func TestSampleLocations(t *testing.T) {
	path := "../exampledata/muncie_simple_transpostion_region.gpkg"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	tr, err := InitModel(ri)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("id,x,y\n")
	rng := rand.New(rand.NewSource(1234))
	for i := 0; i < 100; i++ {
		x, y, _ := tr.Transpose(rng.Int63())
		fmt.Printf("%v,%v,%v\n", i, x, y)
	}
}
