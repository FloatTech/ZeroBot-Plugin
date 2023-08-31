package chess

import (
	"math"
)

// calculateNewRate calculate new rate of the player
func calculateNewRate(whiteRate, blackRate int, whiteScore, blackScore float64) (int, int) {
	k := getKFactor(whiteRate, blackRate)
	exceptionWhite := calculateException(whiteRate, blackRate)
	exceptionBlack := calculateException(blackRate, whiteRate)
	whiteRate = calculateRate(whiteRate, whiteScore, exceptionWhite, k)
	blackRate = calculateRate(blackRate, blackScore, exceptionBlack, k)
	return whiteRate, blackRate
}

func calculateException(rate int, opponentRate int) float64 {
	return 1.0 / (1.0 + math.Pow(10.0, float64(opponentRate-rate)/400.0))
}

func calculateRate(rate int, score float64, exception float64, k int) int {
	newRate := int(math.Round(float64(rate) + float64(k)*(score-exception)))
	if newRate < 1 {
		newRate = 1
	}
	return newRate
}

func getKFactor(rateA, rateB int) int {
	if rateA > 2400 && rateB > 2400 {
		return 16
	}
	if rateA > 2100 && rateB > 2100 {
		return 24
	}
	return 32
}
