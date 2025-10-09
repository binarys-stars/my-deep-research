package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/RanFeng/ilog"
	"github.com/binarys-stars/my-deep-research/biz/consts"
	"github.com/binarys-stars/my-deep-research/biz/eino"
	"github.com/binarys-stars/my-deep-research/biz/infra"
	"github.com/binarys-stars/my-deep-research/biz/model"
	"github.com/binarys-stars/my-deep-research/conf"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"os"
	"strings"
	"time"
)

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ilog.LogLevelKey, ilog.LevelInfo)
	conf.LoadDeerConfig(ctx)
	infra.InitModel()
	infra.InitMCP()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入你的需求： ")
	userPrompt, _ := reader.ReadString('\n')
	userPrompt = strings.TrimSpace(userPrompt) // 去除换行符
	userMessage := []*schema.Message{
		schema.UserMessage(userPrompt),
	}

	genFunc := func(ctx context.Context) *model.State {
		return &model.State{
			MaxPlanIterations: conf.Config.Setting.MaxPlanIterations,
			AutoAcceptedPlan:  true,
			MaxStepNum:        conf.Config.Setting.MaxStepNum,
			Messages:          userMessage,
			Goto:              consts.Coordinator,
		}
	}
	/*	// 1.调用调试服务初始化函数
		err := devops.Init(ctx)
		if err != nil {
			logs.Errorf("[eino dev] init failed, err=%v", err)
		}
	*/
	// 2.编译目标调试的编排产物

	r := eino.Builder[string, string, *model.State](ctx, genFunc)
	/*	sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
	*/
	outChan := make(chan string)
	go func() {
		for out := range outChan {
			fmt.Print(out)
		}
	}()

	_, err := r.Stream(ctx,
		consts.Coordinator,
		compose.WithCallbacks(&infra.LoggerCallback{
			Out: outChan,
		}),
	)
	if err != nil {
		ilog.EventError(ctx, err, "run failed")
	}
	close(outChan)
	ilog.EventInfo(ctx, "run console finish", time.Now())

}
