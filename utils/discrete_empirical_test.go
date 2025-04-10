package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestRead_StormTypeDists(t *testing.T) {
	//register tiledb
	bytes, err := os.ReadFile("/workspaces/hms-mutator/exampledata/st1.csv")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	dists := DescreteEmpiricalDistributionFromBytes(bytes)
	fmt.Println(dists.Sample(-1))
	fmt.Println(dists.Sample(0))
	fmt.Println(dists.Sample(.25))
	fmt.Println(dists.Sample(.5))
	fmt.Println(dists.Sample(.75))
	fmt.Println(dists.Sample(1))
	fmt.Println(dists.Sample(1.1))
	fmt.Println(dists)
}
