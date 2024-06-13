package transposition

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/usace/hms-mutator/hms"
)

func TestInitTransposition(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/muncie_simple_transpostion_region.gpkg"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	tr, err := InitModel(bytes, bytes)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Printf("X Dist: %v \nY Dist: %v", tr.xDist, tr.yDist)
}
func TestSampleLocations(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/muncie_simple_transpostion_region.gpkg"
	tbytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	wpath := "/workspaces/hms-mutator/exampledata/watershedBoundary_2.gpkg"
	wbytes, err := ioutil.ReadFile(wpath)
	if err != nil {
		t.Fail()
	}
	tr, err := InitModel(tbytes, wbytes)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	gpath := "/workspaces/hms-mutator/exampledata/IC_Transpose-v2.grid"
	gbytes, err := ioutil.ReadFile(gpath)
	if err != nil {
		t.Fail()
	}
	gf, _ := hms.ReadGrid(gbytes)
	fmt.Printf("id,name,x,y\n")
	rng := rand.New(rand.NewSource(1234))
	for i := 0; i < 1; i++ {
		ge, _, _ := gf.SelectEvent(rng.Int63())
		x, y, _ := tr.Transpose(rng.Int63(), ge)
		fmt.Printf("%v,%v,%v,%v\n", i, ge.Name, x, y)
	}
}
func TestSimulationCompute(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/TranspositionDomain.gpkg"
	tbytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	wpath := "/workspaces/hms-mutator/exampledata/WatershedBoundary.gpkg"
	wbytes, err := ioutil.ReadFile(wpath)
	if err != nil {
		t.Fail()
	}
	mpath := "/workspaces/hms-mutator/exampledata/POR.met"
	mbytes, err := ioutil.ReadFile(mpath)
	if err != nil {
		t.Fail()
	}
	/*
		cpath := "/workspaces/hms-mutator/exampledata/Dec_2013.control"
		cbytes, err := ioutil.ReadFile(cpath)
		if err != nil {
			t.Fail()
		}
	*/
	gpath := "/workspaces/hms-mutator/exampledata/HMS.grid"
	gbytes, err := ioutil.ReadFile(gpath)
	if err != nil {
		t.Fail()
	}

	gridFile, err := hms.ReadGrid(gbytes)
	metFile, err := hms.ReadMet(mbytes)
	//controlFile, err := hms.ReadControl(cbytes)

	//initialize simulation
	sim, err := InitTranspositionSimulation(tbytes, wbytes, metFile, gridFile)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	} else {
		//compute simulation for given seed set
		m, ge, _, err := sim.Compute(1234, 4321, true, len(gridFile.Events))
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

/*
func Test_EventConfiguration(t *testing.T) {
	epath := "/workspaces/hms-mutator/exampledata/eventconfiguration2.json"
	seedset := make(map[string]plugin.SeedSet, 1)
	seeds := plugin.SeedSet{
		EventSeed:       5920220759044230130,
		RealizationSeed: 3501447260771739518,
	}
	seedset["hms-mutator-muncie-alt"] = seeds
	ec := plugin.EventConfiguration{
		RealizationNumber: 1,
		Seeds:             seedset,
	}
	bytes, err := json.Marshal(ec)
	os.Remove(epath)
	ioutil.WriteFile(epath, bytes, fs.ModeAppend)
	if err != nil {
		t.Fail()
	}
}*/
