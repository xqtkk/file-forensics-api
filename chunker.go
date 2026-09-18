package main

import (
	"context"
	"fmt"
)

// Event — упрощённое представление события для отправки в LLM
type Event struct {
	ID          int    `json:"id"`
	EventTime   string `json:"event_time"`
	EventName   string `json:"event_name"`
	UserName    string `json:"user_name"`
	SourceIP    string `json:"source_ip"`
	Region      string `json:"region"`
	EventSource string `json:"event_source"`
}

// Chunk — порция событий для одного запроса к LLM
type Chunk struct {
	Index  int     `json:"index"`
	Events []Event `json:"events"`
}

// LoadEvents загружает события из БД в заданном временном диапазоне
func LoadEvents(ctx context.Context) ([]Event, error) {
	rows, err := DB.Query(ctx,
		`SELECT id, event_time, event_name, COALESCE(user_name, ''),
		        COALESCE(source_ip, ''), COALESCE(region, ''),
		        COALESCE(event_source, '')
		 FROM events
		 ORDER BY event_time ASC`)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var t interface{}
		if err := rows.Scan(&e.ID, &t, &e.EventName, &e.UserName, &e.SourceIP, &e.Region, &e.EventSource); err != nil {
			continue
		}
		e.EventTime = fmt.Sprintf("%v", t)
		events = append(events, e)
	}
	return events, nil
}

// ChunkEvents разбивает события на порции по chunkSize
func ChunkEvents(events []Event, chunkSize int) []Chunk {
	if chunkSize <= 0 {
		chunkSize = 100
	}

	var chunks []Chunk
	for i := 0; i < len(events); i += chunkSize {
		end := i + chunkSize
		if end > len(events) {
			end = len(events)
		}
		chunks = append(chunks, Chunk{
			Index:  len(chunks),
			Events: events[i:end],
		})
	}
	return chunks
}