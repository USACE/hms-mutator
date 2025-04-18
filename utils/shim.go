package utils

import (
	"bytes"
	"errors"
	"io"

	"github.com/usace/cc-go-sdk"
)

func GetFile(pm cc.PluginManager, datasource cc.DataSource, index string) ([]byte, error) {
	data := make([]byte, 0)
	store, err := pm.GetStore(datasource.StoreName)
	if err != nil {
		return data, err
	}
	s3DataStore, ok := store.Session.(*cc.S3DataStore)
	if !ok {
		return data, err
	}
	reader, err := s3DataStore.Get(datasource.Paths[index], "")
	if err != nil {
		return data, err
	}
	data, err = io.ReadAll(reader)
	return data, err
}
func PutFile(data []byte, pm cc.IOManager, datasource cc.DataSource, index string) error {
	store, err := pm.GetStore(datasource.StoreName)
	if err != nil {
		return err
	}
	s3DataStore, ok := store.Session.(*cc.S3DataStore)
	if !ok {
		return errors.New("session not the right type")
	}
	writer := bytes.NewReader(data)
	_, err = s3DataStore.Put(writer, datasource.Paths[index], "")
	if err != nil {
		return err
	}
	return nil
}
