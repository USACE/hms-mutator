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
func TestSimulationCompute(t *testing.T) {
	path := "../exampledata/muncie_simple_transpostion_region.gpkg"
	gpkgRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	mpath := "../exampledata/AORC.met"
	metRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  mpath,
	}
	cpath := "../exampledata/Dec_2013.control"
	controlRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  cpath,
	}
	gpath := "../exampledata/WhiteRiver_Muncie.grid"
	gridRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  gpath,
	}
	epath := "../exampledata/eventconfiguration.json"
	eventRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  epath,
	}
	ec, err := plugin.LoadEventConfiguration(eventRI)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	//obtain seed set
	ss, err := ec.SeedSet("hms-mutator-muncie-alt")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	//initialize simulation
	sim, err := InitSimulation(gpkgRI, metRI, gridRI, controlRI)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	//compute simulation for given seed set
	m, err := sim.Compute(ss)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	bytes, err := m.WriteBytes()
	fmt.Println(string(bytes))
}