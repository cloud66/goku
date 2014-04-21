package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ServerAddress        string
	ServerPort           int
	HoneybadgerApi       string
	HttpNotifierEndpoint string
}

func (c *Config) Load() error {
	var config Config

	file := filepath.Join(gokuHome(), "goku.toml")
	if _, err := os.Stat(file); err == nil {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if _, err := toml.Decode(string(data), &config); err != nil {
			return err
		}
	}

	c.ServerAddress = config.ServerAddress
	c.HoneybadgerApi = config.HoneybadgerApi
	c.ServerPort = config.ServerPort
	c.HttpNotifierEndpoint = config.HttpNotifierEndpoint

	if c.HoneybadgerApi == "" {
		c.HoneybadgerApi = "2188ca35"
	}
	if c.ServerAddress == "" {
		c.ServerAddress = "127.0.0.1"
	}
	if c.ServerPort == 0 {
		c.ServerPort = 9800
	}

	return nil
}

func gokuHome() string {
	return filepath.Join(os.Getenv("HOME"), ".goku")
}
