package hms

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestReadMetModel(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/SST.met"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	m, err := ReadMet(bytes)
	if !strings.Contains(m.lines[0], PrecipStartKeyword) {
		t.Fail()
	}
	if err != nil {
		t.Fail()
	}
}
func TestReadMca(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/Uncertainty_1.mca"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	m, _ := ReadMca(bytes)
	if !m.HasRealizations {
		t.Fail()
	}
}
func TestReadManipulateWriteMetModel(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/AORC.met"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	m, _ := ReadMet(bytes)
	m.UpdateStormCenter("1", "2")
	m.UpdateStormName("updated name")
	//m.UpdateTimeShift("45")
	outbytes, _ := m.WriteBytes()
	fmt.Println(string(outbytes))
}
func TestReadControl(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/Dec_2013.control"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	c, _ := ReadControl(bytes)
	if c.Name != "Dec_2013" {
		t.Fail()
	}
	if c.StartDate != "19 December 2013" {
		t.Fail()
	}
	if c.StartTime != "23:59" { //this is a work around
		t.Fail()
	}
}
func TestControlOffset(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/Dec_2013.control"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	c, _ := ReadControl(bytes)
	//control time is 19 December 2013 23:59
	mins := c.ComputeOffset("20DEC2013:2359") //grid time needs to be offset backward.
	if mins < 0 {
		t.Fail()
	}
	mins = c.ComputeOffset("18DEC2013:2359") //gridtime needs to be offset forward
	if mins > 0 {
		t.Fail()
	}
}
func TestReadGrid(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/IC_Transpose-v2.grid"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	g, _ := ReadGrid(bytes)
	if len(g.Events) != 24 {
		t.Fail()
	}
}
func TestSelectGrid(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/WhiteRiver_Muncie.grid"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	g, _ := ReadGrid(bytes)
	rnd := rand.New(rand.NewSource(1234))
	for i := 0; i < 100; i++ {
		e, _, _ := g.SelectEvent(rnd.Int63())
		fmt.Printf("%v: %v\n", i, e.Name)
	}
}

func TestReadSelectUpdateWriteGrid(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/IC_Transpose-v2.grid"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	g, _ := ReadGrid(bytes)
	rnd := rand.New(rand.NewSource(1234))
	e, temp, _ := g.SelectEvent(rnd.Int63())
	_ = e.UpdateDSSFile("Storm")
	s := string(g.ToBytes(e, temp))
	fmt.Println(s)
}

func TestReadBootstrapSelectUpdateWriteGrid(t *testing.T) {
	path := "/workspaces/hms-mutator/exampledata/IC_Transpose-v2.grid"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fail()
	}
	g, _ := ReadGrid(bytes)
	rnd := rand.New(rand.NewSource(1234))
	g.Bootstrap(rnd.Int63(), len(g.Events))
	for _, pge := range g.Events {
		fmt.Printf("Event Name: %v \n", pge.Name)
	}
	e, temp, _ := g.SelectEvent(rnd.Int63())
	_ = e.UpdateDSSFile("Storm")
	s := string(g.ToBytes(e, temp))
	fmt.Println(s)
}
func TestTimeSubstitution(t *testing.T) {
	control := time.Date(2017, 9, 17, 1, 0, 0, 0, time.UTC)
	fmt.Println(control)
	grid := time.Date(1981, 2, 16, 0, 0, 0, 0, time.UTC)
	fmt.Println(grid)
	duration := -control.Sub(grid).Minutes()
	fmt.Println(duration)
	rounded := math.Round(duration)
	fmt.Println(rounded)
}
