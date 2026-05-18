package service

import (
	"strings"

	"newapi/common"
	"newapi/setting/system_setting"
)

func PaymentReturnURL(suffix string) string {
	base := strings.TrimRight(system_setting.ServerAddress, "/")
	return base + common.ThemeAwarePath(suffix)
}
