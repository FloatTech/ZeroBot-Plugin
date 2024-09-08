// Package niuniu 牛牛大作战
package niuniu

import (
	"fmt"
	"math"
	"math/rand"
)

func useWeiGe(niuniu float64) (string, float64) {
	reduce := math.Abs(hitGlue(niuniu))
	niuniu += reduce
	return randomChoice([]string{
		fmt.Sprintf("哈哈，你这一用道具，牛牛就像是被激发了潜能，增加了%.2fcm！看来今天是个大日子呢！", reduce),
		fmt.Sprintf("你这是用了什么神奇的道具？牛牛竟然增加了%.2fcm，简直是牛气冲天！", reduce),
		fmt.Sprintf("“使用道具后，你的牛牛就像是开启了加速模式，一下增加了%.2fcm，这成长速度让人惊叹！", reduce),
	}), niuniu
}

func usePhilter(niuniu float64) (string, float64) {
	reduce := math.Abs(hitGlue(niuniu))
	niuniu -= reduce
	return randomChoice([]string{
		fmt.Sprintf("你使用媚药,咿呀咿呀一下使当前长度发生了一些变化，当前长度%.2f", niuniu),
		fmt.Sprintf("看来你追求的是‘微观之美’，故意使用道具让牛牛凹进去了%.2fcm！", reduce),
		fmt.Sprintf("‘缩小奇迹’在你身上发生了，牛牛凹进去了%.2fcm，你的选择真是独特！", reduce),
	}), niuniu
}

func useArtifact(myLength, adduserniuniu float64) (string, float64, float64) {
	difference := myLength - adduserniuniu
	var (
		change float64
	)
	if difference > 0 {
		change = hitGlue(myLength + adduserniuniu)
	} else {
		change = hitGlue((myLength + adduserniuniu) / 2)
	}
	myLength += change
	return randomChoice([]string{
		fmt.Sprintf("凭借神秘道具的力量，你让对方在你的长度面前俯首称臣！你的长度增加了%.2fcm，当前长度达到了%.2fcm", change, myLength),
		fmt.Sprintf("神器在手，天下我有！你使用道具后，长度猛增%.2fcm，现在的总长度是%.2fcm，无人能敌！", change, myLength),
		fmt.Sprintf("这就是道具的魔力！你轻松增加了%.2fcm，让对手望尘莫及，当前长度为%.2fcm！", change, myLength),
		fmt.Sprintf("道具一出，谁与争锋！你的长度因道具而增长%.2fcm，现在的长度是%.2fcm，霸气尽显！", change, myLength),
		fmt.Sprintf("使用道具的你，如同获得神助！你的长度增长了%.2fcm，达到%.2fcm的惊人长度，胜利自然到手！", change, myLength),
	}), myLength, adduserniuniu - change/1.3
}

func useShenJi(myLength, adduserniuniu float64) (string, float64, float64) {
	difference := myLength - adduserniuniu
	var (
		change float64
	)
	if difference > 0 {
		change = hitGlue(myLength + adduserniuniu)
	} else {
		change = hitGlue((myLength + adduserniuniu) / 2)
	}
	myLength -= change
	var r string
	if myLength > 0 {
		r = randomChoice([]string{
			fmt.Sprintf("哦吼！？看来你的牛牛因为使用了神秘道具而缩水了呢🤣🤣🤣！缩小了%.2fcm！", change),
			fmt.Sprintf("哈哈，看来这个道具有点儿调皮，让你的长度缩水了%.2fcm！现在你的长度是%.2fcm，下次可得小心使用哦！", change, myLength),
			fmt.Sprintf("使用道具后，你的牛牛似乎有点儿害羞，缩水了%.2fcm！现在的长度是%.2fcm，希望下次它能挺直腰板！", change, myLength),
			fmt.Sprintf("哎呀，这个道具的效果有点儿意外，你的长度减少了%.2fcm，现在只有%.2fcm了！下次选道具可得睁大眼睛！", change, myLength),
		})
	} else {
		r = randomChoice([]string{
			fmt.Sprintf("哦哟，小姐姐真是玩得一手好游戏，使用道具后数值又降低了%.2fcm，小巧得更显魅力！", change),
			fmt.Sprintf("看来小姐姐喜欢更加精致的风格，使用道具后，数值减少了%.2fcm，更加迷人了！", change),
			fmt.Sprintf("小姐姐的每一次变化都让人惊喜，使用道具后，数值减少了%.2fcm，更加优雅动人！", change),
			fmt.Sprintf("小姐姐这是在展示什么是真正的精致小巧，使用道具后，数值减少了%.2fcm，美得不可方物！", change),
		})
	}
	return r, myLength, adduserniuniu + 0.7*change
}

func generateRandomStingTwo(niuniu float64) (string, float64) {
	probability := rand.Intn(100 + 1)
	reduce := math.Abs(hitGlue(niuniu))
	switch {
	case probability <= 40:
		niuniu += reduce
		return randomChoice([]string{
			fmt.Sprintf("你嘿咻嘿咻一下，促进了牛牛发育，牛牛增加%.2fcm了呢！", reduce),
			fmt.Sprintf("你打了个舒服痛快的🦶呐，牛牛增加了%.2fcm呢！", reduce),
		}), niuniu
	case probability <= 60:
		return randomChoice([]string{
			"你打了个🦶，但是什么变化也没有，好奇怪捏~",
			"你的牛牛刚开始变长了，可过了一会又回来了，什么变化也没有，好奇怪捏~",
		}), niuniu
	default:
		niuniu -= reduce
		if niuniu < 0 {
			return randomChoice([]string{
				fmt.Sprintf("哦吼！？看来你的牛牛凹进去了%.2fcm呢！", reduce),
				fmt.Sprintf("你突发恶疾！你的牛牛凹进去了%.2fcm！", reduce),
				fmt.Sprintf("笑死，你因为打🦶过度导致牛牛凹进去了%.2fcm！🤣🤣🤣", reduce),
			}), niuniu
		} else {
			return randomChoice([]string{
				fmt.Sprintf("阿哦，你过度打🦶，牛牛缩短%.2fcm了呢！", reduce),
				fmt.Sprintf("你的牛牛变长了很多，你很激动地继续打🦶，然后牛牛缩短了%.2fcm呢！", reduce),
				fmt.Sprintf("小打怡情，大打伤身，强打灰飞烟灭！你过度打🦶，牛牛缩短了%.2fcm捏！", reduce),
			}), niuniu
		}
	}
}

func generateRandomString(niuniu float64) string {
	switch {
	case niuniu <= -100:
		return "wtf？你已经进化成魅魔了！魅魔在击剑时有20%的几率消耗自身长度吞噬对方牛牛呢。"
	case niuniu <= -50:
		return "嗯....好像已经穿过了身体吧..从另一面来看也可以算是凸出来的吧?"
	case niuniu <= -25:
		return randomChoice([]string{
			"这名女生，你的身体很健康哦！",
			"WOW,真的凹进去了好多呢！",
			"你已经是我们女孩子的一员啦！",
		})
	case niuniu <= -10:
		return randomChoice([]string{
			"你已经是一名女生了呢，",
			"从女生的角度来说，你发育良好(,",
			"你醒啦？你已经是一名女孩子啦！",
			"唔...可以放进去一根手指了都...",
		})
	case niuniu <= 0:
		return randomChoice([]string{
			"安了安了，不要伤心嘛，做女生有什么不好的啊。",
			"不哭不哭，摸摸头，虽然很难再长出来，但是请不要伤心啦啊！",
			"加油加油！我看好你哦！",
			"你醒啦？你现在已经是一名女孩子啦！",
		})
	case niuniu <= 10:
		return randomChoice([]string{
			"你行不行啊？细狗！",
			"虽然短，但是小小的也很可爱呢。",
			"像一只蚕宝宝。",
			"长大了。",
		})
	case niuniu <= 25:
		return randomChoice([]string{
			"唔...没话说",
			"已经很长了呢！",
		})
	case niuniu <= 50:
		return randomChoice([]string{
			"话说这种真的有可能吗？",
			"厚礼谢！",
		})
	case niuniu <= 100:
		return randomChoice([]string{
			"已经突破天际了嘛...",
			"唔...这玩意应该不会变得比我高吧？",
			"你这个长度会死人的...！",
			"你马上要进化成牛头人了！！",
			"你是什么怪物，不要过来啊！！",
		})
	default:
		return "惊世骇俗！你已经进化成牛头人了！牛头人在击剑时有20%的几率消耗自身长度吞噬对方牛牛呢。"
	}
}

// fencing 击剑对决逻辑，返回对决结果和myLength的变化值
func fencing(myLength, oppoLength float64) (string, float64, float64) {
	devourLimit := 0.27

	probability := rand.Intn(100) + 1

	switch {
	case oppoLength <= -100 && myLength > 0 && 10 < probability && probability <= 20:
		oppoLength *= 0.85
		change := hitGlue(oppoLength) + rand.Float64()*math.Log2(math.Abs(0.5*(myLength+oppoLength)))
		myLength = change
		return fmt.Sprintf("对方身为魅魔诱惑了你，你同化成魅魔！当前长度%.2fcm！", -myLength), -myLength, oppoLength

	case oppoLength >= 100 && myLength > 0 && 10 < probability && probability <= 20:
		oppoLength *= 0.85
		change := math.Min(math.Abs(devourLimit*myLength), math.Abs(1.5*myLength))
		myLength += change
		return fmt.Sprintf("对方以牛头人的荣誉摧毁了你的牛牛！当前长度%.2fcm！", myLength), myLength, oppoLength

	case myLength <= -100 && oppoLength > 0 && 10 < probability && probability <= 20:
		myLength *= 0.85
		change := hitGlue(myLength+oppoLength) + rand.Float64()*math.Log2(math.Abs(0.5*(myLength+oppoLength)))
		oppoLength -= change
		myLength -= change
		return fmt.Sprintf("你身为魅魔诱惑了对方，吞噬了对方部分长度！当前长度%.2fcm！", myLength), myLength, oppoLength

	case myLength >= 100 && oppoLength > 0 && 10 < probability && probability <= 20:
		myLength -= oppoLength
		oppoLength = 0.01
		return fmt.Sprintf("你以牛头人的荣誉摧毁了对方的牛牛！当前长度%.2fcm！", myLength), myLength, oppoLength

	default:
		return determineResultBySkill(myLength, oppoLength)
	}
}

// determineResultBySkill 根据击剑技巧决定结果
func determineResultBySkill(myLength, oppoLength float64) (string, float64, float64) {
	probability := rand.Intn(100) + 1
	winProbability := calculateWinProbability(myLength, oppoLength) * 100
	return applySkill(myLength, oppoLength,
		0 < probability && float64(probability) <= winProbability)
}

// calculateWinProbability 计算胜率
func calculateWinProbability(heightA, heightB float64) float64 {
	var pA float64
	if heightA > heightB {
		pA = 0.7 + 0.2*(heightA-heightB)/heightA
	} else {
		pA = 0.7 - 0.2*(heightB-heightA)/heightB
	}
	heightRatio := math.Max(heightA, heightB) / math.Min(heightA, heightB)
	reductionRate := 0.1 * (heightRatio - 1)
	reduction := pA * reductionRate
	adjustedPA := pA - reduction
	return math.Max(adjustedPA, 0.01)
}

// applySkill 应用击剑技巧并生成结果
func applySkill(myLength, oppoLength float64, increaseLength1 bool) (string, float64, float64) {
	reduce := fence(oppoLength)
	if increaseLength1 {
		myLength += reduce
		oppoLength -= 0.8 * reduce
		if myLength < 0 {
			return fmt.Sprintf("哦吼！？你的牛牛在长大欸！长大了%.2fcm！", reduce), myLength, oppoLength
		}
		return fmt.Sprintf("你以绝对的长度让对方屈服了呢！你的长度增加%.2fcm，当前长度%.2fcm！", reduce, myLength), myLength, oppoLength

	}
	myLength -= reduce
	oppoLength += 0.8 * reduce
	if myLength < 0 {
		return fmt.Sprintf("哦吼！？看来你的牛牛因为击剑而凹进去了呢🤣🤣🤣！凹进去了%.2fcm！", reduce), myLength, oppoLength
	}
	return fmt.Sprintf("对方以绝对的长度让你屈服了呢！你的长度减少%.2fcm，当前长度%.2fcm！", reduce, myLength), myLength, oppoLength

}

// fence
func fence(rd float64) float64 {
	r := hitGlue(rd)*2 + rand.Float64()*math.Log2(rd)
	if rand.Intn(2) == 1 {
		return rd - rand.Float64()*r
	}
	return float64(int(r * rand.Float64()))
}

func hitGlue(l float64) float64 {
	if l == 0 {
		l = 0.1
	}
	l = math.Abs(l)
	switch {
	case l > 1 && l <= 10:
		return rand.Float64() * math.Log2(l)
	case 10 < l && l <= 100:
		return rand.Float64() * math.Log2(l*1.5) / 2
	case 100 < l && l <= 1000:
		return rand.Float64() * math.Log10(l*1.5) / 2
	case l > 1000:
		return rand.Float64() * math.Log10(l) / 2
	default:
		return rand.Float64()
	}
}
