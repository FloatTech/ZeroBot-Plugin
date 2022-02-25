package reborn

import (
	wr "github.com/mroth/weightedrand"
)

type rate []struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

var (
	areac     *wr.Chooser
	gender, _ = wr.NewChooser(
		wr.Choice{Item: "男孩子", Weight: 50707},
		wr.Choice{Item: "女孩子", Weight: 48292},
		wr.Choice{Item: "雌雄同体", Weight: 1001},
	)
)

func randcoun() string {
	return areac.Pick().(string)
}

func randgen() string {
	return gender.Pick().(string)
}
