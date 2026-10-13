package wordcount

import (
	"bytes"
	"image/png"
	"testing"

	chart "github.com/wcharczuk/go-chart/v2"
)

func TestWordCountChartRendersPositiveFrequencies(t *testing.T) {
	for _, tc := range []struct {
		name   string
		counts []float64
	}{
		{"single occurrence", []float64{1}},
		{"single frequent word", []float64{7}},
		{"equal single occurrences", []float64{1, 1, 1}},
		{"equal frequent words", []float64{7, 7, 7}},
		{"different frequencies", []float64{1, 3, 7}},
		{"large equal frequencies", []float64{1000000, 1000000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bars := make([]chart.Value, len(tc.counts))
			for i, count := range tc.counts {
				bars[i] = chart.Value{Label: "word", Value: count}
			}
			graph := chart.BarChart{Height: 500, BarWidth: 25, Bars: bars, YAxis: wordCountYAxis(bars)}
			var output bytes.Buffer
			if err := graph.Render(chart.PNG, &output); err != nil {
				t.Fatalf("render actual chart: %v", err)
			}
			img, err := png.Decode(bytes.NewReader(output.Bytes()))
			if err != nil {
				t.Fatalf("decode actual PNG: %v", err)
			}
			if img.Bounds().Dx() <= 0 || img.Bounds().Dy() != 500 {
				t.Fatalf("invalid image size: %v", img.Bounds())
			}
		})
	}
}
