package conf

import (
	"go-template/internal/base/logger"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// depth 是配置文件的搜索深度
var depth = 8

func getConfigPath() string {
	var path string
	if os.Getenv("AMBER_CONFIG") != "" {
		path = os.Getenv("AMBER_CONFIG")
		return path
	}
	var pathOptions []string
	for i := 0; i <= depth; i++ {
		pathOptions = append(pathOptions, strings.Repeat("../", i)+"config.yaml")
	}
	for _, p := range pathOptions {
		if _, err := os.Stat(p); err == nil {
			path = p
			break
		}
	}

	if path != "" {
		// 假设 workDir 是当前工作目录的路径
		workDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		// 将相对路径转换为绝对路径
		path, err = filepath.Abs(filepath.Join(workDir, path))
		if err != nil {
			panic(err)
		}
	}

	return path
}

func ProviderConfig(logger *logger.Logger) *Config {
	var path = getConfigPath()
	createConfigIfNotExists(path)

	if path == "" {
		logger.Sugar.Fatal("config file not found, created.")
	} else {
		logger.Sugar.Infof("config file found at %s", path)
	}

	c := &Config{}
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		logger.Sugar.Fatal(err)
	}
	if err := v.Unmarshal(c); err != nil {
		panic(err)
	}

	return c
}
