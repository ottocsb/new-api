package model

// InitColumnQuotingForTest 初始化跨数据库的保留字列名（`key`/`group` 等）。
// 生产路径由 InitDB 负责调用 initCol；外部包(如 service)的测试自行建库时，
// 需要显式调用本函数，否则拼接 SQL 会得到空列名而报语法错误。
func InitColumnQuotingForTest() {
	initCol()
}
