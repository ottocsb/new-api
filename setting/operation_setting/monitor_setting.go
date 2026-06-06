package operation_setting

import (
	"os"
	"strconv"
	"strings"

	"newapi/setting/config"
)

type MonitorSetting struct {
	AutoTestChannelEnabled bool    `json:"auto_test_channel_enabled"`
	AutoTestChannelMinutes float64 `json:"auto_test_channel_minutes"`
	// AutoTestChannelExcludeIds 为不参与定期/批量渠道测试的渠道 ID，逗号分隔。
	// 适用于附属生图等测试昂贵且大概率无法通过、但实际可用的渠道。
	AutoTestChannelExcludeIds string `json:"auto_test_channel_exclude_ids"`
}

// 默认配置
var monitorSetting = MonitorSetting{
	AutoTestChannelEnabled: false,
	AutoTestChannelMinutes: 10,
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("monitor_setting", &monitorSetting)
}

func GetMonitorSetting() *MonitorSetting {
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err == nil && frequency > 0 {
			monitorSetting.AutoTestChannelEnabled = true
			monitorSetting.AutoTestChannelMinutes = float64(frequency)
		}
	}
	return &monitorSetting
}

// GetAutoTestChannelExcludeIdSet 解析排除渠道 ID 列表为集合，忽略空白与非法项；
// 未配置时返回 nil（对 nil map 取值安全，调用方无需额外判空）。
func (s *MonitorSetting) GetAutoTestChannelExcludeIdSet() map[int]bool {
	if strings.TrimSpace(s.AutoTestChannelExcludeIds) == "" {
		return nil
	}
	set := make(map[int]bool)
	for _, part := range strings.Split(s.AutoTestChannelExcludeIds, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.Atoi(part); err == nil {
			set[id] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	return set
}
