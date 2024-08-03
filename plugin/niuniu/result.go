package niuniu

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"math/rand"
)

// fencing 击剑对决逻辑，返回对决结果和myLength的变化值
func fencing(myLength, oppoLength float64) (string, float64) {
	lossLimit := 0.25
	devourLimit := 0.27

	probability := rand.Intn(100) + 1

	switch {
	case oppoLength <= -100 && myLength > 0 && 10 < probability && probability <= 20:
		oppoLength *= 0.85
		change := -math.Min(math.Abs(lossLimit*myLength), math.Abs(1.5*myLength))
		myLength += change
		return fmt.Sprintf("对方身为魅魔诱惑了你，你同化成魅魔！当前长度%.2fcm！", myLength), change

	case oppoLength >= 100 && myLength > 0 && 10 < probability && probability <= 20:
		oppoLength *= 0.85
		change := -math.Min(math.Abs(devourLimit*myLength), math.Abs(1.5*myLength))
		myLength += change
		return fmt.Sprintf("对方以牛头人的荣誉摧毁了你的牛牛！当前长度%.2fcm！", myLength), change

	case myLength <= -100 && oppoLength > 0 && 10 < probability && probability <= 20:
		myLength *= 0.85
		change := math.Min(math.Abs(lossLimit*oppoLength), math.Abs(1.5*oppoLength))
		oppoLength -= change
		return fmt.Sprintf("你身为魅魔诱惑了对方，吞噬了对方部分长度！当前长度%.2fcm！", myLength), change

	case myLength >= 100 && oppoLength > 0 && 10 < probability && probability <= 20:
		myLength *= 0.85
		change := math.Min(math.Abs(devourLimit*oppoLength), math.Abs(1.5*oppoLength))
		oppoLength -= change
		return fmt.Sprintf("你以牛头人的荣誉摧毁了对方的牛牛！当前长度%.2fcm！", myLength), change

	default:
		return determineResultBySkill(myLength, oppoLength)
	}
}

// determineResultBySkill 根据击剑技巧决定结果
func determineResultBySkill(myLength, oppoLength float64) (string, float64) {
	probability := rand.Intn(100) + 1
	winProbability := calculateWinProbability(myLength, oppoLength) * 100

	if 0 < probability && float64(probability) <= winProbability {
		return applySkill(myLength, oppoLength, true)
	} else {
		return applySkill(myLength, oppoLength, false)
	}
}

// calculateWinProbability 计算胜率
func calculateWinProbability(heightA, heightB float64) float64 {
	pA := 0.9
	heightRatio := math.Max(heightA, heightB) / math.Min(heightA, heightB)
	reductionRate := 0.1 * (heightRatio - 1)
	reduction := pA * reductionRate
	adjustedPA := pA - reduction
	return math.Max(adjustedPA, 0.01)
}

// applySkill 应用击剑技巧并生成结果
func applySkill(myLength, oppoLength float64, increaseLength1 bool) (string, float64) {
	reduce := fence(oppoLength)
	var change float64

	if increaseLength1 {
		myLength += reduce
		oppoLength -= 0.8 * reduce
		change = reduce

		if myLength < 0 {
			return fmt.Sprintf("哦吼！？你的牛牛在长大欸！长大了%.2fcm！", reduce), change
		}
		return fmt.Sprintf("你以绝对的长度让对方屈服了呢！你的长度增加%.2fcm，当前长度%.2fcm！", reduce, myLength), change

	} else {
		myLength -= reduce
		oppoLength += 0.8 * reduce
		change = -reduce

		if myLength < 0 {
			return fmt.Sprintf("哦吼！？看来你的牛牛因为击剑而凹进去了呢🤣🤣🤣！凹进去了%.2fcm！", reduce), change
		}
		return fmt.Sprintf("对方以绝对的长度让你屈服了呢！你的长度减少%.2fcm，当前长度%.2fcm！", reduce, myLength), change
	}
}

// fence 简单模拟击剑技巧效果
func fence(oppoLength float64) float64 {
	return float64(rand.Intn(5)+1) + rand.Float64()
}

// randomLong 生成一个随机的数值
func randomLong() decimal.Decimal {
	return decimal.NewFromFloat(float64(rand.Intn(9)+1) + float64(rand.Intn(100))/100)
}

// hitGlue 调整传入的值
func hitGlue(l decimal.Decimal) float64 {
	l = l.Sub(decimal.NewFromInt(1))
	randomFactor := decimal.NewFromFloat(rand.Float64())
	adjustedValue := randomFactor.Mul(l).Div(decimal.NewFromInt(2))
	f, _ := adjustedValue.Float64()
	return f
}
