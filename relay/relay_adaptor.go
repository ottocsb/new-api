package relay

import (
	"newapi/constant"
	"newapi/relay/channel"
	"newapi/relay/channel/advancedcustom"
	"newapi/relay/channel/ali"
	"newapi/relay/channel/aws"
	"newapi/relay/channel/baidu"
	"newapi/relay/channel/baidu_v2"
	"newapi/relay/channel/claude"
	"newapi/relay/channel/cloudflare"
	"newapi/relay/channel/codex"
	"newapi/relay/channel/cohere"
	"newapi/relay/channel/coze"
	"newapi/relay/channel/deepseek"
	"newapi/relay/channel/dify"
	"newapi/relay/channel/gemini"
	"newapi/relay/channel/jina"
	"newapi/relay/channel/minimax"
	"newapi/relay/channel/mistral"
	"newapi/relay/channel/mokaai"
	"newapi/relay/channel/moonshot"
	"newapi/relay/channel/ollama"
	"newapi/relay/channel/openai"
	"newapi/relay/channel/palm"
	"newapi/relay/channel/perplexity"
	"newapi/relay/channel/replicate"
	"newapi/relay/channel/siliconflow"
	"newapi/relay/channel/submodel"
	"newapi/relay/channel/tencent"
	"newapi/relay/channel/vertex"
	"newapi/relay/channel/volcengine"
	"newapi/relay/channel/xai"
	"newapi/relay/channel/xunfei"
	"newapi/relay/channel/zhipu"
	"newapi/relay/channel/zhipu_4v"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeZhipuV4:
		return &zhipu_4v.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	case constant.APITypeAws:
		return &aws.Adaptor{}
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	case constant.APITypeMokaAI:
		return &mokaai.Adaptor{}
	case constant.APITypeVolcEngine:
		return &volcengine.Adaptor{}
	case constant.APITypeBaiduV2:
		return &baidu_v2.Adaptor{}
	case constant.APITypeOpenRouter:
		return &openai.Adaptor{}
	case constant.APITypeXinference:
		return &openai.Adaptor{}
	case constant.APITypeXai:
		return &xai.Adaptor{}
	case constant.APITypeCoze:
		return &coze.Adaptor{}
	case constant.APITypeMoonshot:
		return &moonshot.Adaptor{} // Moonshot uses Claude API
	case constant.APITypeSubmodel:
		return &submodel.Adaptor{}
	case constant.APITypeMiniMax:
		return &minimax.Adaptor{}
	case constant.APITypeReplicate:
		return &replicate.Adaptor{}
	case constant.APITypeCodex:
		return &codex.Adaptor{}
	case constant.APITypeAdvancedCustom:
		return &advancedcustom.Adaptor{}
	}
	return nil
}
