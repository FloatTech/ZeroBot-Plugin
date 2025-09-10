// Package bank 银行
package bank

import (
	"strconv"
	"strings"
	"time"
	"fmt"
	"math"
	"sort"
	"regexp"

    "github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 活期利率范围（1% - 3%）
const (
    CurrentRateMin = 0.01 // 最小活期利率
    CurrentRateMax = 0.03 // 最大活期利率
)

// 定期存款利率表(根据期限不同)
var fixedDepositRates = map[int]float64{
    3:  0.1,        // 3 天期， 利率 10%
    7:  0.36,       // 7 天期， 利率 36%
    15: 1,          // 15 天期，利率 100%
    30: 3,          // 30 天期，利率 300%
}

// LoanDailyRate 贷款日利率
const LoanDailyRate = 25

// 重生用户利率调整比例(百分比)
const (
	// 活期存款利率下调比例(例如0.1表示下调10%)
	CurrentRateDecreasePercent = 0.1
	// 定期存款利率下调比例
	FixedRateDecreasePercent = 0.1
	// 贷款利率上调比例
	LoanRateIncreasePercent = 0.5
	// 可贷款金额上调比例
	LoanAmountIncreasePercent = 0.5
	// 基础可贷款金额上限
	BaseMaxLoanAmount = 20000
)

const (
    // RebirthBuffName 重生buff的名称
	RebirthBuffName = "rebirth"
    // MaxRebirthStack buff最大能叠加的次数
	MaxRebirthStack = 3
    // RebirthBuffDuration 每层buff的持续时间
    RebirthBuffDuration = 7 * 24 * time.Hour
)

var (
    // 注册插件引擎
    engine = control.Register("bank", &ctrl.Options[*zero.Ctx]{
        DisableOnDefault: false,
        Brief:            "银行",
        Help: "- 银行活期存款 [金额/全部] (活期利率每天都会在1%~3%波动)\n" +
            "- 银行定期存款 [金额] [期限(3/7/15/30天)]\n" +
            "- 银行活期取款 [金额/全部] \n" +
            "- 银行定期取款 [全部] [定期款序列号] \n" +
            "- 银行贷款 [金额] [期限(天)](到了还款日后会强制还款哦~)\n" +
            "- 银行还款 [金额]\n" +
            "- 查看我的存款\n" +
            "- 查看我的贷款\n" +
            "- 查看今日利率\n" +
            "- 我要信仰之跃\n",
        PrivateDataFolder: "bank",
    })

    accountPath   = engine.DataFolder() + "accounts/"
    interestPath  = engine.DataFolder() + "interest.yaml"
    accounts      = make(map[int64]*Account) // 内存中的账户数据
    dailyInterest = &InterestInfo{CurrentRate: 0.01} // 默认活期利率 1%
)

var numRegex = regexp.MustCompile(`\d+`)

func init() {
	// 银行活期存款命令处理
	engine.OnPrefix("银行活期存款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		param := strings.TrimSpace(ctx.State["args"].(string))
		if param == "" {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请输入存款金额，例如：银行活期存款 100 或 银行活期存款 全部"))
			return
		}

		uid := ctx.Event.UserID
		var amount int
		var err error

		// 处理“全部”存款的情况
		if param == "全部" {
			amount = wallet.GetWalletOf(uid) // 获取钱包全部余额
		} else {
			// 原逻辑：提取数字金额
			numStr := numRegex.FindString(param)
			if numStr == "" {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("输入的金额无效，请输入正整数或'全部'，例如：100 或 全部"))
				return
			}
			amount, err = strconv.Atoi(numStr)
			if err != nil || amount <= 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("输入的金额无效，请输入正整数或'全部'"))
				return
			}
		}

		// 检查金额有效性（钱包余额是否足够）
		if wallet.GetWalletOf(uid) < amount {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("钱包余额不足，无法存款"))
			return
		}
		if amount <= 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("存款金额必须大于0"))
			return
		}

		// 执行存款操作（扣除钱包金额，增加活期余额）
		err = wallet.InsertWalletOf(uid, -amount)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("存款失败：", err))
			return
		}

		account := GetOrCreateAccount(uid)
		account.CurrentBalance += amount
		account.LastUpdate = time.Now()
		_ = SaveAccount(account)

		stacks := getRebirthStacks(account)
		currentBaseRate := dailyInterest.CurrentRate
		actualRate := currentBaseRate
		rateDesc := ""
		if stacks > 0 {
			actualRate *= (1 - CurrentRateDecreasePercent*float64(stacks))
			rateDesc = fmt.Sprintf("（重生buff x%d生效，基础利率下调%.0f%%）", 
				stacks, CurrentRateDecreasePercent*float64(stacks)*100)
		}

		rateInfo := fmt.Sprintf(
			"活期基础利率：%.2f%%/天\n你的实际利率：%.2f%%/天%s",
			currentBaseRate*100,
			actualRate*100,
			rateDesc,
		)

        walletBalance := wallet.GetWalletOf(uid)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(
			fmt.Sprintf(
				"活期存款成功！\n当前活期余额：%d%s\n当前钱包余额：%d%s\n%s\n当前重生buff: %d/%d层",
				account.CurrentBalance,
				wallet.GetWalletName(),
                walletBalance,
                wallet.GetWalletName(),
				rateInfo,
				stacks,
				MaxRebirthStack,
			),
		))
	})
	// 银行定期存款命令处理
	engine.OnPrefix("银行定期存款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		params := strings.Fields(strings.TrimSpace(ctx.State["args"].(string)))
		if len(params) < 2 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("格式错误！示例：银行定期存款 1000 3（存款1000，期限3天）"))
			return
		}

		amount, err := strconv.Atoi(params[0])
		term, errTerm := strconv.Atoi(params[1])
		if err != nil || errTerm != nil || amount <= 0 || term <= 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("金额或期限无效，请输入正整数！"))
			return
		}

		baseRate, exists := fixedDepositRates[term]
		if !exists {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("不支持的存款期限，支持的期限：3、7、15、30天"))
			return
		}

		uid := ctx.Event.UserID
		if wallet.GetWalletOf(uid) < amount {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("钱包余额不足，无法存款"))
			return
		}

		err = wallet.InsertWalletOf(uid, -amount)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("存款失败：", err))
			return
		}

		account := GetOrCreateAccount(uid)
		stacks := getRebirthStacks(account)

		actualRate := baseRate
		rateAdjustDesc := ""
		if stacks > 0 {
			actualRate *= (1 - FixedRateDecreasePercent*float64(stacks))
			rateAdjustDesc = fmt.Sprintf("（重生buff x%d生效，利率基础下调%.0f%%）", 
				stacks, FixedRateDecreasePercent*float64(stacks)*100)
		}

		interest := float64(amount) * actualRate
		totalAmount := float64(amount) + interest

		rateInfo := fmt.Sprintf(
			"基础利率：%.2f%%\n实际利率：%.2f%%%s",
			baseRate*100,
			actualRate*100,
			rateAdjustDesc,
		)

		now := time.Now()
		fixedDeposit := FixedDeposit{
			Amount:       amount,
			TermDays:     term,
			Rate:         actualRate,
			StartDate:    now,
			MaturityDate: now.AddDate(0, 0, term),
		}

		account.FixedDeposits = append(account.FixedDeposits, fixedDeposit)
		account.LastUpdate = now
		if err := SaveAccount(account); err != nil {
			// 错误处理逻辑：记录错误日志
			fmt.Printf("保存账户失败: %v", err)
			return
		}

		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf(
			"定期存款成功！\n金额：%d%s\n期限：%d天\n%s\n可得利息：%.2f%s\n可得总资金（本金+利息）：%.2f%s\n到期日：%s\n当前重生buff: %d/%d层",
			amount, wallet.GetWalletName(),
			term,
			rateInfo,
			interest, wallet.GetWalletName(),
			totalAmount, wallet.GetWalletName(),
			fixedDeposit.MaturityDate.Format("2006-01-02"),
			stacks,
			MaxRebirthStack,
		)))
	})

	// 银行活期取款
	engine.OnPrefix("银行活期取款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		params := strings.Fields(strings.TrimSpace(ctx.State["args"].(string)))
		if len(params) < 1 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请输入取款金额，例如：\n银行活期取款 100\n银行活期取款 全部"))
			return
		}

		amountStr := params[0]
		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)
		var amount int
		var err error

		if amountStr == "全部" {
			amount = account.CurrentBalance
		} else {
			amount, err = strconv.Atoi(amountStr)
			if err != nil || amount <= 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("输入的金额无效，请输入正整数或'全部'"))
				return
			}
		}

		if account.CurrentBalance < amount {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("活期余额不足，无法取款"))
			return
		}

		account.CurrentBalance -= amount
		account.LastUpdate = time.Now()
		if err := SaveAccount(account); err != nil {
			fmt.Printf("更新账户保存失败: %v", err)
			return
		}

		err = wallet.InsertWalletOf(uid, amount)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("取款失败：", err))
			return
		}

        walletBalance := wallet.GetWalletOf(uid)
        ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(
            fmt.Sprintf(
                "活期取款成功！\n当前活期余额：%d%s\n当前钱包余额：%d%s",
                account.CurrentBalance,
                wallet.GetWalletName(),
                walletBalance,
                wallet.GetWalletName(),
            ),
        ))
	})

	// 银行定期取款
	engine.OnPrefix("银行定期取款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		params := strings.Fields(strings.TrimSpace(ctx.State["args"].(string)))
		if len(params) < 2 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请输入取款金额和定期存款序号，例如：\n银行定期取款 全部 1\n银行定期取款 100 1（取出第1笔定期存款的部分金额）"))
			return
		}

		amountStr := params[0]
		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)
		
		// 解析定期存款序号
		index, err := strconv.Atoi(params[1])
		if err != nil || index < 1 || index > len(account.FixedDeposits) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("无效的定期存款序号"))
			return
		}
		index-- // 转换为0-based索引

		fixedDeposit := &account.FixedDeposits[index]
		now := time.Now()
		isMature := now.After(fixedDeposit.MaturityDate) || now.Equal(fixedDeposit.MaturityDate)

		var totalAmount int
		var amount int

		// 处理全部取出的情况
		if amountStr == "全部" {
			if isMature {
				// 到期取款，计算利息
				interest := int(float64(fixedDeposit.Amount) * fixedDeposit.Rate)
				totalAmount = fixedDeposit.Amount + interest
			} else {
				// 提前取款，扣除手续费(损失50%本金)
				penalty := fixedDeposit.Amount / 2
				totalAmount = fixedDeposit.Amount - penalty
				ctx.SendChain(message.Text("提示：定期存款未到期提前取出，将扣除50%本金作为手续费"))
			}
		} else {
			// 处理部分取出的情况
			amount, err = strconv.Atoi(amountStr)
			if err != nil || amount <= 0 || amount > fixedDeposit.Amount {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("输入的金额无效，请输入有效的正整数"))
				return
			}

			if isMature {
				// 到期部分取款，计算对应利息
				interest := int(float64(amount) * fixedDeposit.Rate)
				totalAmount = amount + interest
				// 更新定期存款金额
				fixedDeposit.Amount -= amount
			} else {
				// 提前部分取款，扣除手续费(损失50%本金)
				penalty := amount / 2
				totalAmount = amount - penalty
				// 更新定期存款金额
				fixedDeposit.Amount -= amount
				ctx.SendChain(message.Text("提示：定期存款未到期提前取出，将扣除50%本金作为手续费"))
			}
		}

		// 如果是全部取出，则移除该定期存款；否则只更新金额
		if amountStr == "全部" {
			account.FixedDeposits = append(account.FixedDeposits[:index], account.FixedDeposits[index+1:]...)
		}
		
		account.LastUpdate = now
		if err := SaveAccount(account); err != nil {
			fmt.Printf("创建新账户保存失败: %v", err)
			return
		}

		// 增加到钱包
		err = wallet.InsertWalletOf(uid, totalAmount)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("取款失败：", err))
			return
		}

		if isMature {
			// 使用if-else替代三元运算符
			var principal, interest int
			if amountStr == "全部" {
				principal = fixedDeposit.Amount
			} else {
				principal = amount
			}
			interest = totalAmount - principal
			
			replyText := fmt.Sprintf(
				"定期存款到期取款成功！\n本金：%d%s\n利息：%d%s\n总计：%d%s",
				principal, wallet.GetWalletName(),
				interest, wallet.GetWalletName(),
				totalAmount, wallet.GetWalletName(),
			)
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(replyText))
		} else {
			// 使用if-else替代三元运算符
			var principal, penalty int
			if amountStr == "全部" {
				principal = fixedDeposit.Amount
			} else {
				principal = amount
			}
			penalty = principal - totalAmount
			
			replyText := fmt.Sprintf(
				"定期存款提前取款成功！\n本金：%d%s\n扣除手续费：%d%s\n实际到账：%d%s",
				principal, wallet.GetWalletName(),
				penalty, wallet.GetWalletName(),
				totalAmount, wallet.GetWalletName(),
			)
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(replyText))
		}
	})

	// 银行贷款命令处理（限制贷款天数最多10天）
	engine.OnPrefix("银行贷款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		params := strings.Fields(strings.TrimSpace(ctx.State["args"].(string)))
		if len(params) < 2 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("格式错误！示例：银行贷款 1000 7（贷款1000，期限7天，最多10天）"))
			return
		}

		amount, err := strconv.ParseFloat(params[0], 64)
		term, errTerm := strconv.Atoi(params[1])
		// 新增：限制期限必须为1-10天
		if err != nil || errTerm != nil || amount <= 0 || term <= 0 || term > 10 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("金额或期限无效！请输入正整数，且期限需在1-10天以内"))
			return
		}

		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)

		if HasActiveLoan(account) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你有未还清的贷款，无法再次贷款"))
			return
		}

		// 获取当前钱包余额
		walletBalance := wallet.GetWalletOf(uid)
		
		// 动态计算最大可贷金额（保持原有逻辑）
		var maxLoan int
		var limitReason string
		if walletBalance < BaseMaxLoanAmount {
			maxLoan = BaseMaxLoanAmount
			limitReason = fmt.Sprintf("你的钱包余额（%d%s）低于基础贷款额度。", 
				walletBalance, wallet.GetWalletName())
		} else {
			maxLoan = walletBalance * 10
			limitReason = fmt.Sprintf("你的钱包余额（%d%s）充足，可贷余额的10倍", 
				walletBalance, wallet.GetWalletName())
		}

		// 检查贷款金额是否超出上限
		if int(amount) > maxLoan {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(
				fmt.Sprintf("贷款金额超出上限！%s，当前可贷上限：%d%s",
					limitReason,
					maxLoan, wallet.GetWalletName(),
				),
			))
			return
		}

		// 计算利率（保留重生buff对利率的影响）
		stacks := getRebirthStacks(account)
		actualRate := float64(LoanDailyRate)
		if stacks > 0 {
			actualRate *= (1 + LoanRateIncreasePercent*float64(stacks))
		}

		// 创建贷款记录
		now := time.Now()
		loan := Loan{
			Amount:       amount,
			Rate:         actualRate,
			TermDays:     term,
			StartDate:    now,
			RepaidAmount: 0,
			IsOverdue:    false,
		}
		account.Loans = append(account.Loans, loan)
		account.LastUpdate = now
		_ = SaveAccount(account)

		// 发放贷款到钱包
		_ = wallet.InsertWalletOf(uid, int(amount))

		// 计算到期应还总额
		totalRepay := calculateLoanRepay(amount, actualRate, term)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(
			fmt.Sprintf("贷款成功！\n"+
				"金额：%.2f%s\n"+
				"期限：%d天（最长支持10天）\n"+  // 提示用户期限限制
				"日利率：%.2f%%%s\n"+
				"到期应还总额：%.2f%s\n"+
				"到期日：%s",
				amount, wallet.GetWalletName(),
				term,
				actualRate,
				func() string {
					if stacks > 0 {
						return fmt.Sprintf("（基础利率%.2f%% + 重生buff加成%.2f%%）",
							float64(LoanDailyRate),
							(actualRate-float64(LoanDailyRate)),
						)
					}
					return ""
				}(),
				totalRepay, wallet.GetWalletName(),
				now.AddDate(0, 0, term).Format("2006-01-02"),
			),
		))
	})

	// 银行还款（支持超额还款自动退款）
	engine.OnPrefix("银行还款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		param := strings.TrimSpace(ctx.State["args"].(string))
		amount, err := strconv.ParseFloat(param, 64)
		if err != nil || amount <= 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("还款金额无效，请输入正数字（支持小数）！"))
			return
		}

		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)
		if len(account.Loans) == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你暂无未还清的贷款！"))
			return
		}

		loan := &account.Loans[len(account.Loans)-1]
		totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
		remaining := totalRepay - loan.RepaidAmount

		if remaining <= 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("该笔贷款已还清，无需重复还款！"))
			return
		}

		walletBalance := float64(wallet.GetWalletOf(uid))
		if walletBalance < amount {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("钱包余额不足，无法还款！"))
			return
		}

		actualDeduct := amount
		if actualDeduct > remaining {
			actualDeduct = remaining
		}

		err = wallet.InsertWalletOf(uid, -int(math.Round(actualDeduct)))
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("还款失败：", err))
			return
		}

		loan.RepaidAmount += actualDeduct
		account.LastUpdate = time.Now()
		SaveAccount(account)

		if amount > remaining {
			refund := amount - remaining
			err = wallet.InsertWalletOf(uid, int(math.Round(refund)))
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("还款成功，但退款时发生错误：", err))
				return
			}
			
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf(
				"还款成功！该笔贷款已还清 ✅\n你实际支付了 %.2f%s，应还 %.2f%s，\n多付的 %.2f%s 已退还到你的钱包",
				amount, wallet.GetWalletName(), remaining, wallet.GetWalletName(),
				refund, wallet.GetWalletName(),
			)))
		} else {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf(
				"还款成功！\n已还 %.2f%s，剩余需还 %.2f%s",
				actualDeduct, wallet.GetWalletName(),
				totalRepay - loan.RepaidAmount, wallet.GetWalletName(),
			)))
		}
	})

	// 查看我的存款（整合活期和定期存款）
	engine.OnPrefix("查看我的存款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)
		
		var replyBuilder strings.Builder
		replyBuilder.WriteString("📊 你的存款信息如下：\n\n")
		replyBuilder.WriteString(fmt.Sprintf("💴 活期存款：%d%s\n\n", account.CurrentBalance, wallet.GetWalletName()))
		replyBuilder.WriteString("⏳ 定期存款：")
		
		if len(account.FixedDeposits) == 0 {
			replyBuilder.WriteString("暂无定期存款")
		} else {
			for i, deposit := range account.FixedDeposits {
				now := time.Now()
				isMature := now.After(deposit.MaturityDate) || now.Equal(deposit.MaturityDate)
				status := "未到期"
				if isMature {
					status = "已到期"
				}
				
				daysLeft := ""
				if !isMature {
					days := int(deposit.MaturityDate.Sub(now).Hours() / 24)
					daysLeft = fmt.Sprintf("，剩余%d天到期", days)
				}
				
				interest := int(float64(deposit.Amount) * deposit.Rate)
				total := deposit.Amount + interest
				
				replyBuilder.WriteString(fmt.Sprintf(
					"    \n%d.金额：%d%s\n期限：%d天\n状态：%s%s\n",
					i+1,
					deposit.Amount, wallet.GetWalletName(),
					deposit.TermDays,
					status,
					daysLeft,
				))
				replyBuilder.WriteString(fmt.Sprintf(
					"到期可获利息：%d%s\n到期本息总额：%d%s",
					interest, wallet.GetWalletName(),
					total, wallet.GetWalletName(),
				))
			}
		}
		
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(replyBuilder.String()))
	})

	// 查看我的贷款（只显示未还清的贷款）
	engine.OnFullMatch("查看我的贷款").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)
		if len(account.Loans) == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你暂无未还清的贷款！"))
			return
		}

		var activeLoans []Loan
		for _, loan := range account.Loans {
			totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
			if totalRepay - float64(loan.RepaidAmount) > 0 {
				activeLoans = append(activeLoans, loan)
			}
		}

		if len(activeLoans) == 0 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你暂无未还清的贷款！"))
			return
		}

		reply := message.Message{message.Reply(ctx.Event.MessageID), message.Text("你的未还清贷款信息：\n")}
		for _, loan := range activeLoans {
			totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
			interest := totalRepay - loan.Amount
			status := "正常"
			if loan.IsOverdue {
				status = "逾期 ❗"
			}
			maturityDateStr := loan.StartDate.AddDate(0, 0, loan.TermDays).Format("2006-01-02") 
			reply = append(reply, message.Text(fmt.Sprintf(
				"金额：%.2f%s\n利息：%.2f%s\n期限：%d天\n到期日：%s\n已还：%.2f%s\n剩余：%.2f%s\n状态：%s",
				loan.Amount, wallet.GetWalletName(),
				interest, wallet.GetWalletName(), 
				loan.TermDays,
				maturityDateStr,
				loan.RepaidAmount, wallet.GetWalletName(),
				totalRepay-float64(loan.RepaidAmount), wallet.GetWalletName(),
				status,
			)))
		}
		ctx.Send(reply)
	})

	// 查看今日利率
	engine.OnFullMatch("查看今日利率").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		checkAndUpdateInterest()

		reply := message.Message{message.Reply(ctx.Event.MessageID)}
		reply = append(reply, message.Text(fmt.Sprintf("今日活期存款利率：%.2f%%\n\n", dailyInterest.CurrentRate*100)))
		reply = append(reply, message.Text("定期存款利率："))
		
		var terms []int
		for term := range fixedDepositRates {
			terms = append(terms, term)
		}
		sort.Ints(terms)
		
		for _, term := range terms {
			rate := fixedDepositRates[term]
			reply = append(reply, message.Text(fmt.Sprintf("\n%d天：%.2f%%", term, rate*100)))
		}
		reply = append(reply, message.Text(fmt.Sprintf("\n\n贷款日利率：%.2f%%", float64(LoanDailyRate))))

		ctx.Send(reply)
	})

	// 信仰之跃功能（根据重生次数调整文本）
	engine.OnPrefix("我要信仰之跃").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		account := GetOrCreateAccount(uid)

		// 清理过期buff
		cleanupExpiredBuffs(account)

		// 检查当前重生buff层数
		currentStacks := getRebirthStacks(account)
		if currentStacks >= MaxRebirthStack {
			// 获取最早过期buff的剩余时间
			remainingTime, hasBuff := getEarliestRebirthBuffExpiry(account)
			
			timeInfo := "当前无有效重生buff"
			if hasBuff {
				if remainingTime <= 0 {
					timeInfo = "最早的重生buff已过期，即将自动清除"
				} else {
					timeInfo = fmt.Sprintf("最早的重生buff将在 %s 后过期，届时可再次进行信仰之跃", formatDuration(remainingTime))
				}
			}

			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(
				fmt.Sprintf("你的重生buff已达到最大层数%d层，无法继续进行信仰之跃！\n%s", 
					MaxRebirthStack, timeInfo),
			))
			return
		}

		// 检查是否符合无力偿还贷款的条件
		if !hasUnpayableLoan(account) {
			walletBalance := wallet.GetWalletOf(account.UserID)
			hasActiveDeposit := account.CurrentBalance > 0
			for _, fd := range account.FixedDeposits {
				if fd.Amount > 0 {
					hasActiveDeposit = true
					break
				}
			}
			
			hasUnpaidLoan := false
			for _, loan := range account.Loans {
				totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
				if totalRepay - float64(loan.RepaidAmount) > 0 {
					hasUnpaidLoan = true
					break
				}
			}
			
			reason := "当前未满足："
			if walletBalance != 0 {
				reason += fmt.Sprintf("钱包余额不为零（当前: %d%s）；", walletBalance, wallet.GetWalletName())
			}
			if hasActiveDeposit {
				reason += "银行存款（活期+定期）不为零；"
			}
			if !hasUnpaidLoan {
				reason += "没有未还清的贷款"
			}
			
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("你不符合信仰之跃的条件，需同时满足：\n1. 钱包余额为零\n2. 银行存款（活期+定期）为零\n3. 存在未还清的贷款\n"+reason))
			return
		}
		
		// 执行重生操作
		applyRebirth(account)
		
		// 添加重生buff
		addRebirthBuff(account)
		
		// 保存账户变更
		account.LastUpdate = time.Now()
		if err := SaveAccount(account); err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("信仰之跃失败，请稍后重试："+err.Error()))
			return
		}
		
		// 获取更新后的层数（新buff已添加，所以是当前实际层数）
		newStacks := getRebirthStacks(account)
		
		// 根据重生层数准备对应文本
		var text1, text2 string
		switch newStacks {
		case 1:
			// 第一次重生（层数为1）
			text1 = "上一世，你失败了，这一世，你要拿回属于你的一切.jpg"
			text2 = "但你似乎也发现了，银行对你带有些许的恶意"
		case 2:
			// 第二次重生（层数为2）
			text1 = "上一世...咦，你似乎发觉自己好像说过这句话了，好奇怪啊"
			text2 = "银行的恶意变深了...？你突然发觉自己似乎重生了不止一次"
		case 3:
			// 第三次重生（层数为3）
			text1 = "上...不对，这是我第几次重生了？"
			text2 = "银行的恶意达到了顶峰，这次重生，似乎已经抵达了某种上限。"
		// default:
		// 	// 超过3层时的默认文本（应对可能的扩展）
		// 	text1 = "轮回往复，你已经记不清这是第几次重生了"
		// 	text2 = "银行的恶意达到了顶峰，你感觉自己快要触碰到世界的真相..."
		}
		
		// 处理延迟消息
		msgID, ok := ctx.Event.MessageID.(int64)
		if !ok {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("消息ID解析失败"))
			return
		}
		go func(ctx *zero.Ctx, msgID int64) {
			ctx.SendChain(message.Reply(msgID), message.Text("正在为你生成场景天台...生成完毕。你一跃而下，成功完成了信仰之跃！"))
			time.Sleep(time.Second * 2)

			ctx.SendChain(message.Reply(msgID), message.Text(fmt.Sprintf(
				"你所有贷款已清零\n你获得了「重生」buff\n当前重生buff: %d/%d层（每层7天后自动消除）",
				newStacks, MaxRebirthStack,
			)))
			time.Sleep(time.Second * 2)

			ctx.SendChain(message.Reply(msgID), message.Text(text1))
			time.Sleep(time.Second * 2)

			ctx.SendChain(message.Reply(msgID), message.Text(text2))
		}(ctx, msgID)
	})
}