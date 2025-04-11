package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/usace/cc-go-sdk"
	"github.com/usace/filesapi"
)

func WriteLocalBytes(b []byte, destinationRoot string, destinationPath string) error {
	if _, err := os.Stat(destinationRoot); os.IsNotExist(err) {
		os.MkdirAll(destinationRoot, 0644) //do i need to trim filename?
	}
	return os.WriteFile(destinationPath, b, 0644)
}
func ListAllPaths(ioManager cc.IOManager, StoreKey string, DirectoryKey string, filter string) ([]string, error) {
	store, err := ioManager.GetStore(StoreKey)
	var pathList []string
	if err != nil {
		return pathList, err
	}
	session, ok := store.Session.(cc.S3DataStore)
	if !ok {
		return pathList, fmt.Errorf("%v was not an s3datastore type", StoreKey)
	}
	rawSession, ok := session.GetSession().(filesapi.FileStore)
	if !ok {
		return pathList, errors.New("could not convert s3datastore raw session into filestore type")
	}
	pageIdx := 0 //does page index start with 0 or 1?
	input := filesapi.ListDirInput{
		Path:   filesapi.PathConfig{Path: DirectoryKey},
		Page:   pageIdx,
		Size:   filesapi.DEFAULTMAXKEYS,
		Filter: filter,
	}
	for {
		fapiresult, err := rawSession.ListDir(input)
		if err != nil {
			//check if there are files in the list?
			return pathList, err
		}
		list := *fapiresult
		for _, s := range list {
			pathList = append(pathList, s.Path)
		}
		if len(list) < 1000 {
			return pathList, nil
		} else {
			pageIdx++
		}
	}
}
