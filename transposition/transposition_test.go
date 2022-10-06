package transposition

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/usace/hms-mutator/hms"
	"github.com/usace/wat-go-sdk/plugin"
)

func TestInitTransposition(t *testing.T) {
	path := "../exampledata/muncie_simple_transpostion_region.gpkg"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	tr, err := InitModel(ri, ri)
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
	wpath := "../exampledata/watershedBoundary.gpkg"
	wri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  wpath,
	}
	tr, err := InitModel(ri, wri)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	ge := hms.PrecipGridEvent{
		Name:      "test",
		StartTime: "test",
		CenterX:   1.0,
		CenterY:   2.0,
		Lines:     []string{},
	}
	fmt.Printf("id,name,x,y\n")
	rng := rand.New(rand.NewSource(1234))
	for i := 0; i < 1; i++ {
		x, y, _ := tr.Transpose(rng.Int63(), ge)
		fmt.Printf("%v,%v,%v,%v\n", i, ge.Name, x, y)
	}
}
func TestSimulationCompute(t *testing.T) {
	path := "../exampledata/muncie_simple_transpostion_region.gpkg"
	gpkgRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	wpath := "../exampledata/watershedBoundary.gpkg"
	wgpkgRI := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  wpath,
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
	sim, err := InitSimulation(gpkgRI, wgpkgRI, metRI, gridRI, controlRI)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		//compute simulation for given seed set
		m, ge, err := sim.Compute(ss)
		if err != nil {
			fmt.Println(err)
			t.Fail()
		}
		fmt.Printf("%v\n", ge.Name)
		bytes, err := m.WriteBytes()
		if err != nil {
			fmt.Println(err)
			t.Fail()
		}
		fmt.Println(string(bytes))
	}

}
