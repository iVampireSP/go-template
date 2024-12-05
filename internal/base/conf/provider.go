package conf

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// depth 是配置文件的搜索深度
var depth = 8

func getConfigPath() string {
	var path string
	if os.Getenv("CONFIG") != "" {
		path = os.Getenv("CONFIG")
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

func NewConfig() *Config {
	var path = getConfigPath()
	createConfigIfNotExists(path)

	if path == "" {
		panic("config file not found, created on app root.")
	} else {
		println("config file found:", path)
	}

	c := &Config{}
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(c); err != nil {
		panic(err)
	}

	return c
}
