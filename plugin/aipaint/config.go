package aipaint

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/FloatTech/floatbox/file"
)

// 配置结构体
type serverConfig struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
	file    string
}

func newServerConfig(file string) *serverConfig {
	return &serverConfig{
		file: file,
	}
}

func (cfg *serverConfig) save(baseURL, token string) (err error) {
	cfg.BaseURL = baseURL
	cfg.Token = token
	reader, err := os.Create(cfg.file)
	if err != nil {
		return err
	}
	defer reader.Close()
	return json.NewEncoder(reader).Encode(cfg)
}

func (cfg *serverConfig) load() (aipaintServer, token string, err error) {
	if cfg.BaseURL != "" && cfg.Token != "" {
		aipaintServer = cfg.BaseURL
		token = cfg.Token
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
	aipaintServer = cfg.BaseURL
	token = cfg.Token
	return
}
