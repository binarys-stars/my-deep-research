package infra

import (
	"context"
	"github.com/binarys-stars/my-deep-research/biz/model"
	"github.com/binarys-stars/my-deep-research/conf"
	"github.com/cloudwego/eino-ext/components/model/openai"
	openai3 "github.com/cloudwego/eino-ext/libs/acl/openai"
	"github.com/getkin/kin-openapi/openapi3gen"
)

var (
	ChatModel *openai.ChatModel
	PlanModel *openai.ChatModel
)

func InitModel() {
	// 初始化 ChatModel
	config := &openai.ChatModelConfig{
		BaseURL: conf.Config.Model.BaseURL,
		APIKey:  conf.Config.Model.APIKey,
		Model:   conf.Config.Model.DefaultModel,
	}
	ChatModel, _ = openai.NewChatModel(context.Background(), config)

	// 初始化 PlanModel
	/*	先使用openapi3gen.NewSchemaRefForValue()根据model.Plan{}结构体自动生成对应的OpenAPI Schema定义
		创建增强版配置，除基础设置外，还通过ResponseFormat指定了模型必须返回符合该JSON Schema的结构化输出
		设置JSONSchema格式，包含名称、非严格模式和生成的Schema值
		最后创建专用的PlanModel实例*/
	planSchema, _ := openapi3gen.NewSchemaRefForValue(&model.Plan{}, nil)
	planConfig := &openai.ChatModelConfig{
		BaseURL: conf.Config.Model.BaseURL,
		APIKey:  conf.Config.Model.APIKey,
		Model:   conf.Config.Model.DefaultModel,
		ResponseFormat: &openai3.ChatCompletionResponseFormat{
			Type: openai3.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai3.ChatCompletionResponseFormatJSONSchema{
				Name:   "plan",
				Strict: false,
				Schema: planSchema.Value,
			},
		},
	}
	PlanModel, _ = openai.NewChatModel(context.Background(), planConfig)
}
