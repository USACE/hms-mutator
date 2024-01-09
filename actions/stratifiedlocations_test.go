package actions

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/usace/cc-go-sdk"
	"github.com/usace/hms-mutator/hms"
)

func TestStratifiedLocations(t *testing.T) {
	parameters := make(map[string]any)
	parameters["spacing"] = 0.00902 // 1 km in the duwamish area (42,-124) is .00902 decimal degrees https://forest.moscowfsl.wsu.edu/fswepp/rc/kmlatcon.html
	action := cc.Action{
		Name:        "stratified_locations",
		Type:        "compute",
		Description: "create stratified locations for storms",
		Parameters:  parameters,
	}
	outputDataSource := cc.DataSource{
		Name:      "outputdestination",
		ID:        &uuid.NameSpaceDNS,
		Paths:     []string{"/app/data"},
		DataPaths: []string{},
		StoreName: "Local",
	}
	gfbytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/D_Transpose.grid")
	if err != nil {
		t.Fail()
	}
	gf, err := hms.ReadGrid(gfbytes)
	if err != nil {
		t.Fail()
	}
	pbytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/TranspositionDomain.gpkg")
	if err != nil {
		t.Fail()
	}
	wbytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/WatershedBoundary.gpkg")
	if err != nil {
		t.Fail()
	}
	sc, err := InitStratifiedCompute(action, gf, pbytes, wbytes, outputDataSource)
	if err != nil {
		t.Fail()
	}
	_, err = sc.Compute()
	if err != nil {
		t.Fail()
	}
}
func TestValidStratifiedLocations(t *testing.T) {
	parameters := make(map[string]any)
	parameters["spacing"] = 0.00902 * 4 // 1 km in the duwamish area (42,-124) is .00902 decimal degrees https://forest.moscowfsl.wsu.edu/fswepp/rc/kmlatcon.html
	parameters["acceptance_threshold"] = 0.001
	action := cc.Action{
		Name:        "stratified_locations",
		Type:        "compute",
		Description: "create stratified locations for storms",
		Parameters:  parameters,
	}
	outputDataSource := cc.DataSource{
		Name:      "outputdestination",
		ID:        &uuid.NameSpaceDNS,
		Paths:     []string{"/app/data"},
		DataPaths: []string{},
		StoreName: "Local",
	}
	gfbytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/D_Transpose.grid")
	if err != nil {
		t.Fail()
	}
	gf, err := hms.ReadGrid(gfbytes)
	if err != nil {
		t.Fail()
	}
	tbytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/TranspositionDomain.gpkg")
	if err != nil {
		t.Fail()
	}
	wbytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/WatershedBoundary.gpkg")
	if err != nil {
		t.Fail()
	}
	sc, err := InitStratifiedCompute(action, gf, tbytes, wbytes, outputDataSource)
	if err != nil {
		t.Fail()
	}
	inputRoot := cc.DataSource{
		Name:      "inputRoot",
		ID:        &uuid.NameSpaceDNS,
		Paths:     []string{"/workspaces/hms-mutator/exampledata"},
		DataPaths: []string{},
		StoreName: "",
	}
	pm, err := cc.InitPluginManager()
	if err != nil {
		t.Fail()
	}
	vmap, err := sc.DetermineValidLocations(inputRoot, pm)
	if err != nil {
		t.Fail()
	}
	for k, v := range vmap {
		fmt.Print(fmt.Sprintf("%v: %v\n", k, len(v.Coordinates)))
	}
}
