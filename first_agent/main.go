package main

import (
	"context"
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
	chatModel, err := chatModel.WithTools(toolInfos)
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
