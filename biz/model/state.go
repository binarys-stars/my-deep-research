package model

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RanFeng/ilog"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func init() {
	ctx := context.Background()
	err := compose.RegisterSerializableType[State]("DeerState")
	if err != nil {
		ilog.EventError(ctx, err, "RegisterSerializableType")
		panic(err)
	}
}

type State struct {
	// 用户输入的信息
	Messages []*schema.Message `json:"messages,omitempty"`

	// 子图共享变量
	Goto                           string `json:"goto,omitempty"`
	CurrentPlan                    *Plan  `json:"current_plan,omitempty"`
	Locale                         string `json:"locale,omitempty"`
	PlanIterations                 int    `json:"plan_iterations,omitempty"`
	BackgroundInvestigationResults string `json:"background_investigation_results"`
	InterruptFeedback              string `json:"interrupt_feedback,omitempty"`

	// 全局配置变量
	MaxPlanIterations             int  `json:"max_plan_iterations,omitempty"`
	MaxStepNum                    int  `json:"max_step_num,omitempty"`
	AutoAcceptedPlan              bool `json:"auto_accepted_plan"`
	EnableBackgroundInvestigation bool `json:"enable_background_investigation"`
}

func (s *State) MarshalJSON() ([]byte, error) {
	ctx := context.Background()
	res, err := json.Marshal(s)
	if err != nil {
		ilog.EventError(ctx, err, "MarshalJSON")
		return nil, err
	}
	return res, nil
}

func (s *State) UnmarshalJSON(data []byte) error {
	ctx := context.Background()
	type Alias State
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		ilog.EventError(ctx, err, "UnmarshalJSON")
		return err
	}
	*s = State(tmp)
	return nil
}

// DeerCheckPoint 全局状态存储点
type DeerCheckPoint struct {
	buf map[string][]byte
}

func (d *DeerCheckPoint) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	if d.buf == nil {
		d.buf = make(map[string][]byte)
	}
	if v, ok := d.buf[checkPointID]; ok {
		return v, true, nil
	}
	return nil, false, fmt.Errorf("key not found")
}

func (d *DeerCheckPoint) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	if d.buf == nil {
		d.buf = make(map[string][]byte)
	}
	d.buf[checkPointID] = checkPoint
	return nil
}

var deerCheckPoint = DeerCheckPoint{
	buf: make(map[string][]byte),
}

func NewDeerCheckPoint(ctx context.Context) compose.CheckPointStore {
	return &deerCheckPoint
}
