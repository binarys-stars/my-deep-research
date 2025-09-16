package main

import (
	"context"
	"github.com/binarys-stars/my-deep-research/logs"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"os/exec"
	"runtime"
)

func GetDuckDuckGoTool() tool.InvokableTool {
	ctx := context.Background()
	// 创建 DuckDuckGo 工具
	searchTool, err := duckduckgo.NewTextSearchTool(ctx, &duckduckgo.Config{})
	if err != nil {
		logs.Errorf("NewTextSearchTool failed, err=%v", err)
		return nil
	}
	return searchTool
}

type BrowserParam struct {
	Url string `json:"url"`
}

func OpenBrowser(_ context.Context, params *BrowserParam) (string, error) {
	// 打开浏览器
	logs.Infof("open browser, url=%s", params.Url)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", params.Url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "chrome", params.Url)
	case "darwin":
		cmd = exec.Command("open", params.Url)
	default:
		return "unsupported platform", nil
	}
	err := cmd.Run()
	if err != nil {
		logs.Errorf("open browser failed, err=%v", err)
		return "open browser failed", err
	}
	return "open browser success", nil
}
func GetBrowserTool() tool.InvokableTool {
	// 创建浏览器工具
	info := &schema.ToolInfo{
		Name: "open_browser",
		Desc: "打开浏览器，参数是一个包含url字段的json对象，例如：{\"url\": \"https://www.google.com\"}。如果不支持打开浏览器，请返回错误信息",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"url": {
				Type: "string",
				Desc: "要打开的URL",
			},
		}),
	}
	return utils.NewTool(info, OpenBrowser)
}
