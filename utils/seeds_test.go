package utils

import (
	"fmt"
	"testing"

	"github.com/usace/cc-go-sdk"
	tiledb "github.com/usace/cc-go-sdk/tiledb-store"
)

func TestReadSeedsFromTDB(t *testing.T) {
	//register tiledb
	cc.DataStoreTypeRegistry.Register("TILEDB", tiledb.TileDbEventStore{})
	pm, err := cc.InitPluginManager()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	seeds, err := ReadSeedsFromTiledb(pm.IOManager, "store", "seeds", "hms-mutator")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(seeds)
}
