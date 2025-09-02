package bank

import (
	"os"
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

    processLoanOverdue() 

	// 添加自动扣款任务
	autoDeductLoanPayment()
}
