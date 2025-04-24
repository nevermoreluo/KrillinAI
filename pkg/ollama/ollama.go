package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"krillin-ai/log"
	"net/http"

	goopenai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// Ollama 客户端结构体
type OllamaClient struct {
	apiURL string // Ollama API 地址（默认 http://localhost:11434/api/generate）
	model  string // 使用的模型名称（如 "llama2"、"mistral"）
}

// 新建 Ollama 客户端
func NewOllamaClient(model string, apiURL ...string) *OllamaClient {
	// 设置默认 API 地址
	baseURL := "http://localhost:11434/api/generate"
	if len(apiURL) > 0 && apiURL[0] != "" {
		baseURL = apiURL[0]
	}
	return &OllamaClient{
		apiURL: baseURL,
		model:  model,
	}
}

// 实现聊天完成接口
func (c *OllamaClient) ChatCompletion(query string) (string, error) {
	// 构造请求体
	reqBody, err := json.Marshal(map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{
				"role":    goopenai.ChatMessageRoleSystem,
				"content": "You are an assistant that helps with subtitle translation.",
			}, {
				"role":    goopenai.ChatMessageRoleUser,
				"content": query,
			},
		},
		// 可选参数（可根据需求添加 temperature、top_p 等）
		// "temperature": 0.7,
		// "stream":      false,
	})
	if err != nil {
		log.GetLogger().Error("ollama marshal request failed", zap.Error(err))
		return "", err
	}

	// 发送 POST 请求
	resp, err := http.Post(c.apiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.GetLogger().Error("ollama http request failed", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Response string `json:"response"` // Ollama 响应内容
		Error    string `json:"error"`    // 错误信息
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.GetLogger().Error("ollama decode response failed", zap.Error(err))
		return "", err
	}

	if result.Error != "" {
		log.GetLogger().Error("ollama api error", zap.String("error", result.Error))
		return "", fmt.Errorf(result.Error)
	}

	return result.Response, nil
}
