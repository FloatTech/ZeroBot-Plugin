package moyu

import (
	"fmt"
	"testing"
	"time"

	reg "github.com/fumiama/go-registry"
)

var sr = reg.NewRegedit("reilia.fumiama.top:32664", "", "fumiama", "--")

func TestGetHoliday(t *testing.T) {
	registry.Connect()
	h := GetHoliday("元旦")
	registry.Close()
	t.Fatal(h)
}

func TestSetHoliday(t *testing.T) {
	err := sr.Connect()
	if err != nil {
		t.Fatal(err)
	}

	err = SetHoliday("元旦", 1, 2025, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("春节", 7, 2025, 1, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("清明节", 1, 2025, 4, 4)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("劳动节", 1, 2025, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("端午节", 3, 2024, 6, 8)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("中秋节", 3, 2024, 9, 15)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("国庆节", 7, 2024, 10, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = sr.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func SetHoliday(name string, dur, year int, month time.Month, day int) error {
	return sr.Set("holiday/"+name, fmt.Sprintf("%d_%d_%d_%d", dur, year, month, day))
}
