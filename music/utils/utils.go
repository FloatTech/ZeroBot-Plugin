package utils

import (
	"fmt"

	zero "github.com/wdvxdr1123/ZeroBot"
)

type CQMusic struct {
	Type    string
	Url     string
	Audio   string
	Title   string
	Content string
	Image   string
}

func SendError(event zero.Event, err error) {
	zero.Send(event, fmt.Sprintf("ERROR: %v", err))
}
