package main

import (
	"context"
	"github.com/binarys-stars/my-deep-research/logs"
	"github.com/cloudwego/eino-examples/flow/agent/deer-go/biz/infra"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"log"
)

func main() {
	ctx := context.Background()
	// 创建llm
	log.Printf("===create llm===\n")
	chatModel := createOpenAIChatModel(ctx)

	// 创建工具
	log.Printf("===get tools===\n")
	tools := getTools()

	// 构造工具信息
	log.Printf("===build tool infos===\n")
	toolInfos := buildToolsInfo(ctx, tools)
	// 工具信息绑定到 ChatModel
	if toolInfos == nil {
		logs.Errorf("build tool infos failed")
		return
	}

	log.Printf("===bind tools to chat model===\n")
	chatModel, err = chatModel.WithTools(toolInfos)
	if err != nil {
		logs.Errorf("bind tools to chat model failed, err=%v", err)
		return
	}

	// 创建工具节点
	log.Printf("===build tools node===\n")
	toolsNode := buildToolsNode(ctx, tools)
	if toolsNode == nil {
		logs.Errorf("build tools node failed")
		return
	}

	// 创建输入信息
	log.Printf("===create input messages===\n")
	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: "帮我在bilibili上打开一个关于缓解焦虑的雨声的视频",
		},
	}

	err = runAgentByReAct(ctx, chatModel, compose.ToolsNodeConfig{
		Tools: tools,
	}, messages)
	if err != nil {
		logs.Errorf("run agent by react failed, err=%v", err)
		return
	}

}

func getTools() []tool.BaseTool {
	tools := []tool.BaseTool{
		GetDuckDuckGoTool(),
		GetBrowserTool(),
	}

	return tools
}

func buildToolsInfo(ctx context.Context, tools []tool.BaseTool) []*schema.ToolInfo {
	var toolInfos []*schema.ToolInfo
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			logs.Errorf("get tool info failed, err=%v", err)
			return nil
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos
}

func buildToolsNode(ctx context.Context, tools []tool.BaseTool) *compose.ToolsNode {
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: tools,
	})
	if err != nil {
		logs.Errorf("create tools node failed, err=%v", err)
		return nil
	}
	return toolsNode
}

func runAgentByChain(ctx context.Context, chatModel model.ChatModel, tools *compose.ToolsNode, messages []*schema.Message) error {
	// 通过Chain构建Agent
	// 构建完整的处理链
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.
		AppendChatModel(chatModel, compose.WithNodeName("chat_model_1")).
		AppendToolsNode(tools, compose.WithNodeName("search_tools"))

	//编译并运行 chain
	cagent, err := chain.Compile(ctx)
	if err != nil {
		return err
	}

	// 执行Agent
	resp, err := cagent.Invoke(ctx, messages)

	if err != nil {
		return err
	}

	// 输出结果
	for idx, msg := range resp {
		logs.Infof("\n")
		logs.Infof("message index: %d, role: %s", idx, msg.Role)
		logs.Infof("content: %s", msg.Content)
	}

	return nil
}

func runAgentByReAct(ctx context.Context, chatModel model.ToolCallingChatModel, tools compose.ToolsNodeConfig, messages []*schema.Message) error {
	ragent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      tools,
	})
	if err != nil {
		return err
	}
	// 执行Agent
	rmsg, err := ragent.Generate(ctx, messages, agent.WithComposeOptions(compose.WithCallbacks(&infra.LoggerCallback{})))
	if err != nil {
		return err
	}
	// 输出结果
	logs.Infof("\n")
	logs.Infof("message index: %d, role: %s, content: %s", 0, rmsg.Role, rmsg.Content)

	return nil
}
