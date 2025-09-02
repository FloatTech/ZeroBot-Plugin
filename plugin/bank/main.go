package bank

import (
    "time"

	zero "github.com/wdvxdr1123/ZeroBot"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

// 定期存款利率表(根据期限不同)
var fixedDepositRates = map[int]float64{
    3:  0.1,        // 3 天期， 利率 10%
    7:  0.36,       // 7 天期， 利率 36%
    15: 1,          // 15 天期，利率 100%
    30: 3,          // 30 天期，利率 300%
}

// 贷款日利率
const LoanDailyRate = 25 //日利率 25%

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
	RebirthBuffName = "rebirth"
	MaxRebirthStack = 3 // 最大叠加次数
    RebirthBuffDuration = 7 * 24 * time.Hour // 每层buff持续时间
)

var (
    // 注册插件引擎
    engine = control.Register("bank", &ctrl.Options[*zero.Ctx]{
        DisableOnDefault: false,
        Brief:            "银行",
        Help: "- 银行活期存款 [金额] (ps：活期利率每天都会在1%~3%之间波动)\n" +
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
    accounts      = make(map[int64]*BankAccount) // 内存中的账户数据
    dailyInterest = &InterestInfo{CurrentRate: 0.01} // 默认活期利率 1%
)