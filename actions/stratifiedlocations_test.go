package actions

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/usace/cc-go-sdk"
	"github.com/usace/hms-mutator/hms"
	"github.com/usace/hms-mutator/utils"
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
		Paths:     []string{"/workspaces/hms-mutator/exampledata/1979-02-05.tif"},
		DataPaths: []string{},
		StoreName: "",
	}
	//pm, err := cc.InitPluginManager()
	if err != nil {
		t.Fail()
	}
	output, err := sc.DetermineValidLocations(inputRoot)
	if err != nil {
		t.Fail()
	}
	root := "/workspaces/hms-mutator/exampledata/results"
	for k, v := range output.StormMap {
		fp := fmt.Sprintf("%v/%v.csv", root, k)
		utils.WriteLocalBytes(v.ToBytes(), root, fp)
		fmt.Print(fmt.Sprintf("%v: %v\n", k, len(v.Coordinates)))
	}
	fp2 := fmt.Sprintf("%v/%v.csv", root, "AllStormsAllLocations")
	fp3 := fmt.Sprintf("%v/%v.csv", root, "AllStormsValidLocations")
	outbytes := make([]byte, 0)
	outbytes = append(outbytes, "StormName,X,Y,IsValid\n"...)
	validoutbytes := make([]byte, 0)
	validoutbytes = append(validoutbytes, "StormName,X,Y,IsValid\n"...)
	//create random list of ints
	indexes := make([]int, len(output.AllStormsAllLocations))
	rand := rand.New(rand.NewSource(945631))
	for i := 0; i < len(indexes); i++ {
		j := rand.Intn(i + 1)
		if i != j {
			indexes[i] = indexes[j]
		}
		indexes[j] = i
	}
	for i, _ := range output.AllStormsAllLocations {
		outbytes = append(outbytes, fmt.Sprintf("%v,%v,%v,%v\n", output.AllStormsAllLocations[indexes[i]].StormName, output.AllStormsAllLocations[indexes[i]].Coordinate.X, output.AllStormsAllLocations[indexes[i]].Coordinate.Y, output.AllStormsAllLocations[indexes[i]].IsValid)...)
		if output.AllStormsAllLocations[indexes[i]].IsValid {
			validoutbytes = append(validoutbytes, fmt.Sprintf("%v,%v,%v,%v\n", output.AllStormsAllLocations[indexes[i]].StormName, output.AllStormsAllLocations[indexes[i]].Coordinate.X, output.AllStormsAllLocations[indexes[i]].Coordinate.Y, output.AllStormsAllLocations[indexes[i]].IsValid)...)
		}
	}
	utils.WriteLocalBytes(outbytes, root, fp2)
	utils.WriteLocalBytes(validoutbytes, root, fp3)
}
