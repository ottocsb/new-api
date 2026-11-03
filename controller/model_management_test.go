package controller

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"newapi/common"
	"newapi/constant"
	"newapi/model"
	"newapi/pkg/billingexpr"
	"newapi/service"
	"newapi/setting/billing_setting"
	"newapi/setting/config"
	"newapi/setting/ratio_setting"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func modelManagementDB(t *testing.T, kind, dsn string) *gorm.DB {
	t.Helper()
	database, isolatedDSN := newAuditTestDatabase(t, kind, dsn)
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMain, previousLog := common.MainDatabaseType(), common.LogDatabaseType()
	previousMaster, previousSQLite := common.IsMasterNode, common.SQLitePath
	previousRedis, previousMemory := common.RedisEnabled, common.MemoryCacheEnabled
	previousOptions := common.OptionMap
	previousConfig := config.GlobalConfig.ExportAllConfigs()
	restoreRatios := []struct {
		value   string
		restore func(string) error
	}{
		{ratio_setting.ModelPrice2JSONString(), ratio_setting.UpdateModelPriceByJSONString},
		{ratio_setting.ModelRatio2JSONString(), ratio_setting.UpdateModelRatioByJSONString},
		{ratio_setting.CompletionRatio2JSONString(), ratio_setting.UpdateCompletionRatioByJSONString},
		{ratio_setting.CacheRatio2JSONString(), ratio_setting.UpdateCacheRatioByJSONString},
		{ratio_setting.CreateCacheRatio2JSONString(), ratio_setting.UpdateCreateCacheRatioByJSONString},
		{ratio_setting.ImageRatio2JSONString(), ratio_setting.UpdateImageRatioByJSONString},
		{ratio_setting.AudioRatio2JSONString(), ratio_setting.UpdateAudioRatioByJSONString},
		{ratio_setting.AudioCompletionRatio2JSONString(), ratio_setting.UpdateAudioCompletionRatioByJSONString},
	}
	common.IsMasterNode = false
	common.RedisEnabled, common.MemoryCacheEnabled = false, false
	common.OptionMap = map[string]string{}
	if kind == "sqlite" {
		common.SQLitePath = isolatedDSN
		isolatedDSN = "local"
	}
	t.Setenv("SQL_DSN", isolatedDSN)
	t.Setenv("LOG_SQL_DSN", "")
	require.NoError(t, model.InitDB())
	database = model.DB
	model.LOG_DB = database
	require.NoError(t, database.AutoMigrate(&model.Model{}, &model.Vendor{}, &model.Channel{}, &model.Ability{}, &model.Option{}, &model.User{}, &model.AuditLog{}))
	for _, value := range restoreRatios {
		require.NoError(t, value.restore("{}"))
	}
	config.UpdateConfigFromMap(config.GlobalConfig.Get("billing_setting"), map[string]string{"billing_mode": "{}", "billing_expr": "{}", "plugin_billing_expr": "{}"})
	var version string
	query := "SELECT version()"
	if kind == "sqlite" {
		query = "SELECT sqlite_version()"
	}
	require.NoError(t, database.Raw(query).Scan(&version).Error)
	t.Logf("database version: %s", version)
	t.Cleanup(func() {
		for _, value := range restoreRatios {
			require.NoError(t, value.restore(value.value))
		}
		config.UpdateConfigFromMap(config.GlobalConfig.Get("billing_setting"), map[string]string{"billing_mode": previousConfig["billing_setting.billing_mode"], "billing_expr": previousConfig["billing_setting.billing_expr"], "plugin_billing_expr": previousConfig[billing_setting.PluginBillingExprOption]})
		common.OptionMap = previousOptions
		common.IsMasterNode, common.SQLitePath = previousMaster, previousSQLite
		common.RedisEnabled, common.MemoryCacheEnabled = previousRedis, previousMemory
		common.SetDatabaseTypes(previousMain, previousLog)
		connection, err := database.DB()
		if err == nil {
			require.NoError(t, connection.Close())
		}
		model.DB, model.LOG_DB = previousDB, previousLogDB
	})
	return database
}

func modelManagementRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body any, output any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := common.Marshal(body)
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, path, bytes.NewReader(encoded))
	context.Set("role", common.RoleRootUser)
	handler(context)
	if output != nil {
		require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), output), recorder.Body.String())
	}
	return recorder
}

func TestModelPricingConversionDatabaseMatrix(t *testing.T) {
	previousQuota := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = previousQuota })
	for _, dialect := range []struct{ kind, env string }{{"sqlite", ""}, {"mysql", "TEST_MYSQL_DSN"}, {"postgres", "TEST_POSTGRES_DSN"}} {
		t.Run(dialect.kind, func(t *testing.T) {
			if dialect.env != "" && os.Getenv(dialect.env) == "" {
				t.Skip("set " + dialect.env + " to run this database")
			}
			db := modelManagementDB(t, dialect.kind, os.Getenv(dialect.env))
			contract, err := os.ReadFile("../pkg/billingexpr/testdata/frontend_simulation.json")
			require.NoError(t, err)
			var fixtures []struct {
				Name, Expression string
				Conversion       *struct {
					ModelName string              `json:"model_name"`
					Pricing   model.PricingValues `json:"pricing"`
				}
			}
			require.NoError(t, common.Unmarshal(contract, &fixtures))
			for _, fixture := range fixtures {
				if fixture.Conversion == nil {
					continue
				}
				t.Run("frontend_contract/"+fixture.Name, func(t *testing.T) {
					var response struct {
						Success bool
						Data    model.ModelPricingConversion
					}
					modelManagementRequest(t, PreviewModelPricingConversion, http.MethodPost, "/api/option/model_pricing/convert", fixture.Conversion, &response)
					require.True(t, response.Success)
					assert.Equal(t, fixture.Expression, response.Data.Expression)
				})
			}
			// The preview must resolve the draft, even when the running process
			// still has a different saved output multiplier for that model.
			require.NoError(t, ratio_setting.UpdateCompletionRatioByJSONString(`{"conversion-defaults":99}`))
			for _, tc := range []struct {
				name               string
				draft              model.PricingValues
				expression, reason string
			}{
				{"conversion-defaults", model.PricingValues{"ModelRatio": float64(2)}, `tier("base", p * 4 + c * 4)`, ""},
				{"gpt-4-32k", model.PricingValues{"ModelRatio": 30.0, "CacheRatio": 1.0}, `tier("base", p * 60 + c * 120)`, ""},
				{"conversion-free-cache", model.PricingValues{"ModelRatio": 0.0, "CacheRatio": 0.0}, `tier("base", p * 0 + c * 0 + cr * 0)`, ""},
				{"claude-3-7-sonnet-20250219", model.PricingValues{"ModelRatio": float64(1.5), "CacheRatio": float64(0.1), "CreateCacheRatio": float64(1.25)}, `tier("base", p * 3 + c * 15 + cr * 0.3 + cc * 3.75 + cc1h * 6)`, ""},
				{"conversion-image-default", model.PricingValues{"ModelRatio": float64(2), "ImageRatio": float64(1)}, `tier("base", p * 4 + c * 4)`, ""},
				{"conversion-image-free", model.PricingValues{"ModelRatio": float64(2), "ImageRatio": float64(0)}, `tier("base", p * 4 + c * 4 + cr * 4 + img * 0)`, ""},
				{"deepseek-chat", model.PricingValues{"ModelRatio": float64(0.135), "CacheRatio": float64(0.25)}, `tier("base", p * 0.27 + c * 0.27 + cr * 0.0675)`, ""},
				{"gpt-4o-custom", model.PricingValues{"ModelRatio": float64(2)}, `tier("base", p * 4 + c * 16)`, ""},
				{"gemini-2.5-pro-custom", model.PricingValues{"ModelRatio": float64(2)}, `tier("base", p * 4 + c * 32)`, ""},
				{"vendor/claude-sonnet-4", model.PricingValues{"ModelRatio": float64(2), "CompletionRatio": float64(0)}, `tier("base", p * 4 + c * 0 + cr * 4 + cc * 5 + cc1h * 8)`, ""},
				{"conversion-custom", model.PricingValues{"ModelRatio": float64(2), "CompletionRatio": float64(3), "CacheRatio": float64(0), "CreateCacheRatio": float64(1.5), "ImageRatio": float64(2)}, `tier("base", p * 4 + c * 12 + cr * 0 + cc * 6 + img * 8)`, ""},
				{"deepseek-reasoner", model.PricingValues{"ModelRatio": 0.275, "CacheRatio": 0.25}, `tier("base", p * 0.55 + c * 0.55 + cr * 0.1375)`, ""},
				{"gpt-5.6-sol", model.PricingValues{"ModelRatio": float64(2), "CompletionRatio": float64(2)}, `tier("base", p * 4 + c * 8)`, ""},
				{"gpt-5.5", model.PricingValues{"ModelRatio": float64(2), "CompletionRatio": float64(2)}, `tier("base", p * 4 + c * 8)`, ""},
				{"gpt-6-astra", model.PricingValues{"ModelRatio": float64(2), "CompletionRatio": float64(2)}, `tier("base", p * 4 + c * 8)`, ""},
				{"conversion-free", model.PricingValues{"ModelPrice": float64(0)}, `tier("request", fixed(0))`, ""},
				{"conversion-fixed", model.PricingValues{"ModelPrice": float64(0.25), "ModelRatio": float64(7)}, `tier("request", fixed(0.25))`, ""},
				{"gpt-4o-2024-05-13", model.PricingValues{"ModelRatio": float64(2), "CompletionRatio": float64(99)}, `tier("base", p * 4 + c * 12)`, ""},
				{"gpt-image-2", model.PricingValues{"ModelPrice": float64(1)}, `tier("image", fixed(1)) * image_count`, ""},
				// Bailian image models are served by the alibaba task plugin, so their pricing is a task usage expression.
				{"video-summarizer", model.PricingValues{"ModelPrice": float64(1)}, `tier("request", fixed(1))`, ""},
				{"sora-transcript-helper", model.PricingValues{"ModelPrice": float64(1)}, `tier("request", fixed(1))`, ""},
				{"gpt-realtime", model.PricingValues{"ModelRatio": float64(1)}, "", "Realtime pricing must be converted manually."},
				{"conversion-audio", model.PricingValues{"ModelRatio": float64(1), "AudioRatio": float64(2)}, `tier("base", p * 2 + c * 2 + ai * 4 + ao * 4)`, ""},
			} {
				t.Run(tc.name, func(t *testing.T) {
					before, err := model.GetModelPricingSnapshot([]string{tc.name})
					require.NoError(t, err)
					var response struct {
						Success bool
						Data    model.ModelPricingConversion
					}
					modelManagementRequest(t, PreviewModelPricingConversion, http.MethodPost, "/api/option/model_pricing/convert", map[string]any{"model_name": tc.name, "pricing": tc.draft}, &response)
					require.True(t, response.Success)
					assert.Equal(t, tc.expression, response.Data.Expression)
					assert.Equal(t, tc.reason, response.Data.UnsupportedReason)
					if tc.name == "claude-3-7-sonnet-20250219" {
						// Both providers include images in input; Claude reports cache
						// separately and does not report an image token breakdown.
						for _, usage := range []dto.Usage{
							{PromptTokens: 1000, CompletionTokens: 100, UsageSemantic: "openai", PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 200, CachedCreationTokens: 50, ImageTokens: 300}},
							{PromptTokens: 750, CompletionTokens: 100, UsageSemantic: "anthropic", ClaudeCacheCreation5mTokens: 50, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 200, CachedCreationTokens: 50}},
						} {
							tokens := service.BuildTieredTokenParams(&usage, usage.UsageSemantic == "anthropic", billingexpr.UsedVars(response.Data.Expression))
							cost, _, err := billingexpr.RunExpr(response.Data.Expression, tokens)
							require.NoError(t, err)
							assert.Equal(t, float64(750), tokens.P)
							assert.Equal(t, 3997.5, cost, "image input must remain billed exactly once")
						}
						usage := dto.Usage{PromptTokens: 750, CompletionTokens: 100, UsageSemantic: "anthropic", ClaudeCacheCreation5mTokens: 30, ClaudeCacheCreation1hTokens: 20, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 200, CachedCreationTokens: 50}}
						tokens := service.BuildTieredTokenParams(&usage, true, billingexpr.UsedVars(response.Data.Expression))
						cost, _, err := billingexpr.RunExpr(response.Data.Expression, tokens)
						require.NoError(t, err)
						assert.Equal(t, 4042.5, cost, "mixed TTLs preserve both legacy cache-write prices")
					}
					if tc.name == "deepseek-chat" {
						for _, usage := range []dto.Usage{
							{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 200}},
							{PromptTokens: 800, CompletionTokens: 100, UsageSemantic: "anthropic", PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 200}},
						} {
							tokens := service.BuildTieredTokenParams(&usage, usage.UsageSemantic == "anthropic", billingexpr.UsedVars(response.Data.Expression))
							cost, _, err := billingexpr.RunExpr(response.Data.Expression, tokens)
							require.NoError(t, err)
							assert.Equal(t, 256.5, cost, "cache reads remain billed across both response formats")
						}
					}
					after, err := model.GetModelPricingSnapshot([]string{tc.name})
					require.NoError(t, err)
					assert.Equal(t, before.Entries, after.Entries, "preview must not write pricing")
					if tc.expression == "" {
						return
					}
					pricing := tc.draft
					pricing["billing_setting.billing_mode"] = "tiered_expr"
					pricing["billing_setting.billing_expr"] = tc.expression
					change := model.ModelPricingChange{ModelName: tc.name, ExpectedVersion: before.Entries[0].Version, Pricing: pricing}
					require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{change}))
					after, err = model.GetModelPricingSnapshot([]string{tc.name})
					require.NoError(t, err)
					assert.Equal(t, pricing, after.Entries[0].Configured)
					assert.ErrorIs(t, model.UpdateModelPricing([]model.ModelPricingChange{change}), model.ErrModelPricingConflict)
					change.ExpectedVersion = after.Entries[0].Version
					change.Pricing["billing_setting.billing_mode"] = "ratio"
					require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{change}))
					after, err = model.GetModelPricingSnapshot([]string{tc.name})
					require.NoError(t, err)
					assert.Equal(t, "ratio", after.Entries[0].Effective["billing_setting.billing_mode"])
					if _, fixed := pricing["ModelPrice"]; !fixed {
						var preview struct {
							Success bool
							Data    model.ModelPricingDescription
						}
						modelManagementRequest(t, PreviewModelPricing, http.MethodPost, "/api/option/model_pricing/preview", map[string]any{"model_name": tc.name, "pricing": pricing}, &preview)
						require.True(t, preview.Success)
						assert.Equal(t, after.Entries[0].Effective, preview.Data.Effective)
						require.NotEmpty(t, response.Data.Effective)
						for field, value := range response.Data.Effective {
							assert.Equal(t, value, preview.Data.Effective[field], "conversion preview field %s", field)
						}
						cache, _ := ratio_setting.GetCacheRatio(tc.name)
						createCache, _ := ratio_setting.GetCreateCacheRatio(tc.name)
						image, _ := ratio_setting.GetImageRatio(tc.name)
						assert.Equal(t, ratio_setting.GetCompletionRatio(tc.name), preview.Data.Effective["CompletionRatio"])
						assert.Equal(t, cache, preview.Data.Effective["CacheRatio"])
						assert.Equal(t, createCache, preview.Data.Effective["CreateCacheRatio"])
						assert.Equal(t, image, preview.Data.Effective["ImageRatio"])
					}
				})
			}
			require.NoError(t, (&model.Channel{Name: "Mapped image", Models: "conversion-alias", Type: 1, ModelMapping: common.GetPointer(`{"conversion-alias":"conversion-hop","conversion-hop":"gpt-image-2"}`)}).Insert())
			preview, err := model.PreviewModelPricingConversion("conversion-alias", model.PricingValues{"ModelPrice": float64(1)})
			require.NoError(t, err)
			assert.Equal(t, `tier("image", fixed(1)) * image_count`, preview.Expression)
			assert.Empty(t, preview.UnsupportedReason)

			t.Run("cache_writes_follow_configured_prices_and_claude_names", func(t *testing.T) {
				for _, tc := range []struct {
					name       string
					configured *float64
					mode       model.CacheWriteMode
				}{
					{"unrecognized-model", nil, model.CacheWriteNone},
					{"gpt-5.6-terra", nil, model.CacheWriteNone},
					{"gpt-5.6-terra", common.GetPointer(1.25), model.CacheWriteOpenAITTL},
					{"custom-GPT-alias", common.GetPointer(0.0), model.CacheWriteOpenAITTL},
					{"vendor/claude-sonnet-4-6-high", nil, model.CacheWriteClaudeTTL},
					{"custom-CLAUDE-alias", common.GetPointer(0.0), model.CacheWriteClaudeTTL},
					{"deepseek-reasoner", common.GetPointer(0.0), model.CacheWriteStandard},
					{"custom-cache", common.GetPointer(1.5), model.CacheWriteStandard},
				} {
					t.Run(tc.name+"/"+string(tc.mode), func(t *testing.T) {
						draft := model.PricingValues{"ModelRatio": 2.0, "CompletionRatio": 2.0, "CacheRatio": 0.25, "billing_setting.billing_mode": "ratio"}
						if tc.configured != nil {
							draft["CreateCacheRatio"] = *tc.configured
						}
						var response struct {
							Success bool
							Data    model.ModelPricingConversion
						}
						modelManagementRequest(t, PreviewModelPricingConversion, http.MethodPost, "/api/option/model_pricing/convert", map[string]any{"model_name": tc.name, "pricing": draft}, &response)
						require.True(t, response.Success)
						require.NotEmpty(t, response.Data.Expression)
						assert.Equal(t, tc.mode, response.Data.CacheWriteMode)
						var prices struct {
							Success bool
							Data    model.ModelPricingDescription
						}
						modelManagementRequest(t, PreviewModelPricing, http.MethodPost, "/api/option/model_pricing/preview", map[string]any{"model_name": tc.name, "pricing": draft}, &prices)
						require.True(t, prices.Success)
						assert.Equal(t, response.Data.Effective, prices.Data.Effective)
						assert.Equal(t, response.Data.CacheWriteMode, prices.Data.CacheWriteMode)
						vars := billingexpr.UsedVars(response.Data.Expression)
						assert.Equal(t, tc.mode != model.CacheWriteNone, vars["cc"])
						assert.Equal(t, tc.mode == model.CacheWriteClaudeTTL, vars["cc1h"])
						assert.Equal(t, tc.mode == model.CacheWriteOpenAITTL, vars["cc12h"])
						if tc.configured != nil && *tc.configured == 0 {
							cost, _, err := billingexpr.RunExpr(response.Data.Expression, billingexpr.TokenParams{CC: 100, CC1h: 100, CC12h: 100})
							require.NoError(t, err)
							assert.Zero(t, cost, "explicit zero cache prices must stay free")
						}
					})
				}
				// GPT names price the 12h write at twice the configured write price.
				openai, err := model.PreviewModelPricingConversion("gpt-5.6-terra", model.PricingValues{"ModelRatio": 5.0, "CompletionRatio": 5.0, "CacheRatio": 0.1, "CreateCacheRatio": 1.25})
				require.NoError(t, err)
				assert.Equal(t, model.CacheWriteOpenAITTL, openai.CacheWriteMode)
				assert.Equal(t, `tier("base", p * 10 + c * 50 + cr * 1 + cc * 12.5 + cc12h * 25)`, openai.Expression)
				// The billing name remains authoritative for cache terms.
				require.NoError(t, (&model.Channel{Name: "Cache alias", Type: 1, Models: "opaque-cache-alias", ModelMapping: common.GetPointer(`{"opaque-cache-alias":"claude-sonnet-4-6"}`)}).Insert())
				alias, err := model.PreviewModelPricingConversion("opaque-cache-alias", model.PricingValues{"ModelRatio": 2.0})
				require.NoError(t, err)
				assert.Equal(t, `tier("base", p * 4 + c * 4)`, alias.Expression)
			})

			t.Run("same_price_cache_reads_preserve_overlapping_categories", func(t *testing.T) {
				for _, tc := range []struct {
					field, term             string
					ratio                   float64
					details                 dto.InputTokenDetails
					cost, unpricedCacheCost float64
				}{
					{"CreateCacheRatio", "cc * 75", 1.25, dto.InputTokenDetails{CachedTokens: 800, CacheWriteTokens: 500}, 97500, 79500},
					{"ImageRatio", "img * 120", 2, dto.InputTokenDetails{CachedTokens: 600, ImageTokens: 500}, 108000, 102000},
				} {
					t.Run(tc.field, func(t *testing.T) {
						converted, err := model.PreviewModelPricingConversion("overlapping-cache", model.PricingValues{"ModelRatio": 30.0, "CompletionRatio": 2.0, "CacheRatio": 1.0, tc.field: tc.ratio})
						require.NoError(t, err)
						assert.Equal(t, `tier("base", p * 60 + c * 120 + cr * 60 + `+tc.term+`)`, converted.Expression)
						usage := dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: tc.details}
						params := service.BuildTieredTokenParams(&usage, false, billingexpr.UsedVars(converted.Expression))
						cost, _, err := billingexpr.RunExpr(converted.Expression, params)
						require.NoError(t, err)
						assert.Equal(t, tc.cost, cost, "preserve the existing overlapping-category charge")
						unpriced := `tier("base", p * 60 + c * 120 + ` + tc.term + `)`
						params = service.BuildTieredTokenParams(&usage, false, billingexpr.UsedVars(unpriced))
						cost, _, err = billingexpr.RunExpr(unpriced, params)
						require.NoError(t, err)
						assert.Equal(t, tc.unpricedCacheCost, cost, "existing expressions without cr retain their OpenAI normalization")
					})
				}
			})
			t.Run("media_prices_and_request_adjustments", func(t *testing.T) {
				for _, quantity := range []string{"129", "-1", "0.5", "18446744073709551615"} {
					recorder := httptest.NewRecorder()
					ctx, _ := gin.CreateTestContext(recorder)
					ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", strings.NewReader(`{"model":"z-image","parameters":{"n":`+quantity+`}}`))
					ctx.Request.Header.Set("Content-Type", "application/json")
					Relay(ctx, types.RelayFormatOpenAIImage)
					assert.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
				}
				conflicting, err := model.PreviewModelPricingConversion("gemini-2.5-flash-conflict", model.PricingValues{"ModelRatio": 0.15, "AudioRatio": 10.0})
				require.NoError(t, err)
				assert.Contains(t, conflicting.UnsupportedReason, "Gemini and OpenAI audio prices differ")
				aliName := "z-image-media-conversion"
				require.NoError(t, (&model.Channel{Name: aliName, Models: aliName, Type: 17, Status: common.ChannelStatusEnabled}).Insert())
				aliConversion, err := model.PreviewModelPricingConversion(aliName, model.PricingValues{"ModelPrice": 0.04})
				require.NoError(t, err)
				require.Empty(t, aliConversion.UnsupportedReason)
				// Ali image billing lives in the alibaba task plugin, where prompt_extend is a
				// usage fact; the legacy conversion no longer adds a per-channel request multiplier.
				assert.Equal(t, `tier("image", fixed(0.04)) * image_count`, aliConversion.Expression)
				assert.Empty(t, aliConversion.BillingDetails.RequestRules)
				quantity := 3
				aliCost, _, err := billingexpr.RunExprWithRequest(aliConversion.Expression, billingexpr.TokenParams{}, billingexpr.RequestInput{ImageCount: &quantity, Body: []byte(`{"parameters":{"prompt_extend":true}}`)})
				require.NoError(t, err)
				assert.Equal(t, 120000.0, aliCost)
				for _, tc := range []struct {
					name    string
					pricing model.PricingValues
					usage   dto.Usage
					want    float64
				}{
					{"gemini-2.5-flash", model.PricingValues{"ModelRatio": 0.15, "CompletionRatio": 2.0, "CacheRatio": 0.1}, dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 100, AudioTokens: 200}}, 473},
					{"gemini-2.5-flash-zero", model.PricingValues{"ModelRatio": 0.0}, dto.Usage{PromptTokens: 1000, PromptTokensDetails: dto.InputTokenDetails{AudioTokens: 200}}, 200},
					{"media-audio", model.PricingValues{"ModelRatio": 1.0, "CompletionRatio": 3.0, "AudioRatio": 2.0, "AudioCompletionRatio": 4.0}, dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{TextTokens: 800, AudioTokens: 200}, CompletionTokenDetails: dto.OutputTokenDetails{TextTokens: 50, AudioTokens: 50}}, 3500},
					{"media-free-audio", model.PricingValues{"ModelRatio": 1.0, "AudioRatio": 0.0}, dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{AudioTokens: 200}, CompletionTokenDetails: dto.OutputTokenDetails{AudioTokens: 100}}, 1600},
					{"media-audio-cache", model.PricingValues{"ModelRatio": 1.0, "CompletionRatio": 3.0, "AudioRatio": 2.0, "CacheRatio": 0.1}, dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 400, AudioTokens: 200}, CompletionTokenDetails: dto.OutputTokenDetails{AudioTokens: 50}}, 2900},
					{"media-text-cache", model.PricingValues{"ModelRatio": 1.0, "CompletionRatio": 3.0, "AudioRatio": 2.0, "CacheRatio": 0.1}, dto.Usage{PromptTokens: 1000, CompletionTokens: 100, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 400}}, 1880},
					{"gemini-2.5-flash-overlap", model.PricingValues{"ModelRatio": 1.0}, dto.Usage{PromptTokens: 1000, PromptTokensDetails: dto.InputTokenDetails{CachedTokens: 800, AudioTokens: 500}}, 2100},
				} {
					t.Run(tc.name, func(t *testing.T) {
						converted, err := model.PreviewModelPricingConversion(tc.name, tc.pricing)
						require.NoError(t, err)
						require.Empty(t, converted.UnsupportedReason)
						require.NotNil(t, converted.BillingDetails.AudioInputPrice)
						params := service.BuildTieredTokenParams(&tc.usage, false, billingexpr.UsedVars(converted.Expression))
						cost, _, err := billingexpr.RunExpr(converted.Expression, params)
						require.NoError(t, err)
						assert.InDelta(t, tc.want, cost, 1e-9)
					})
				}
				converted, err := model.PreviewModelPricingConversion("dall-e-3", model.PricingValues{"ModelPrice": 0.04})
				require.NoError(t, err)
				require.Empty(t, converted.UnsupportedReason)
				assert.NotContains(t, converted.Expression, "256x256")
				assert.NotContains(t, converted.Expression, "512x512")
				smallImage, err := model.PreviewModelPricingConversion("dall-e-2", model.PricingValues{"ModelPrice": 0.02})
				require.NoError(t, err)
				require.Empty(t, smallImage.UnsupportedReason)
				assert.NotContains(t, smallImage.Expression, "1792")
				smallCost, _, err := billingexpr.RunExprWithRequest(smallImage.Expression, billingexpr.TokenParams{}, billingexpr.RequestInput{Body: []byte(`{"size":"256x256"}`), ImageCount: &quantity})
				require.NoError(t, err)
				assert.Equal(t, 24000.0, smallCost)
				for _, tc := range []struct {
					size, quality string
					count         int
					cost          float64
				}{
					{"1024x1024", "standard", 3, 120000},
					{"1024x1024", "hd", 2, 160000},
					{"1024x1792", "hd", 3, 360000},
				} {
					body, err := common.Marshal(map[string]any{"size": tc.size, "quality": tc.quality})
					require.NoError(t, err)
					cost, _, err := billingexpr.RunExprWithRequest(converted.Expression, billingexpr.TokenParams{}, billingexpr.RequestInput{Body: body, ImageCount: &tc.count})
					require.NoError(t, err)
					assert.InDelta(t, tc.cost, cost, 1e-9)
				}
				for _, channelType := range []int{20, 58} {
					name := fmt.Sprintf("gemini-3.5-flash-media-%d", channelType)
					require.NoError(t, (&model.Channel{Name: name, Models: name, Type: channelType}).Insert())
					require.NoError(t, db.Create(&model.Model{ModelName: name, Endpoints: `{"custom-text":"/v1/custom"}`}).Error)
					converted, err := model.PreviewModelPricingConversion(name, model.PricingValues{"ModelRatio": 0.75, "CompletionRatio": 6.0, "CacheRatio": 0.1})
					require.NoError(t, err)
					assert.Empty(t, converted.UnsupportedReason)
					assert.Equal(t, `tier("base", p * 1.5 + c * 9 + cr * 0.15)`, converted.Expression)
					assert.Nil(t, converted.BillingDetails.AudioInputPrice)
				}
			})
			require.NoError(t, db.Create(&model.Model{ModelName: "conversion-metadata", Endpoints: `{"image-generation":"/v1/images/generations"}`}).Error)
			preview, err = model.PreviewModelPricingConversion("conversion-metadata", model.PricingValues{"ModelPrice": float64(1)})
			require.NoError(t, err)
			assert.Equal(t, `tier("image", fixed(1)) * image_count`, preview.Expression)
			assert.Empty(t, preview.UnsupportedReason)
			for _, tc := range []struct {
				name, rule string
				nameRule   int
				endpoint   constant.EndpointType
			}{
				{"catalog-image-prefix-opaque", "catalog-image-prefix-", model.NameRulePrefix, constant.EndpointTypeImageGeneration},
				{"opaque-catalog-image-suffix", "-catalog-image-suffix", model.NameRuleSuffix, constant.EndpointTypeImageGeneration},
				{"before-catalog-image-middle-after", "catalog-image-middle", model.NameRuleContains, constant.EndpointTypeImageGeneration},
				{"opaque-catalog-video", "opaque-catalog-video", model.NameRuleExact, constant.EndpointTypeOpenAIVideo},
			} {
				t.Run("endpoint_metadata/"+tc.name, func(t *testing.T) {
					endpoints, err := common.Marshal(map[string]string{string(tc.endpoint): "/v1/fixture"})
					require.NoError(t, err)
					require.NoError(t, db.Create(&model.Model{ModelName: tc.rule, NameRule: tc.nameRule, Endpoints: string(endpoints)}).Error)
					converted, err := model.PreviewModelPricingConversion(tc.name, model.PricingValues{"ModelPrice": 1.0})
					require.NoError(t, err)
					if tc.endpoint == constant.EndpointTypeOpenAIVideo {
						assert.Empty(t, converted.Expression)
						assert.Equal(t, "Video pricing must be converted manually.", converted.UnsupportedReason)
					} else {
						assert.Equal(t, `tier("image", fixed(1)) * image_count`, converted.Expression)
					}
				})
			}
		})
	}
}

func TestMetadataSyncLocaleAndEndpointValidation(t *testing.T) {
	for _, locale := range []struct{ input, expected string }{{"zh", "zh"}, {"zh-CN", "zh"}, {"en", "en"}, {"ja", "ja"}} {
		t.Run(locale.input, func(t *testing.T) {
			normalized, ok := normalizeLocale(locale.input)
			assert.True(t, ok)
			assert.Equal(t, locale.expected, normalized)
		})
	}
	_, valid := normalizeLocale("invalid")
	assert.False(t, valid)
	for _, endpoint := range []string{`1`, `null`, `{"openai":false}`, `{"openai":{"path":"invalid"}}`} {
		assert.Error(t, model.ValidateModelEndpoints(endpoint))
	}
}

func TestVendorManagementDatabaseMatrix(t *testing.T) {
	for _, dialect := range []struct{ kind, env string }{{"sqlite", ""}, {"mysql", "TEST_MYSQL_DSN"}, {"postgres", "TEST_POSTGRES_DSN"}} {
		t.Run(dialect.kind, func(t *testing.T) {
			if dialect.env != "" && os.Getenv(dialect.env) == "" {
				t.Skip("set " + dialect.env)
			}
			db := modelManagementDB(t, dialect.kind, os.Getenv(dialect.env))

			t.Run("pricing_reads_keep_default_brands_without_writing_vendors", func(t *testing.T) {
				channel := model.Channel{Name: "Vendor fixture", Type: 1, Status: common.ChannelStatusEnabled}
				require.NoError(t, db.Create(&channel).Error)
				require.NoError(t, db.Create(&model.Ability{Model: "gemini-vendor-fixture", Group: "default", ChannelId: channel.Id, Enabled: true}).Error)
				model.RefreshPricing()
				model.GetPricing()
				vendors := model.GetVendors()
				require.Len(t, vendors, 1)
				assert.Equal(t, "Google", vendors[0].Name)
				assert.Equal(t, "Gemini.Color", vendors[0].Icon)
				assert.Negative(t, vendors[0].ID)
				var count int64
				require.NoError(t, db.Model(&model.Vendor{}).Count(&count).Error)
				assert.Zero(t, count)
				saved := model.Vendor{Name: "Google", Icon: "Gemini.Color"}
				require.NoError(t, saved.Insert())
				assert.Equal(t, saved.Id, model.GetVendors()[0].ID)
				require.NoError(t, saved.Delete())
				assert.Equal(t, vendors[0].ID, model.GetVendors()[0].ID)
				require.NoError(t, db.Model(&model.Vendor{}).Count(&count).Error)
				assert.Zero(t, count, "refresh must not recreate a deleted vendor")
			})
			t.Run("metadata_ownership_preview_merge_delete_and_rollback", func(t *testing.T) {
				source := model.Vendor{Name: "  Vendor Source  ", Icon: "Gemini.Color"}
				target := model.Vendor{Name: "Vendor Target", Description: "Keep target", Icon: "OpenAI"}
				require.NoError(t, source.Insert())
				require.NoError(t, target.Insert())
				assert.Equal(t, "Vendor Source", source.Name)
				assert.Error(t, (&model.Vendor{Name: "vendor source"}).Insert())
				assert.Error(t, (&model.Vendor{Name: " "}).Insert())
				require.NoError(t, db.Model(&model.Vendor{}).Where("id = ?", source.Id).Update("status", 0).Error)
				loaded, err := model.GetVendorByID(source.Id)
				require.NoError(t, err)
				staleVersion := loaded.Version
				edit := model.Vendor{Id: source.Id, Name: source.Name, Description: "Updated source", Icon: source.Icon, Version: loaded.Version}
				require.NoError(t, edit.Update())
				updated, err := model.GetVendorByID(source.Id)
				require.NoError(t, err)
				assert.Equal(t, loaded.CreatedTime, updated.CreatedTime)
				assert.Zero(t, updated.Status)
				edit.Version = staleVersion
				assert.ErrorIs(t, edit.Update(), model.ErrVendorConflict)

				one := model.Model{ModelName: "vendor-model-one", VendorID: source.Id, Icon: "Custom", Description: "Preserve description", Status: 0, SyncOfficial: 0}
				rule := model.Model{ModelName: "vendor-rule-", VendorID: source.Id, NameRule: model.NameRulePrefix, Status: 1, SyncOfficial: 1}
				require.NoError(t, one.Insert())
				require.NoError(t, rule.Insert())
				priceBefore, err := model.GetModelPricingSnapshot([]string{one.ModelName})
				require.NoError(t, err)
				require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{{ModelName: one.ModelName, ExpectedVersion: priceBefore.Entries[0].Version, Pricing: model.PricingValues{"ModelPrice": float64(0.25)}}}))
				priceBefore, err = model.GetModelPricingSnapshot([]string{one.ModelName})
				require.NoError(t, err)
				channel := model.Channel{Name: "Unchanged vendor channel", Type: 1, Status: common.ChannelStatusEnabled}
				require.NoError(t, db.Create(&channel).Error)
				ability := model.Ability{Model: one.ModelName, Group: "default", ChannelId: channel.Id, Enabled: true}
				require.NoError(t, db.Create(&ability).Error)
				linked, total, err := model.SearchVendors("", 0, 20, "linked")
				require.NoError(t, err)
				require.Len(t, linked, 1)
				assert.EqualValues(t, 1, total)
				assert.EqualValues(t, 2, linked[0].ModelCount)
				unlinked, _, err := model.SearchVendors("Vendor Target", 0, 20, "unlinked")
				require.NoError(t, err)
				require.Len(t, unlinked, 1)
				var references *model.VendorReferenceError
				err = model.DeleteVendors([]int{source.Id, target.Id})
				require.ErrorAs(t, err, &references)
				assert.EqualValues(t, 2, references.Counts[source.Id])
				_, err = model.GetVendorByID(target.Id)
				require.NoError(t, err, "bulk delete must not partially delete unreferenced vendors")
				var response struct {
					Success         bool
					Code            string
					ReferenceCounts map[int]int64 `json:"reference_counts"`
				}
				recorder := modelManagementRequest(t, PreviewVendorOperation, http.MethodPost, "/api/vendors/operations/preview", model.VendorOperation{Action: "delete", VendorIDs: []int{source.Id}}, &response)
				assert.Equal(t, http.StatusConflict, recorder.Code)
				assert.Equal(t, "VENDOR_REFERENCED", response.Code)
				assert.EqualValues(t, 2, response.ReferenceCounts[source.Id])

				disappearing := model.Vendor{Name: "Preview target"}
				require.NoError(t, disappearing.Insert())
				staleAssignment := model.VendorOperation{Action: "assign", ModelIDs: []int{one.Id}, TargetVendorID: disappearing.Id}
				stalePreview, err := model.PreviewVendorOperation(staleAssignment)
				require.NoError(t, err)
				staleAssignment.ExpectedVersion = stalePreview.Version
				require.NoError(t, disappearing.Delete())
				recorder = modelManagementRequest(t, ApplyVendorOperation, http.MethodPost, "/api/vendors/operations", staleAssignment, &response)
				assert.Equal(t, http.StatusConflict, recorder.Code)
				assert.Equal(t, "VENDOR_CONFLICT", response.Code)

				assign := model.VendorOperation{Action: "assign", ModelIDs: []int{one.Id}, TargetVendorID: target.Id}
				preview, err := model.PreviewVendorOperation(assign)
				require.NoError(t, err)
				require.Len(t, preview.Models, 1)
				assign.ExpectedVersion = preview.Version
				one.Description = "Updated in the same timestamp"
				require.NoError(t, db.Model(&model.Model{}).Where("id = ?", one.Id).Update("description", one.Description).Error)
				_, err = model.ApplyVendorOperation(assign)
				assert.ErrorIs(t, err, model.ErrVendorConflict)
				preview, err = model.PreviewVendorOperation(assign)
				require.NoError(t, err)
				assign.ExpectedVersion = preview.Version
				target.Description = "New target description"
				require.NoError(t, target.Update())
				_, err = model.ApplyVendorOperation(assign)
				assert.ErrorIs(t, err, model.ErrVendorConflict)
				preview, err = model.PreviewVendorOperation(assign)
				require.NoError(t, err)
				assign.ExpectedVersion = preview.Version
				result, err := model.ApplyVendorOperation(assign)
				require.NoError(t, err)
				assert.Equal(t, []int{one.Id}, result.UpdatedModels)
				var after model.Model
				after = model.Model{}
				require.NoError(t, db.First(&after, one.Id).Error)
				assert.Equal(t, target.Id, after.VendorID)
				assert.Equal(t, one.Description, after.Description)
				assert.Equal(t, one.Icon, after.Icon)
				assert.Equal(t, one.Status, after.Status)
				after = model.Model{}
				require.NoError(t, db.First(&after, rule.Id).Error)
				assert.Equal(t, source.Id, after.VendorID)
				assign.TargetVendorID = 0
				preview, err = model.PreviewVendorOperation(assign)
				require.NoError(t, err)
				assign.ExpectedVersion = preview.Version
				_, err = model.ApplyVendorOperation(assign)
				require.NoError(t, err)
				after = model.Model{}
				require.NoError(t, db.First(&after, one.Id).Error)
				assert.Zero(t, after.VendorID)
				after.VendorID = -1001
				assert.Error(t, after.Update(), "display-only vendors cannot become stored references")

				merge := model.VendorOperation{Action: "merge", VendorIDs: []int{source.Id}, TargetVendorID: target.Id}
				preview, err = model.PreviewVendorOperation(merge)
				require.NoError(t, err)
				merge.ExpectedVersion = preview.Version
				require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("vendor_delete_failure", func(tx *gorm.DB) {
					if tx.Statement.Table == "vendors" {
						tx.AddError(errors.New("injected vendor delete failure"))
					}
				}))
				_, err = model.ApplyVendorOperation(merge)
				require.Error(t, err)
				require.NoError(t, db.Callback().Delete().Remove("vendor_delete_failure"))
				after = model.Model{}
				require.NoError(t, db.First(&after, rule.Id).Error)
				assert.Equal(t, source.Id, after.VendorID, "ownership updates roll back when deletion fails")
				_, err = model.GetVendorByID(source.Id)
				require.NoError(t, err)
				_, err = model.ApplyVendorOperation(merge)
				require.NoError(t, err)
				_, err = model.GetVendorByID(source.Id)
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
				retained, err := model.GetVendorByID(target.Id)
				require.NoError(t, err)
				assert.Equal(t, target.Description, retained.Description)
				assert.Equal(t, "OpenAI", retained.Icon)
				assert.EqualValues(t, 1, retained.ModelCount)
				after = one
				after.VendorID = source.Id
				assert.Error(t, after.Update(), "deleted vendors cannot acquire new references")
				priceAfter, err := model.GetModelPricingSnapshot([]string{one.ModelName})
				require.NoError(t, err)
				assert.Equal(t, priceBefore.Entries[0], priceAfter.Entries[0], "assignment and merge preserve model pricing")
				var retainedChannel model.Channel
				require.NoError(t, db.First(&retainedChannel, channel.Id).Error)
				assert.Equal(t, channel.Name, retainedChannel.Name)
				assert.Equal(t, channel.Status, retainedChannel.Status)
				var retainedAbility model.Ability
				require.NoError(t, db.Where("model = ? AND channel_id = ?", one.ModelName, channel.Id).First(&retainedAbility).Error)
				assert.Equal(t, ability.Group, retainedAbility.Group)
				assert.Equal(t, ability.Enabled, retainedAbility.Enabled)
			})
			t.Run("concurrent_create_and_delete_never_orphan_model", func(t *testing.T) {
				vendor := model.Vendor{Name: "Concurrent owner"}
				require.NoError(t, vendor.Insert())
				var createErr, deleteErr error
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					createErr = (&model.Model{ModelName: "concurrent-owned-model", VendorID: vendor.Id}).Insert()
				}()
				go func() { defer wg.Done(); deleteErr = vendor.Delete() }()
				wg.Wait()
				if createErr == nil {
					require.Error(t, deleteErr)
				} else {
					require.NoError(t, deleteErr)
				}
				var models []model.Model
				require.NoError(t, db.Where("model_name = ?", "concurrent-owned-model").Find(&models).Error)
				if len(models) != 0 {
					_, err := model.GetVendorByID(models[0].VendorID)
					require.NoError(t, err)
				}
			})
		})
	}
}

func TestModelDeletionDatabaseMatrix(t *testing.T) {
	for _, dialect := range []struct{ kind, env string }{{"sqlite", ""}, {"mysql", "TEST_MYSQL_DSN"}, {"postgres", "TEST_POSTGRES_DSN"}} {
		t.Run(dialect.kind, func(t *testing.T) {
			if dialect.env != "" && os.Getenv(dialect.env) == "" {
				t.Skip("set " + dialect.env + " to run this database")
			}
			db := modelManagementDB(t, dialect.kind, os.Getenv(dialect.env))

			var response struct {
				Success bool
				Data    model.ModelDeleteResult
			}
			metadataOnly := model.Model{ModelName: "metadata-only", Status: 1}
			require.NoError(t, metadataOnly.Insert())
			channel := model.Channel{Name: "Retained channel", Type: 1, Key: "fixture-key", Models: metadataOnly.ModelName, Group: "default", Status: common.ChannelStatusEnabled}
			require.NoError(t, db.Create(&channel).Error)
			require.NoError(t, channel.UpdateAbilities(db))
			recorder := modelManagementRequest(t, func(c *gin.Context) {
				c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(metadataOnly.Id)}}
				DeleteModelMeta(c)
			}, http.MethodDelete, "/api/models/"+strconv.Itoa(metadataOnly.Id), nil, &response)
			require.True(t, response.Success, recorder.Body.String())
			assert.Equal(t, model.ModelDeleteResult{DeletedCount: 1}, response.Data)
			var retained model.Channel
			require.NoError(t, db.First(&retained, channel.Id).Error)
			assert.Equal(t, channel.Models, retained.Models)
			var count int64
			require.NoError(t, db.Model(&model.Ability{}).Where("channel_id = ?", channel.Id).Count(&count).Error)
			assert.EqualValues(t, 1, count)

			for _, rule := range []int{model.NameRuleExact, model.NameRulePrefix, model.NameRuleContains, model.NameRuleSuffix} {
				t.Run(fmt.Sprintf("rule_%d_exact_names_only_atomic_and_cached", rule), func(t *testing.T) {
					name := fmt.Sprintf("delete-rule-%d", rule)
					first := model.Model{ModelName: name, NameRule: rule, Status: 1}
					second := model.Model{ModelName: name + "-second", Status: 1}
					require.NoError(t, first.Insert())
					require.NoError(t, second.Insert())
					mapping := `{"` + name + `":"upstream-name"}`
					priority, weight := int64(7), uint(9)
					channels := []model.Channel{
						{Name: "Enabled", Type: 1, Key: "fixture-key", Models: name + "," + name + "-keep," + second.ModelName, Group: "default,vip", Status: common.ChannelStatusEnabled, ModelMapping: &mapping, Priority: &priority, Weight: &weight},
						{Name: "Disabled", Type: 1, Models: name + ",prefix-" + name, Group: "disabled-group", Status: common.ChannelStatusManuallyDisabled},
						{Name: "Last model", Type: 1, Models: name, Group: "last-model-group", Status: common.ChannelStatusEnabled},
						{Name: "Case-sensitive name", Type: 1, Models: strings.ToUpper(name), Group: "case-group", Status: common.ChannelStatusEnabled},
					}
					for i := range channels {
						require.NoError(t, db.Create(&channels[i]).Error)
						require.NoError(t, channels[i].UpdateAbilities(db))
					}
					common.MemoryCacheEnabled = true
					model.InitChannelCache()
					cached, err := model.GetRandomSatisfiedChannel("default", name, 0, nil)
					require.NoError(t, err)
					require.NotNil(t, cached)
					baseline, err := model.GetModelPricingSnapshot([]string{name})
					require.NoError(t, err)
					require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{{ModelName: name, ExpectedVersion: baseline.EmptyVersion, Pricing: model.PricingValues{"ModelPrice": float64(2)}}}))
					pricingBefore, err := model.GetModelPricingSnapshot([]string{name, second.ModelName})
					require.NoError(t, err)
					if rule != model.NameRuleExact {
						_, err := model.DeleteModelMetadata([]int{first.Id, second.Id}, true, true)
						require.EqualError(t, err, "only exact-match models can be removed from channels")
						after, err := model.GetModelPricingSnapshot([]string{name, second.ModelName})
						require.NoError(t, err)
						assert.Equal(t, pricingBefore, after)
						require.NoError(t, db.Model(&model.Model{}).Where("id IN ?", []int{first.Id, second.Id}).Count(&count).Error)
						assert.EqualValues(t, 2, count)
						for _, original := range channels {
							var after model.Channel
							require.NoError(t, db.First(&after, original.Id).Error)
							assert.Equal(t, original, after)
						}
						return
					}
					body := map[string]any{"model_ids": []int{first.Id, second.Id, first.Id}, "remove_from_channels": true}
					// A failure at the final metadata delete must undo every earlier
					// channel/ability write and keep the published cache intact.
					require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("fail_model_deletion", func(tx *gorm.DB) {
						if tx.Statement.Table == "models" {
							tx.AddError(errors.New("injected model deletion failure"))
						}
					}))
					result, err := model.DeleteModelMetadata([]int{first.Id, second.Id}, true, false)
					require.Error(t, err)
					assert.Zero(t, result)
					require.NoError(t, db.Callback().Delete().Remove("fail_model_deletion"))
					for _, original := range channels {
						var after model.Channel
						require.NoError(t, db.First(&after, original.Id).Error)
						assert.Equal(t, original, after)
					}
					cached, err = model.GetRandomSatisfiedChannel("default", name, 0, nil)
					require.NoError(t, err)
					require.NotNil(t, cached)
					recorder := modelManagementRequest(t, BatchDeleteModelMeta, http.MethodPost, "/api/models/delete", body, &response)
					require.True(t, response.Success, recorder.Body.String())
					assert.Equal(t, model.ModelDeleteResult{DeletedCount: 2, UpdatedChannels: 3}, response.Data)
					for i, original := range channels {
						var after model.Channel
						require.NoError(t, db.First(&after, original.Id).Error)
						original.Models = []string{name + "-keep", "prefix-" + name, "", strings.ToUpper(name)}[i]
						assert.Equal(t, original, after, "only the model list changes")
					}
					var abilities []model.Ability
					require.NoError(t, db.Where("channel_id IN ?", []int{channels[0].Id, channels[1].Id, channels[2].Id, channels[3].Id}).Find(&abilities).Error)
					assert.Len(t, abilities, 4)
					for _, ability := range abilities {
						assert.NotEqual(t, name, ability.Model)
						assert.NotEqual(t, second.ModelName, ability.Model)
						assert.NotEmpty(t, ability.Model)
						if ability.ChannelId == channels[0].Id {
							assert.Equal(t, &priority, ability.Priority)
							assert.Equal(t, weight, ability.Weight)
							assert.True(t, ability.Enabled)
						}
						if ability.ChannelId == channels[1].Id {
							assert.False(t, ability.Enabled)
						}
					}
					for _, group := range []string{"default", "vip", "last-model-group"} {
						cached, _ = model.GetRandomSatisfiedChannel(group, name, 0, nil)
						assert.Nil(t, cached)
					}
					cached, err = model.GetRandomSatisfiedChannel("default", name+"-keep", 0, nil)
					require.NoError(t, err)
					require.NotNil(t, cached)
					pricingAfter, err := model.GetModelPricingSnapshot([]string{name, second.ModelName})
					require.NoError(t, err)
					assert.Equal(t, pricingBefore, pricingAfter)
					require.NoError(t, db.Model(&model.Model{}).Where("id IN ?", []int{first.Id, second.Id}).Count(&count).Error)
					assert.Zero(t, count)
					_, err = model.DeleteModelMetadata([]int{first.Id}, true, false)
					assert.Error(t, err, "stale selections cannot delete newly created records")
				})
			}
			for _, removeChannels := range []bool{false, true} {
				t.Run(fmt.Sprintf("pricing_removal_channels_%t", removeChannels), func(t *testing.T) {
					name := fmt.Sprintf("remove-pricing-%t", removeChannels)
					metadata := model.Model{ModelName: name, NameRule: model.NameRuleExact, Status: 1}
					require.NoError(t, metadata.Insert())
					keep := name + "-keep"
					channel := model.Channel{Name: "Independent pricing removal", Type: 1, Models: name + "," + keep, Group: "pricing-removal", Status: common.ChannelStatusEnabled}
					require.NoError(t, db.Create(&channel).Error)
					require.NoError(t, channel.UpdateAbilities(db))
					model.InitChannelCache()
					baseline, err := model.GetModelPricingSnapshot([]string{name, keep})
					require.NoError(t, err)
					pricing := model.PricingValues{"ModelRatio": float64(1), "ModelPrice": float64(0), "CompletionRatio": float64(2), "CacheRatio": float64(0.1), "CreateCacheRatio": float64(1.25), "ImageRatio": float64(3), "AudioRatio": float64(4), "AudioCompletionRatio": float64(5), "billing_setting.billing_mode": "tiered_expr", "billing_setting.billing_expr": `tier("base", p * 2 + c * 4)`}
					require.NoError(t, model.UpdateModelPricing([]model.ModelPricingChange{
						{ModelName: name, ExpectedVersion: baseline.EmptyVersion, Pricing: pricing},
						{ModelName: keep, ExpectedVersion: baseline.EmptyVersion, Pricing: model.PricingValues{"ModelPrice": float64(9)}},
					}))
					before, err := model.GetModelPricingSnapshot([]string{name, keep})
					require.NoError(t, err)
					body := map[string]any{"model_ids": []int{metadata.Id}, "remove_from_channels": removeChannels, "remove_pricing": true}
					for _, single := range []bool{false, true} {
						recorder := modelManagementRequest(t, func(c *gin.Context) {
							c.Set("role", common.RoleAdminUser)
							if single {
								c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(metadata.Id)}}
								DeleteModelMeta(c)
							} else {
								BatchDeleteModelMeta(c)
							}
						}, http.MethodPost, "/api/models/delete?remove_pricing=true", body, nil)
						assert.Equal(t, http.StatusForbidden, recorder.Code, "pricing permissions cannot be bypassed through deletion")
					}
					updates := 0
					require.NoError(t, db.Callback().Update().Before("gorm:update").Register("fail_deleted_pricing", func(tx *gorm.DB) {
						if tx.Statement.Table == "options" {
							updates++
							if updates == 3 {
								tx.AddError(errors.New("injected pricing deletion failure"))
							}
						}
					}))
					_, err = model.DeleteModelMetadata([]int{metadata.Id}, removeChannels, true)
					require.Error(t, err)
					require.NoError(t, db.Callback().Update().Remove("fail_deleted_pricing"))
					after, err := model.GetModelPricingSnapshot([]string{name, keep})
					require.NoError(t, err)
					assert.Equal(t, before, after, "partial option writes roll back")
					assert.Equal(t, billing_setting.BillingModeTieredExpr, billing_setting.GetBillingMode(name), "failed deletion must not publish new runtime pricing")
					var retained model.Model
					require.NoError(t, db.First(&retained, metadata.Id).Error)
					var channelAfter model.Channel
					require.NoError(t, db.First(&channelAfter, channel.Id).Error)
					assert.Equal(t, channel.Models, channelAfter.Models)
					_, err = model.DeleteModelMetadata([]int{metadata.Id, 999999}, removeChannels, true)
					assert.Error(t, err, "a missing model aborts the whole batch")
					recorder := modelManagementRequest(t, BatchDeleteModelMeta, http.MethodPost, "/api/models/delete", body, &response)
					require.True(t, response.Success, recorder.Body.String())
					after, err = model.GetModelPricingSnapshot([]string{name, keep})
					require.NoError(t, err)
					assert.Empty(t, after.Entries[0].Configured)
					assert.Equal(t, before.Entries[1], after.Entries[1], "name rules do not expand pricing deletion")
					assert.Equal(t, billing_setting.BillingModeRatio, billing_setting.GetBillingMode(name))
					_, hasExpr := billing_setting.GetBillingExpr(name)
					assert.False(t, hasExpr)
					require.NoError(t, db.First(&channelAfter, channel.Id).Error)
					expectedModels := channel.Models
					if removeChannels {
						expectedModels = keep
					}
					assert.Equal(t, expectedModels, channelAfter.Models)
					require.NoError(t, db.Model(&model.Model{}).Where("id = ?", metadata.Id).Count(&count).Error)
					assert.Zero(t, count)
				})
			}
			for _, ids := range [][]int{nil, {0}, {-1}, make([]int, 1001)} {
				_, err := model.DeleteModelMetadata(ids, true, false)
				assert.Error(t, err)
			}
		})
	}
}
