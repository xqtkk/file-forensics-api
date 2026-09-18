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
		Model:  "llama3.2:3b",
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Minute}

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

	return fmt.Sprintf(`Ты — аналитик по цифровой криминалистике. Проанализируй следующие события из облачных логов (AWS CloudTrail).

Задача:
1. Найди подозрительные события или паттерны.
2. Определи возможную стадию атаки (разведка, эксплуатация, закрепление, эскалация привилегий, эксфильтрация).
3. Опиши, что могло произойти.

События (чанк #%d):
%s

Ответ верни СТРОГО в формате JSON:
{
  "suspicious_events": [
    {"event_name": "...", "reason": "..."}
  ],
  "attack_stage": "название стадии или none",
  "summary": "краткое описание на русском"
}`, chunk.Index, string(eventsJSON))
}