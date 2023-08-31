package chess

import (
	"math"
	"testing"
)

func TestCalculateNewRate(t *testing.T) {
	type args struct {
		whiteRate  int
		blackRate  int
		whiteScore float64
		blackScore float64
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{
			name: "test1",
			args: args{
				whiteRate:  1613,
				blackRate:  1573,
				whiteScore: 0.5,
				blackScore: 0.5,
			},
			want:  1611,
			want1: 1575,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := calculateNewRate(tt.args.whiteRate, tt.args.blackRate, tt.args.whiteScore, tt.args.blackScore)
			if got != tt.want {
				t.Errorf("CalculateNewRate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("CalculateNewRate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_calculateException(t *testing.T) {
	type args struct {
		rate         int
		opponentRate int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test1",
			args: args{
				rate:         1613,
				opponentRate: 1573,
			},
			want: 0.5573116,
		},
		{
			name: "test2",
			args: args{
				rate:         1613,
				opponentRate: 1613,
			},
			want: 0.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateException(tt.args.rate, tt.args.opponentRate); math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("calculateException() = %v, want %v", got, tt.want)
			}
		})
	}
}
