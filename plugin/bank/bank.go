// Package bank 银行
package bank

import (
	"os"
	"strconv"
	"time"
	"math/rand"
	"math"
	"fmt"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/AnimeAPI/wallet"
	"gopkg.in/yaml.v3"
)

func init() {
	// 初始化数据目录
	_ = os.MkdirAll(accountPath, 0755)
	
	// 加载账户数据
	loadAllAccounts()

	// 加载利息信息
	loadInterestInfo()

	// 启动定时任务
	initCronJobs()

	// 检查贷款是否逾期
    processLoanOverdue() 

	// 添加自动扣款任务
	autoDeductLoanPayment()
}

// 修改initCronJobs函数，添加自动扣款调用
func initCronJobs() {
	go func() {
		for {
			now := time.Now()
			nextDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(24 * time.Hour)
			time.Sleep(nextDay.Sub(now))

			// 每日任务
			calculateDailyCurrentInterest()
			processMatureFixedDeposits()
			processLoanOverdue()
			autoDeductLoanPayment() 	// 添加自动扣款任务
			cleanupAllExpiredBuffs() 	// 每日清理所有用户的过期buff
		}
	}()
}

// GetOrCreateAccount 获取或创建账户
func GetOrCreateAccount(uid int64) *Account {
	account, ok := accounts[uid]
	if !ok {
		account = &Account{
			UserID:         uid,
			CurrentBalance: 0,
			FixedDeposits:  []FixedDeposit{},
			Loans:          []Loan{},
			LastUpdate:     time.Now(),
		}
		accounts[uid] = account
	}
	return account
}

// 加载所有账户数据
func loadAllAccounts() {
	files, err := os.ReadDir(accountPath)
	if err != nil {
		return
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		uid, err := strconv.ParseInt(f.Name(), 10, 64)
		if err != nil {
			continue
		}
		data, err := os.ReadFile(accountPath + f.Name())
		if err != nil {
			continue
		}
		var account Account
		if err := yaml.Unmarshal(data, &account); err != nil {
			continue
		}
		accounts[uid] = &account
	}
}

// SaveAccount 保存账户信息
func SaveAccount(account *Account) error {
	data, err := yaml.Marshal(account)
	if err != nil {
		return err
	}
	return os.WriteFile(accountPath+strconv.FormatInt(account.UserID, 10), data, 0644)
}

// 检查并更新今日利率
func checkAndUpdateInterest() {
	today := time.Now().Format("2006-01-02")
	if dailyInterest.LastDate != today {
		rateRange := CurrentRateMax - CurrentRateMin // 计算利率范围
		// 每天随机生成活期利率（1% - 3%）
		dailyInterest.CurrentRate = CurrentRateMin + rand.Float64()*rateRange
		dailyInterest.LastDate = today
		data, _ := yaml.Marshal(dailyInterest)
		_ = os.WriteFile(interestPath, data, 0644)
	}
}

// 计算每日活期利息
func calculateDailyCurrentInterest() {
	now := time.Now()
	for _, account := range accounts {
		if account.CurrentBalance <= 0 {
			continue
		}
		
		// 基础利率
		rate := dailyInterest.CurrentRate
		
		// 重生用户活期利率下调
		stacks := getRebirthStacks(account)
		if stacks > 0 {
			rate *= (1 - CurrentRateDecreasePercent*float64(stacks))
		}
		
		interest := int(float64(account.CurrentBalance) * rate)
		if interest > 0 {
			account.CurrentBalance += interest
			account.LastUpdate = now
			_ = SaveAccount(account)
		}
	}
}

// 加载利息信息
func loadInterestInfo() {
    if !file.IsExist(interestPath) {
        // 检查 saveInterestInfo 的错误返回值
        if err := saveInterestInfo(); err != nil {
            fmt.Printf("加载利息信息失败：首次创建利息文件时出错 - %v", err)
            return
        }
        return
    }

    data, err := os.ReadFile(interestPath)
    if err != nil {
        fmt.Printf("读取利息信息文件失败：%v", err) // 补充错误日志
        return
    }

    // 检查 yaml 解析错误
    if err := yaml.Unmarshal(data, dailyInterest); err != nil {
        fmt.Printf("解析利息信息失败：%v", err)
    }
}

// 保存利息信息
func saveInterestInfo() error {
	data, err := yaml.Marshal(dailyInterest)
	if err != nil {
		return err
	}
	return os.WriteFile(interestPath, data, 0644)
}

// 处理到期的定期存款
func processMatureFixedDeposits() {
	now := time.Now()
	for _, account := range accounts {
		matureIndex := -1
		for i, fd := range account.FixedDeposits {
			if now.After(fd.MaturityDate) || now.Equal(fd.MaturityDate) {
				matureIndex = i
				break
			}
		}
		if matureIndex != -1 {
			// 取出到期存款及利息
			fd := account.FixedDeposits[matureIndex]
			interest := int(float64(fd.Amount) * fd.Rate / 100)
			account.CurrentBalance += fd.Amount + interest
			// 移除到期存款
			account.FixedDeposits = append(account.FixedDeposits[:matureIndex], account.FixedDeposits[matureIndex+1:]...)
			account.LastUpdate = now
			_ = SaveAccount(account)
		}
	}
}

// CalculateTotalCurrentDeposits 计算银行总活期存款
func CalculateTotalCurrentDeposits() (total int) {
    for _, account := range accounts {
        total += account.CurrentBalance
    }
    return
}

// CalculateTotalFixedDeposits 计算银行总定期存款
func CalculateTotalFixedDeposits() (total int) {
    for _, account := range accounts {
        for _, fd := range account.FixedDeposits {
            total += fd.Amount
        }
    }
    return
}

// CalculateTotalLoans 计算银行总贷款（未还清的贷款总额）
func CalculateTotalLoans() (total float64) {
    for _, account := range accounts {
        for _, loan := range account.Loans {
            totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
            // 累加未还清部分
            if loan.RepaidAmount < totalRepay {
                total += totalRepay - loan.RepaidAmount
            }
        }
    }
    return
}

// HasActiveLoan 检查是否有未还清的贷款
func HasActiveLoan(account *Account) bool {
	for _, loan := range account.Loans {
		if loan.RepaidAmount < loan.Amount {
			return true
		}
	}
	return false
}

// 计算贷款应还总额
func calculateLoanRepay(amount, rate float64, termDays int) float64 {
	dailyRate := rate / 100
	interest := amount * dailyRate * float64(termDays)
	return amount + interest
}

// 处理逾期贷款
func autoDeductLoanPayment() {
	now := time.Now()
	for _, account := range accounts {
		// 只收集逾期且未还清的贷款
		var overdueLoans []int  // 仅处理逾期贷款
		
		// 筛选条件：必须是逾期状态（IsOverdue=true）且有未还清金额
		for i, loan := range account.Loans {
			totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
			remaining := totalRepay - float64(loan.RepaidAmount)
			
			// 仅处理：逾期 + 未还清的贷款
			if loan.IsOverdue && remaining > 0 {
				overdueLoans = append(overdueLoans, i)
			}
			// 非逾期贷款：直接跳过，不做任何处理
		}
		
		// 只对逾期且未还清的贷款执行自动扣款
		for _, i := range overdueLoans {
			deductOverdueLoan(account, i, now)
		}
	}
}

// 处理贷款逾期
func processLoanOverdue() {
	now := time.Now()
	for _, account := range accounts {
		for i := range account.Loans {
			loan := &account.Loans[i]
			// 计算到期日
			maturityDate := loan.StartDate.AddDate(0, 0, loan.TermDays)
			// 检查是否逾期
			if now.After(maturityDate) && !loan.IsOverdue {
				loan.IsOverdue = true
				account.LastUpdate = now
				_ = SaveAccount(account)
			}
		}
	}
}

// 专门处理逾期贷款的扣款逻辑
func deductOverdueLoan(account *Account, loanIndex int, now time.Time) {
	loan := &account.Loans[loanIndex]
	totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
	remaining := totalRepay - float64(loan.RepaidAmount)
	
	// 安全校验：确保是逾期且未还清状态
	if !loan.IsOverdue || remaining <= 0 {
		return
	}
	
	needDeduct := remaining
	fmt.Printf("[逾期自动扣款] 用户%d的逾期贷款开始扣款，剩余应还: %.2f", account.UserID, needDeduct)
	
	// 1. 从钱包扣款
	walletBalance := float64(wallet.GetWalletOf(account.UserID))
	if walletBalance > 0 {
		deduct := math.Min(needDeduct, walletBalance)
		_ = wallet.InsertWalletOf(account.UserID, -int(math.Round(deduct)))
		loan.RepaidAmount += float64(int(math.Round(deduct)))
		needDeduct -= deduct
		fmt.Printf("[逾期自动扣款] 从钱包扣除: %.2f，剩余: %.2f", deduct, needDeduct)
	}
	
	if needDeduct <= 0 {
		account.LastUpdate = now
		_ = SaveAccount(account)
		fmt.Printf("[逾期自动扣款] 用户%d的逾期贷款已还清", account.UserID)
		cleanupCompletedLoans(account) // 清理已还清的逾期贷款记录
		return
	}
	
	// 2. 从活期存款扣款
	if account.CurrentBalance > 0 {
		deduct := math.Min(needDeduct, float64(account.CurrentBalance))
		account.CurrentBalance -= int(math.Round(deduct))
		loan.RepaidAmount += deduct
		needDeduct -= deduct
		fmt.Printf("[逾期自动扣款] 从活期扣除: %.2f，剩余: %.2f", deduct, needDeduct)
	}
	
	if needDeduct <= 0 {
		account.LastUpdate = now
		_ = SaveAccount(account)
		fmt.Printf("[逾期自动扣款] 用户%d的逾期贷款已还清", account.UserID)
		cleanupCompletedLoans(account)
		return
	}
	
	// 3. 从定期存款扣款（逾期贷款手续费70%）
	feeRate := 0.7 // 固定为逾期的高手续费率
	for len(account.FixedDeposits) > 0 && needDeduct > 0 {
		fd := &account.FixedDeposits[0]
		available := float64(fd.Amount) * (1 - feeRate) // 扣除手续费后的可用金额
		
		deduct := math.Min(needDeduct, available)
		loan.RepaidAmount += deduct
		needDeduct -= deduct
		
		// 计算实际消耗的定期本金（含手续费）
		usedPrincipal := int(math.Ceil(deduct / (1 - feeRate)))
		fmt.Printf("[逾期自动扣款] 从定期扣除: %.2f（本金消耗: %d，手续费率: 70%%）", deduct, usedPrincipal)
		
		// 更新定期存款记录
		if usedPrincipal >= fd.Amount {
			account.FixedDeposits = account.FixedDeposits[1:] // 全额取出
		} else {
			fd.Amount -= usedPrincipal // 部分取出
		}
	}
	
	// 最终处理
	account.LastUpdate = now
	if err := SaveAccount(account); err != nil {
		fmt.Printf("[逾期自动扣款] 保存用户%d账户失败: %v", account.UserID, err)
	} else if needDeduct > 0 {
		fmt.Printf("[逾期自动扣款] 用户%d的逾期贷款仍有剩余: %.2f", account.UserID, needDeduct)
	} else {
		fmt.Printf("[逾期自动扣款] 用户%d的逾期贷款已还清", account.UserID)
		cleanupCompletedLoans(account)
	}
}


// 清理账户中已完成的贷款记录
func cleanupCompletedLoans(account *Account) {
	filteredLoans := []Loan{}
	for _, loan := range account.Loans {
		// 保留未还清的贷款（已还金额 < 应还总额）
		totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
		if loan.RepaidAmount < totalRepay {
			filteredLoans = append(filteredLoans, loan)
		}
	}
	// 更新贷款列表（删除已完成记录）
	account.Loans = filteredLoans
}

// 检查用户是否无力偿还贷款（钱包为零+银行存款为零+有未还清贷款）
func hasUnpayableLoan(account *Account) bool {
    // 钱包余额判断
    walletBalance := wallet.GetWalletOf(account.UserID)
    if walletBalance != 0 {
        return false
    }

    // 银行存款判断（活期 + 定期都要为 0）
    if account.CurrentBalance != 0 {
        return false
    }
    for _, fd := range account.FixedDeposits {
        if fd.Amount > 0 {
            return false
        }
    }

    // 未还清贷款判断
    for _, loan := range account.Loans {
        totalRepay := calculateLoanRepay(loan.Amount, loan.Rate, loan.TermDays)
        if totalRepay - float64(loan.RepaidAmount) > 0 {
            return true // 存在未还清贷款
        }
    }
    return false // 无未还清贷款
}

// 执行信仰之跃重生
func applyRebirth(account *Account) {
    // 1. 清零所有贷款
    account.Loans = []Loan{}
    
    // 2. （可选）清理已完成的贷款（如果有）
    cleanupCompletedLoans(account)
    
    // 3. 不添加任何 buff！（buff 由 addRebirthBuff 单独处理）
}

// 获取重生buff的有效层数
func getRebirthStacks(account *Account) int {
	now := time.Now()
	stacks := 0
	for _, buff := range account.Buffs {
		// 直接通过 Buff.Name 判断，无需额外转换
		if buff.Name == RebirthBuffName && now.Before(buff.ExpireTime) {
			stacks++
		}
	}
	// 限制最大层数
	if stacks > MaxRebirthStack {
		return MaxRebirthStack
	}
	return stacks
}

// 向账户添加重生buff（自动处理叠加和过期）
func addRebirthBuff(account *Account) {
	now := time.Now()
	// 先清理过期buff，避免无效数据干扰
	cleanupExpiredBuffs(account)

	currentStacks := getRebirthStacks(account)
	newBuff := Buff{
		Name:       RebirthBuffName,
		ExpireTime: now.Add(RebirthBuffDuration), // 7天后过期
	}

	if currentStacks < MaxRebirthStack {
		// 未达上限，直接添加新buff
		account.Buffs = append(account.Buffs, newBuff)
	} else {
		// 已达上限，替换最早过期的buff
		earliestIndex := 0
		earliestTime := account.Buffs[0].ExpireTime
		for i, buff := range account.Buffs {
			if buff.Name == RebirthBuffName && buff.ExpireTime.Before(earliestTime) {
				earliestIndex = i
				earliestTime = buff.ExpireTime
			}
		}
		account.Buffs[earliestIndex] = newBuff
	}
}

// 清理账户中所有过期的重生buff
func cleanupExpiredBuffs(account *Account) {
	now := time.Now()
	validBuffs := []Buff{}
	for _, buff := range account.Buffs {
		if now.Before(buff.ExpireTime) { // 保留未过期的buff
			validBuffs = append(validBuffs, buff)
		}
	}
	account.Buffs = validBuffs
}

// 清理所有用户的过期buff
func cleanupAllExpiredBuffs() {
	now := time.Now()
	for _, account := range accounts {
		originalCount := len(account.Buffs)
		cleanupExpiredBuffs(account)
		// 若buff数量有变化，保存更新
		if len(account.Buffs) != originalCount {
			account.LastUpdate = now
			_ = SaveAccount(account)
		}
	}
}

// 计算最早过期的重生buff剩余时间
func getEarliestRebirthBuffExpiry(account *Account) (time.Duration, bool) {
    now := time.Now()
    var earliestExpiry time.Time

    // 遍历所有buff，筛选出重生buff（通过 Buff.Name 判断）
    for _, buff := range account.Buffs {
        if buff.Name == RebirthBuffName { // 直接使用 Buff 结构体的 Name 字段
            // 记录最早过期的重生buff
            if earliestExpiry.IsZero() || buff.ExpireTime.Before(earliestExpiry) {
                earliestExpiry = buff.ExpireTime // 直接使用 Buff 结构体的 ExpireTime 字段
            }
        }
    }

    if earliestExpiry.IsZero() {
        return 0, false // 无有效重生buff
    }

    // 计算剩余时间（若已过期则返回0）
    if earliestExpiry.Before(now) {
        return 0, true
    }
    return earliestExpiry.Sub(now), true
}

// 格式化时间间隔为"X小时Y分钟"
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "已过期"
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%d小时%d分钟", hours, minutes)
}