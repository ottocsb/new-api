package relay

import (
	"strconv"

	"newapi/constant"
	"newapi/relay/channel"
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
	"newapi/relay/channel/jimeng"
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
	taskali "newapi/relay/channel/task/ali"
	taskdoubao "newapi/relay/channel/task/doubao"
	taskGemini "newapi/relay/channel/task/gemini"
	"newapi/relay/channel/task/hailuo"
	taskjimeng "newapi/relay/channel/task/jimeng"
	"newapi/relay/channel/task/kling"
	tasksora "newapi/relay/channel/task/sora"
	"newapi/relay/channel/task/suno"
	taskvertex "newapi/relay/channel/task/vertex"
	taskVidu "newapi/relay/channel/task/vidu"
	"newapi/relay/channel/tencent"
	"newapi/relay/channel/vertex"
	"newapi/relay/channel/volcengine"
	"newapi/relay/channel/xai"
	"newapi/relay/channel/xunfei"
	"newapi/relay/channel/zhipu"
	"newapi/relay/channel/zhipu_4v"
	"github.com/gin-gonic/gin"
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
	case constant.APITypeJimeng:
		return &jimeng.Adaptor{}
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
	}
	return nil
}

func GetTaskPlatform(c *gin.Context) constant.TaskPlatform {
	channelType := c.GetInt("channel_type")
	if channelType > 0 {
		return constant.TaskPlatform(strconv.Itoa(channelType))
	}
	return constant.TaskPlatform(c.GetString("platform"))
}

func GetTaskAdaptor(platform constant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case constant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	if channelType, err := strconv.ParseInt(string(platform), 10, 64); err == nil {
		switch channelType {
		case constant.ChannelTypeAli:
			return &taskali.TaskAdaptor{}
		case constant.ChannelTypeKling:
			return &kling.TaskAdaptor{}
		case constant.ChannelTypeJimeng:
			return &taskjimeng.TaskAdaptor{}
		case constant.ChannelTypeVertexAi:
			return &taskvertex.TaskAdaptor{}
		case constant.ChannelTypeVidu:
			return &taskVidu.TaskAdaptor{}
		case constant.ChannelTypeDoubaoVideo, constant.ChannelTypeVolcEngine:
			return &taskdoubao.TaskAdaptor{}
		case constant.ChannelTypeSora, constant.ChannelTypeOpenAI:
			return &tasksora.TaskAdaptor{}
		case constant.ChannelTypeGemini:
			return &taskGemini.TaskAdaptor{}
		case constant.ChannelTypeMiniMax:
			return &hailuo.TaskAdaptor{}
		}
	}
	return nil
}
