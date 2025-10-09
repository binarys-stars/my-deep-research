package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RanFeng/ilog"
	"github.com/binarys-stars/my-deep-research/conf"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"strings"
	"time"
)

const (
	transportStdio = "stdio"
	transportSSE   = "sse"
)

var (
	MCPServer map[string]client.MCPClient
)

func InitMCP() {
	var err error
	MCPServer, err = CreateMCPClients()
	if err != nil {
		panic(fmt.Sprintf("CreateMCPClients failed: %w", err))
	}
}

type MCPConfig struct {
	MCPServers map[string]ServerConfigWrapper `json:"MCPServers"`
}
type ServerConfig interface {
	GetType() string
}

type STDIOServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

type SSEServerConfig struct {
	Url     string   `json:"url"`
	Headers []string `json:"headers,omitempty"`
}

func (s STDIOServerConfig) GetType() string {
	return transportStdio
}
func (s SSEServerConfig) GetType() string {
	return transportSSE
}

type ServerConfigWrapper struct {
	Config ServerConfig
}

func (w *ServerConfigWrapper) UnmarshalJSON(data []byte) error {
	var typeField struct {
		Url string `json:"url"`
	}

	if err := json.Unmarshal(data, &typeField); err != nil {
		return err
	}
	if typeField.Url != "" {
		// If the URL field is present, treat it as an SSE server
		var sse SSEServerConfig
		if err := json.Unmarshal(data, &sse); err != nil {
			return err
		}
		w.Config = sse
	} else {
		// Otherwise, treat it as a STDIOServerConfig
		var stdio STDIOServerConfig
		if err := json.Unmarshal(data, &stdio); err != nil {
			return err
		}
		w.Config = stdio
	}

	return nil
}
func (w ServerConfigWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Config)
}

func CreateMCPClients() (map[string]client.MCPClient, error) {
	// 将 DeerConfig 转换为 MCPConfig
	mcpConfig := &MCPConfig{
		MCPServers: make(map[string]ServerConfigWrapper),
	}

	for name, server := range conf.Config.MCP.Servers {
		mcpConfig.MCPServers[name] = ServerConfigWrapper{
			Config: STDIOServerConfig{
				Command: server.Command,
				Args:    server.Args,
				Env:     server.Env,
			},
		}
	}

	clients := make(map[string]client.MCPClient)

	for name, server := range mcpConfig.MCPServers {
		var mcpClient client.MCPClient
		var err error
		ilog.EventInfo(context.Background(), "load mcp client", name, server.Config.GetType())
		if server.Config.GetType() == transportSSE {
			sseConfig := server.Config.(SSEServerConfig)

			options := []client.ClientOption{}

			if sseConfig.Headers != nil {
				// Parse headers from the conf
				headers := make(map[string]string)
				for _, header := range sseConfig.Headers {
					parts := strings.SplitN(header, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						headers[key] = value
					}
				}
				options = append(options, client.WithHeaders(headers))
			}

			mcpClient, err = client.NewSSEMCPClient(
				sseConfig.Url,
				options...,
			)
			if err == nil {
				err = mcpClient.(*client.SSEMCPClient).Start(context.Background())
			}
		} else {
			stdioConfig := server.Config.(STDIOServerConfig)
			var env []string
			for k, v := range stdioConfig.Env {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
			mcpClient, err = client.NewStdioMCPClient(
				stdioConfig.Command,
				env,
				stdioConfig.Args...)
		}
		if err != nil {
			for _, c := range clients {
				_ = c.Close()
			}
			return nil, fmt.Errorf(
				"failed to create MCP client for %s: %w",
				name,
				err,
			)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ilog.EventInfo(ctx, "Initializing server...", "name", name)
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = mcp.Implementation{
			Name:    "mcphost",
			Version: "0.1.0",
		}
		initRequest.Params.Capabilities = mcp.ClientCapabilities{}

		_, err = mcpClient.Initialize(ctx, initRequest)
		if err != nil {
			_ = mcpClient.Close()
			for _, c := range clients {
				_ = c.Close()
			}
			return nil, fmt.Errorf(
				"failed to initialize MCP client for %s: %w",
				name,
				err,
			)
		}

		clients[name] = mcpClient
	}

	return clients, nil
}
