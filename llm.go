package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ollamaURL = "http://host.docker.internal:11434/api/generate"

// OllamaRequest — запрос к Ollama
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

// OllamaResponse — ответ от Ollama
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// AnalyzeChunk отправляет чанк в LLM и возвращает текстовый анализ
func AnalyzeChunk(ctx context.Context, chunk Chunk) (string, error) {
	// Формируем промпт
	prompt := buildPrompt(chunk)

	reqBody := OllamaRequest{
		Model: "llama3.1:8b",
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Minute}

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка запроса к Ollama: %w", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	return ollamaResp.Response, nil
}

// buildPrompt формирует инструкцию для LLM
func buildPrompt(chunk Chunk) string {
	eventsJSON, _ := json.MarshalIndent(chunk.Events, "", "  ")

	return fmt.Sprintf(`Ты — аналитик DFIR. Проанализируй события AWS CloudTrail.

ЗАДАЧА: найти IP-адрес и пользователя, которые НЕ являются обычными для этого лога.

МЕТОД:
1. Посчитай, сколько раз встречается каждый IP-адрес.
2. Посчитай, сколько раз встречается каждый пользователь.
3. IP или пользователь, который встречается РЕЖЕ ВСЕГО — подозрительный.
4. Если есть события CreateUser, CreateAccessKey, AttachUserPolicy от одного пользователя — это атака.

События (чанк #%d):
%s

Ответ СТРОГО в JSON без пояснений:
{
  "suspicious_ips": ["ip1", "ip2"],
  "suspicious_users": ["user1"],
  "suspicious_events": [
    {"event_name": "...", "user_name": "...", "source_ip": "...", "reason": "..."}
  ],
  "attack_stage": "...",
  "summary": "..."
}`, chunk.Index, string(eventsJSON))
}

// extractJSON вытаскивает JSON-объект из ответа модели.
// Если модель добавила текст до/после — отрезает его.
func extractJSON(s string) string {
	start := -1
	end := -1
	depth := 0

	for i, ch := range s {
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}

	if start == -1 || end == -1 {
		return s // не нашли — возвращаем как есть
	}

	return s[start : end+1]
}