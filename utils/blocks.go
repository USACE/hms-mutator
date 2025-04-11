package utils

import (
	"github.com/usace/cc-go-sdk"
)

type Block struct {
	RealizationIndex int32 `json:"realization_index" eventstore:"realization_index"`
	BlockIndex       int32 `json:"block_index" eventstore:"block_index"`
	BlockEventCount  int32 `json:"block_event_count" eventstore:"block_event_count"`
	BlockEventStart  int64 `json:"block_event_start" eventstore:"block_event_start"` //inclusive - will be one greater than previous event end
	BlockEventEnd    int64 `json:"block_event_end" eventstore:"block_event_end"`     //inclusive - will be one less than event start if event count is 0.
}

// blocks are a recordset under a usernamed dataset in the seed generator manifest named "outputDataset_name" - blocks is the standard name.
func ReadBlocksFromTiledb(pm *cc.PluginManager, tileDbStoreName string, datasetName string) ([]Block, error) {
	blocks := make([]Block, 0)
	//get the recordset
	recordset, err := cc.NewEventStoreRecordset(pm, &blocks, tileDbStoreName, datasetName)
	if err != nil {
		return blocks, err
	}
	result, err := recordset.Read()
	if err != nil {
		return blocks, err
	}
	for i := 0; i < result.Size(); i++ {
		block := Block{}
		err = result.Scan(&block)
		if err != nil {
			return blocks, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}
