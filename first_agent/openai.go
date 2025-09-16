package main

import (
	"context"
	"log"
	"os"
	"runtime"
)

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

package main

import (
"context"
"github.com/cloudwego/eino-examples/internal/logs"
"github.com/cloudwego/eino-ext/components/model/openai"
"github.com/cloudwego/eino/components/model"
"github.com/joho/godotenv"
"log"
"os"
"runtime"
)

func createOpenAIChatModel(ctx context.Context) model.ToolCallingChatModel {
	switch runtime.GOOS {
	case "windows":
		err := godotenv.Load("quickstart\\funAgent\\.env")
		if err != nil {
			log.Printf("Error loading .env file, err=%v", err)
		}
	case "darwin":
		err := godotenv.Load("quickstart/funAgent/.env")
		if err != nil {
			log.Printf("Error loading .env file, err=%v", err)
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
