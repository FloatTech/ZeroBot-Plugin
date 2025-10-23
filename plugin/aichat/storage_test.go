package aichat

import (
	"testing"

	"github.com/FloatTech/zbputils/ctxext"
)

func TestStorage_rate(t *testing.T) {
	s := storage(ctxext.Storage(0))

	// 测试默认值
	if rate := s.rate(); rate != 0 {
		t.Errorf("default rate() = %v, want 0", rate)
	}

	// 设置值并测试
	s = storage((ctxext.Storage)(s).Set(int64(100), bitmaprate))
	if rate := s.rate(); rate != 100 {
		t.Errorf("rate() after set = %v, want 100", rate)
	}
}

func TestStorage_temp(t *testing.T) {
	s := storage(ctxext.Storage(0))

	tests := []struct {
		name     string
		setValue int64
		expected float32
	}{
		{"default temp (0)", 0, 0.70}, // 默认值 70/100
		{"valid temp 50", 50, 0.50},   // 50/100 = 0.50
		{"valid temp 80", 80, 0.80},   // 80/100 = 0.80
		{"max temp 100", 100, 1.00},   // 100/100 = 1.00
		{"over max temp", 127, 1.00},  // 限制为 100/100 = 1.00
		{"negative temp", -10, 0.70},  // 默认值 70/100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = storage((ctxext.Storage)(s).Set(tt.setValue, bitmaptemp))

			result := s.temp()
			if result != tt.expected {
				t.Errorf("temp() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStorage_noagent(t *testing.T) {
	s := storage(ctxext.Storage(0))

	// 测试默认值
	if noagent := s.noagent(); noagent != false {
		t.Errorf("default noagent() = %v, want false", noagent)
	}

	// 设置为 true 并测试
	s = storage((ctxext.Storage)(s).Set(1, bitmapnagt))
	if noagent := s.noagent(); noagent != true {
		t.Errorf("noagent() after set true = %v, want true", noagent)
	}
}

func TestStorage_norecord(t *testing.T) {
	s := storage(ctxext.Storage(0))

	// 测试默认值
	if norecord := s.norecord(); norecord != false {
		t.Errorf("default norecord() = %v, want false", norecord)
	}

	// 设置为 true 并测试
	s = storage((ctxext.Storage)(s).Set(1, bitmapnrec))
	if norecord := s.norecord(); norecord != true {
		t.Errorf("norecord() after set true = %v, want true", norecord)
	}
}

func TestStorage_noreplyat(t *testing.T) {
	s := storage(ctxext.Storage(0))

	// 测试默认值
	if noreplyat := s.noreplyat(); noreplyat != false {
		t.Errorf("default noreplyat() = %v, want false", noreplyat)
	}

	// 设置为 true 并测试
	s = storage((ctxext.Storage)(s).Set(1, bitmapnrat))
	if noreplyat := s.noreplyat(); noreplyat != true {
		t.Errorf("noreplyat() after set true = %v, want true", noreplyat)
	}
}

func TestStorage_Integration(t *testing.T) {
	s := storage(ctxext.Storage(0))

	// 设置各种值
	s = storage((ctxext.Storage)(s).Set(int64(75), bitmaprate))
	s = storage((ctxext.Storage)(s).Set(int64(85), bitmaptemp))
	s = storage((ctxext.Storage)(s).Set(1, bitmapnagt))
	s = storage((ctxext.Storage)(s).Set(0, bitmapnrec))
	s = storage((ctxext.Storage)(s).Set(1, bitmapnrat))

	// 验证所有方法
	if rate := s.rate(); rate != 75 {
		t.Errorf("rate() = %v, want 75", rate)
	}

	if temp := s.temp(); temp != 0.85 {
		t.Errorf("temp() = %v, want 0.85", temp)
	}

	if noagent := s.noagent(); !noagent {
		t.Errorf("noagent() = %v, want true", noagent)
	}

	if norecord := s.norecord(); norecord {
		t.Errorf("norecord() = %v, want false", norecord)
	}

	if noreplyat := s.noreplyat(); !noreplyat {
		t.Errorf("noreplyat() = %v, want true", noreplyat)
	}
}

func BenchmarkStorage_rate(b *testing.B) {
	s := storage(ctxext.Storage(0))

	s = storage((ctxext.Storage)(s).Set(int64(100), bitmaprate))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.rate()
	}
}

func BenchmarkStorage_temp(b *testing.B) {
	s := storage(ctxext.Storage(0))

	s = storage((ctxext.Storage)(s).Set(int64(80), bitmaptemp))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.temp()
	}
}
