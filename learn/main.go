package main

import (
	"context"
	"fmt"
	"github.com/binarys-stars/my-deep-research/logs"
	"github.com/cloudwego/eino/compose"
)

func main() {
	ctx := context.Background()

	// 注册图
	graph := compose.NewGraph[string, string]()

	lambda0 := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		if input == "1" {
			return "毫猫", nil
		}
		if input == "2" {
			return "耄耋", nil
		}
		if input == "3" {
			return "device", nil
		}
		return "", fmt.Errorf("unknown input")
	})
	lambda1 := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		return "喵", nil
	})
	lambda2 := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		return "哈", nil
	})
	lambda3 := compose.InvokableLambda(func(ctx context.Context, input string) (string, error) {
		return "没有人类了", nil
	})
	err := graph.AddLambdaNode("lambda0", lambda0)
	if err != nil {
		logs.Errorf("add lambda node failed, err=%v", err)
		return
	}
	err = graph.AddLambdaNode("lambda1", lambda1)
	if err != nil {
		logs.Errorf("add lambda node failed, err=%v", err)
		return
	}
	err = graph.AddLambdaNode("lambda2", lambda2)
	if err != nil {
		logs.Errorf("add lambda node failed, err=%v", err)
		return
	}
	err = graph.AddLambdaNode("lambda3", lambda3)
	if err != nil {
		logs.Errorf("add lambda node failed, err=%v", err)
		return
	}
	err = graph.AddBranch("lambda0", compose.NewGraphBranch(func(ctx context.Context, input string) (endNode string, err error) {
		resMap := map[string]string{
			"毫猫":   "lambda1",
			"耄耋":   "lambda2",
			"device": "lambda3",
		}
		res, ok := resMap[input]
		if !ok {
			return compose.END, fmt.Errorf("unknown input")
		}
		return res, nil
	}, map[string]bool{
		"lambda1":   true,
		"lambda2":   true,
		"lambda3":   true,
		compose.END: true,
	}))
	if err != nil {
		logs.Errorf("add branch failed, err=%v", err)
		return
	}
	err = graph.AddEdge(compose.START, "lambda0")
	if err != nil {
		logs.Errorf("add edge failed, err=%v", err)
		return
	}
	err = graph.AddEdge("lambda1", compose.END)
	if err != nil {
		logs.Errorf("add edge failed, err=%v", err)
		return
	}
	err = graph.AddEdge("lambda2", compose.END)
	if err != nil {
		logs.Errorf("add edge failed, err=%v", err)
		return
	}
	err = graph.AddEdge("lambda3", compose.END)
	if err != nil {
		logs.Errorf("add edge failed, err=%v", err)
		return
	}
	// 运行图
	runnable, err := graph.Compile(ctx)
	res, err := runnable.Invoke(ctx, "3")
	if err != nil {
		logs.Errorf("compile graph failed, err=%v", err)
		return
	}
	fmt.Println(res)
}
