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

	err = SetHoliday("元旦", 1, 2024, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("春节", 7, 2024, 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("清明节", 1, 2024, 4, 5)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("劳动节", 1, 2024, 5, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("端午节", 1, 2023, 6, 10)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("中秋节", 2, 2023, 9, 29)
	if err != nil {
		t.Fatal(err)
	}
	err = SetHoliday("国庆节", 6, 2023, 10, 1)
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
