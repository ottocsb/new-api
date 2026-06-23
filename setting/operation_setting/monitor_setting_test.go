package operation_setting

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAutoTestChannelExcludeIdSet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[int]bool
	}{
		{"empty", "", nil},
		{"blank", "   ", nil},
		{"all invalid", "abc, ,xyz", nil},
		{"single", "12", map[int]bool{12: true}},
		{"multiple", "12,34,56", map[int]bool{12: true, 34: true, 56: true}},
		{"spaces and empty parts", " 12 , 34 ,, 56 ", map[int]bool{12: true, 34: true, 56: true}},
		{"invalid ignored", "12,abc,34", map[int]bool{12: true, 34: true}},
		{"dedup", "12,12,34", map[int]bool{12: true, 34: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MonitorSetting{AutoTestChannelExcludeIds: tt.input}
			got := s.GetAutoTestChannelExcludeIdSet()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAutoTestChannelExcludeIdSet(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetMonitorSetting_ChannelTestEnabledEnvOverridesEnabledConfig(t *testing.T) {
	orig := monitorSetting
	t.Cleanup(func() { monitorSetting = orig })

	t.Setenv("CHANNEL_TEST_ENABLED", "false")
	t.Setenv("CHANNEL_TEST_FREQUENCY", "5")
	monitorSetting = MonitorSetting{
		AutoTestChannelEnabled: true,
		AutoTestChannelMinutes: 20,
	}

	setting := GetMonitorSetting()

	require.NotNil(t, setting)
	assert.False(t, setting.AutoTestChannelEnabled)
	assert.Equal(t, float64(5), setting.AutoTestChannelMinutes)
}

func TestGetMonitorSetting_ChannelTestEnabledEnvCanEnableDisabledConfig(t *testing.T) {
	orig := monitorSetting
	t.Cleanup(func() { monitorSetting = orig })

	t.Setenv("CHANNEL_TEST_ENABLED", "true")
	monitorSetting = MonitorSetting{
		AutoTestChannelEnabled: false,
		AutoTestChannelMinutes: 12,
	}

	setting := GetMonitorSetting()

	require.NotNil(t, setting)
	assert.True(t, setting.AutoTestChannelEnabled)
	assert.Equal(t, float64(12), setting.AutoTestChannelMinutes)
}
