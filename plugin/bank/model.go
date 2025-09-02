package bank

import "time"

type Buff struct {
	Name       string    `yaml:"name"`       	// buff名称
	ExpireTime time.Time `yaml:"expire_time"` 	// 过期时间
}

// FixedDeposit 定期存款信息
type FixedDeposit struct {
	Amount       int       `yaml:"amount"`       // 存款金额
	TermDays     int       `yaml:"term_days"`    // 存款期限(天)
	Rate         float64   `yaml:"rate"`         // 存款利率
	StartDate    time.Time `yaml:"start_date"`   // 开始日期
	MaturityDate time.Time `yaml:"maturity_date"` // 到期日期
}

// Loan 贷款信息
type Loan struct {
	Amount       float64   `yaml:"amount"`        // 贷款金额
	Rate         float64   `yaml:"rate"`          // 贷款日利率
	TermDays     int       `yaml:"term_days"`     // 贷款期限(天)
	StartDate    time.Time `yaml:"start_date"`    // 贷款发放日期
	RepaidAmount float64   `yaml:"repaid_amount"` // 已还金额
	IsOverdue    bool      `yaml:"is_overdue"`    // 是否逾期
}

// BankAccount 银行账户信息
type BankAccount struct {
	UserID         int64          `yaml:"user_id"`         // 用户ID
	CurrentBalance int            `yaml:"current_balance"` // 活期余额
	FixedDeposits  []FixedDeposit `yaml:"fixed_deposits"`  // 定期存款列表
	Loans          []Loan         `yaml:"loans"`           // 贷款列表
	LastUpdate     time.Time      `yaml:"last_update"`     // 最后更新时间
	Buffs          []Buff       `yaml:"buffs"`           	// 拥有的buff
}

// InterestInfo 利息信息
type InterestInfo struct {
	CurrentRate float64 `yaml:"current_rate"` // 今日活期利率
	LastDate    string  `yaml:"last_date"`    // 最后更新日期
}

