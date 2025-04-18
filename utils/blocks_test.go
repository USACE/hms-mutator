package utils

import (
	"fmt"
	"testing"

	"github.com/usace/cc-go-sdk"
	tiledb "github.com/usace/cc-go-sdk/tiledb-store"
)

func TestReadBlocksFromTDB(t *testing.T) {
	//register tiledb
	cc.DataStoreTypeRegistry.Register("TILEDB", tiledb.TileDbEventStore{})
	pm, err := cc.InitPluginManager()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	blocks, err := ReadBlocksFromTiledb(pm, "store", "blocks")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(blocks)
}
