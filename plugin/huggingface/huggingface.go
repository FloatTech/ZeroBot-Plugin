// Package huggingface ai界的github
package huggingface

import (
	"bytes"
	"encoding/json"

	"github.com/FloatTech/floatbox/web"
)

const (
	huggingfaceSpaceHTTPS = "https://hf.space"
	embed                 = huggingfaceSpaceHTTPS + "/embed"
	pushPath              = "/api/queue/push/"
	statusPath            = "/api/queue/status/"
	defaultAction         = "predict"
	defaultSessionHash    = "zerobot"
	completeStatus        = "COMPLETE"
	wssCompleteStatus     = "process_completed"
	timeoutMax            = 300

	huggingfaceSpaceWss = "wss://spaces.huggingface.tech"
	joinPath            = "/queue/join"

	predictPath = "/api/predict/"
)

type pushRequest struct {
	Action      string        `json:"action,omitempty"`
	FnIndex     int           `json:"fn_index"`
	Data        []interface{} `json:"data"`
	SessionHash string        `json:"session_hash"`
}

type pushResponse struct {
	Hash          string `json:"hash"`
	QueuePosition int    `json:"queue_position"`
}

type statusRequest struct {
	Hash string `json:"hash"`
}

type statusResponse struct {
	Status string `json:"status"`
	Data   struct {
		Data            []interface{} `json:"data"`
		Duration        float64       `json:"duration"`
		AverageDuration float64       `json:"average_duration"`
	}
}

func push(pushURL string, pushReq pushRequest) (pushRes pushResponse, err error) {
	b, err := json.Marshal(pushReq)
	if err != nil {
		return
	}
	data, err := web.PostData(pushURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &pushRes)
	// 上传数据休息
	return
}

func status(statusURL string, statusReq statusRequest) (data []byte, err error) {
	b, err := json.Marshal(statusReq)
	if err != nil {
		return
	}
	data, err = web.PostData(statusURL, "application/json", bytes.NewReader(b))
	return
}
