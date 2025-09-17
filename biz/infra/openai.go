package infra

import (
	"context"
	"github.com/binarys-stars/my-deep-research/logs"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/joho/godotenv"
	"log"
	"os"
	"runtime"
)

func CreateOpenAIChatModel(ctx context.Context) model.ToolCallingChatModel {
	switch runtime.GOOS {
	case "windows":
		err := godotenv.Load("biz/infra/.env")
		if err != nil {
			logs.Errorf("Error loading .env file, err=%v", err)
		}
	case "darwin":
		err := godotenv.Load("biz/infra/.env")
		if err != nil {
			logs.Errorf("Error loading .env file, err=%v", err)
		}
	default:
		logs.Infof("unsupported platform")
	}

	key := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL")

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		Model:   modelName,
		APIKey:  key,
	})
	if err != nil {
		log.Fatalf("create openai chat model failed, err=%v", err)
	}

	return chatModel
}
