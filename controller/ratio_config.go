package controller

import (
	"net/http"

	"newapi/common"
	"newapi/setting/billing_setting"
	"newapi/setting/ratio_setting"

	"github.com/gin-gonic/gin"
)

func GetRatioConfig(c *gin.Context) {
	if !ratio_setting.IsExposeRatioEnabled() {
		common.ApiErrorStatus(c, http.StatusForbidden, common.NewMessage("The ratio config API is not enabled"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    billing_setting.GetPricingSyncData(map[string]any(ratio_setting.GetExposedData())),
	})
}
