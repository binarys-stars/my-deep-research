/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package eino

import (
	"context"
	"fmt"
	"github.com/binarys-stars/my-deep-research/biz/infra"
	"github.com/binarys-stars/my-deep-research/biz/model"
	"os"
	"path/filepath"
	"time"

	"github.com/RanFeng/ilog"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func loadReporterMsg(ctx context.Context, name string, opts ...any) (output []*schema.Message, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		sysPrompt, err := infra.GetPromptTemplate(ctx, name)
		if err != nil {
			ilog.EventInfo(ctx, "get prompt template fail")
			return err
		}

		promptTemp := prompt.FromMessages(schema.Jinja2,
			schema.SystemMessage(sysPrompt),
			schema.MessagesPlaceholder("user_input", true),
		)

		msg := []*schema.Message{}
		msg = append(msg,
			schema.UserMessage(fmt.Sprintf("# Research Requirements\n\n## Task\n\n %v \n\n## Description\n\n %v", state.CurrentPlan.Title, state.CurrentPlan.Thought)),
			schema.SystemMessage("IMPORTANT: Structure your report according to the format in the prompt. Remember to include:\n\n1. Key Points - A bulleted list of the most important findings\n2. Overview - A brief introduction to the topic\n3. Detailed Analysis - Organized into logical sections\n4. Survey Note (optional) - For more comprehensive reports\n5. Key Citations - List all references at the end\n\nFor citations, DO NOT include inline citations in the text. Instead, place all citations in the 'Key Citations' section at the end using the format: `- [Source Title](URL)`. Include an empty line between each citation for better readability.\n\nPRIORITIZE USING MARKDOWN TABLES for data presentation and comparison. Use tables whenever presenting comparative data, statistics, features, or options. Structure tables with clear headers and aligned columns. Example table format:\n\n| Feature | Description | Pros | Cons |\n|---------|-------------|------|------|\n| Feature 1 | Description 1 | Pros 1 | Cons 1 |\n| Feature 2 | Description 2 | Pros 2 | Cons 2 |"),
		)
		for _, step := range state.CurrentPlan.Steps {
			if step.ExecutionRes != nil && *step.ExecutionRes != "" {
				msg = append(msg, schema.UserMessage(fmt.Sprintf("Below are some observations for the research task:\n\n %v", *step.ExecutionRes)))
			}
		}
		variables := map[string]any{
			"locale":              state.Locale,
			"max_step_num":        state.MaxStepNum,
			"max_plan_iterations": state.MaxPlanIterations,
			"CURRENT_TIME":        time.Now().Format("2006-01-02 15:04:05"),
			"user_input":          msg,
		}
		output, err = promptTemp.Format(ctx, variables)
		return err
	})
	return output, err
}

func routerReporter(ctx context.Context, input *schema.Message, opts ...any) (output string, err error) {
	err = compose.ProcessState[*model.State](ctx, func(_ context.Context, state *model.State) error {
		defer func() {
			output = state.Goto
		}()
		ilog.EventInfo(ctx, "report_end", "report", input.Content)
		state.Goto = compose.END
		// 将 input.content 内容写入到.md文件,路径在create下
		// 生成带路径的文件名,文件保存到create目录
		fileName := time.Now().Format("20060102150405") + "_report.md"
		filePath := filepath.Join("create", fileName) // 使用filepath.Join处理路径分隔符

		// 确保create目录存在
		if err := os.MkdirAll("create", 0755); err != nil {
			ilog.EventError(ctx, err, "create directory failed", "dir", "create")
			return nil // 目录创建失败继续执行，不影响主流程
		}

		// 创建文件
		f, err := os.Create(filePath)
		if err != nil {
			ilog.EventError(ctx, err, "create report file failed", "file", filePath)
			return nil // 文件创建失败继续执行，不影响主流程
		}
		defer func() {
			// 确保文件关闭
			if err := f.Close(); err != nil {
				ilog.EventError(ctx, err, "close report file failed", "file", filePath)
			}
		}()

		// 写入内容
		if _, err = f.WriteString(input.Content); err != nil {
			ilog.EventError(ctx, err, "write report content failed", "file", filePath)
			return nil // 写入失败继续执行，不影响主流程
		}
		return nil
	})
	return output, nil
}

func NewReporter[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()

	_ = cag.AddLambdaNode("load", compose.InvokableLambdaWithOption(loadReporterMsg))
	_ = cag.AddChatModelNode("agent", infra.ChatModel)
	_ = cag.AddLambdaNode("router", compose.InvokableLambdaWithOption(routerReporter))

	_ = cag.AddEdge(compose.START, "load")
	_ = cag.AddEdge("load", "agent")
	_ = cag.AddEdge("agent", "router")
	_ = cag.AddEdge("router", compose.END)
	return cag
}
