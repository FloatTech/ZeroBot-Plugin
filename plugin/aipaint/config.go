package aipaint

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/FloatTech/floatbox/file"
)

// 配置结构体
type serverConfig struct {
	BaseURL  string `json:"base_url"`
	Token    string `json:"token"`
	Interval int    `json:"interval"`
	file     string
}

func newServerConfig(file string) *serverConfig {
	return &serverConfig{
		file: file,
	}
}

func (cfg *serverConfig) update(baseURL, token string, interval int) (err error) {
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if token != "" {
		cfg.Token = token
	}
	cfg.Interval = interval
	reader, err := os.Create(cfg.file)
	if err != nil {
		return err
	}
	defer reader.Close()
	return json.NewEncoder(reader).Encode(cfg)
}

func (cfg *serverConfig) load() (err error) {
	if cfg.BaseURL != "" && cfg.Token != "" && cfg.Interval != 0 {
		return
	}
	if file.IsNotExist(cfg.file) {
		err = errors.New("no server config")
		return
	}
	reader, err := os.Open(cfg.file)
	if err != nil {
		return
	}
	defer reader.Close()
	err = json.NewDecoder(reader).Decode(cfg)
	return
}
