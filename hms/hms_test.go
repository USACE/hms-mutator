package hms

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/usace/wat-go-sdk/plugin"
)

func TestReadMetModel(t *testing.T) {
	path := "../exampledata/AORC.met"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	m, _ := ReadMet(ri)
	if !strings.Contains(m.lines[0], PrecipStartKeyword) {
		t.Fail()
	}
}
func TestReadMca(t *testing.T) {
	path := "../exampledata/Uncertainty_1.mca"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	m, _ := ReadMca(ri)
	if !m.HasRealizations {
		t.Fail()
	}
}
func TestReadManipulateWriteMetModel(t *testing.T) {
	path := "../exampledata/AORC.met"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	m, _ := ReadMet(ri)
	m.UpdateStormCenter("1", "2")
	m.UpdateStormName("updated name")
	m.UpdateTimeShift("45")
	bytes, _ := m.WriteBytes()
	fmt.Println(string(bytes))
}
func TestReadControl(t *testing.T) {
	path := "../exampledata/Dec_2013.control"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	c, _ := ReadControl(ri)
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
	path := "../exampledata/Dec_2013.control"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	c, _ := ReadControl(ri)
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
	path := "../exampledata/IC_Transpose_eric.grid"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	g, _ := ReadGrid(ri)
	if len(g.Events) != 2 {
		t.Fail()
	}
}
func TestSelectGrid(t *testing.T) {
	path := "../exampledata/WhiteRiver_Muncie.grid"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	g, _ := ReadGrid(ri)
	rnd := rand.New(rand.NewSource(1234))
	for i := 0; i < 100; i++ {
		e, _ := g.SelectEvent(rnd.Int63())
		fmt.Printf("%v: %v\n", i, e.Name)
	}
}

func TestReadSelectUpdateWriteGrid(t *testing.T) {
	path := "../exampledata/KanawhaHMS.grid"
	ri := plugin.ResourceInfo{
		Store: plugin.LOCAL,
		Root:  "workspaces/hms-mutator/",
		Path:  path,
	}
	g, _ := ReadGrid(ri)
	rnd := rand.New(rand.NewSource(1234))
	e, _ := g.SelectEvent(rnd.Int63())
	_ = e.UpdateDSSFile("data/storms.dss")
	s := string(g.toBytes(e))
	fmt.Println(s)
}
