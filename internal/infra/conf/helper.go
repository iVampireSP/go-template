package conf

import (
	"go-template/configs"
	"os"
)

func createConfigIfNotExists(path string) {
	if path != "" {
		return
	}

	// create if not exists
	var configName = "config.yaml"

	if _, err := os.Stat(configName); os.IsNotExist(err) {
		f, err := os.Create(configName)
		if err != nil {
			panic(err)
		}

		// write default from embed
		_, err = f.Write(configs.Config)
		if err != nil {
			panic(err)
		}

		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}(f)
	}
}
