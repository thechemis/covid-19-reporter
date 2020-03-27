package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const cacheFileName = "cache.json"

func loadCacheFromFile() (cache map[int]string, err error) {

	if _, err := os.Stat(cacheFileName); os.IsNotExist(err) {
		return nil, nil
	}

	var data []byte

	if data, err = ioutil.ReadFile(cacheFileName); err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return cache, nil
}

func saveCacheToFile(cache map[int]string) (err error) {

	var data []byte

	if data, err = json.Marshal(cache); err != nil {
		return err
	}

	return ioutil.WriteFile(cacheFileName, data, 0644)
}
