package main

import (
	"context"
	"encoding/json"
	"time"
)

// CloudTrailEvent — упрощённая структура события AWS CloudTrail
type CloudTrailEvent struct {
	EventTime     string `json:"eventTime"`
	EventName     string `json:"eventName"`
	EventSource   string `json:"eventSource"`
	AWSRegion     string `json:"awsRegion"`
	SourceIP      string `json:"sourceIPAddress"`
	UserIdentity  struct {
		Type     string `json:"type"`
		UserName string `json:"userName"`
	} `json:"userIdentity"`
}

// NormalizedEvent — нормализованное событие для хранения в БД
type NormalizedEvent struct {
	EventTime   time.Time
	EventName   string
	UserName    string
	SourceIP    string
	Region      string
	EventSource string
	Raw         json.RawMessage
}

// NormalizeCloudTrail превращает CloudTrail-событие в NormalizedEvent
func NormalizeCloudTrail(raw json.RawMessage) (*NormalizedEvent, error) {
	var ct CloudTrailEvent
	if err := json.Unmarshal(raw, &ct); err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, ct.EventTime)
	if err != nil {
		return nil, err
	}

	return &NormalizedEvent{
		EventTime:   t,
		EventName:   ct.EventName,
		UserName:    ct.UserIdentity.UserName,
		SourceIP:    ct.SourceIP,
		Region:      ct.AWSRegion,
		EventSource: ct.EventSource,
		Raw:         raw,
	}, nil
}

// SaveEvent сохраняет нормализованное событие в БД
func SaveEvent(ctx context.Context, e *NormalizedEvent) error {
	_, err := DB.Exec(ctx,
		`INSERT INTO events
		 (event_time, event_name, user_name, source_ip, region, event_source, raw)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.EventTime, e.EventName, e.UserName, e.SourceIP, e.Region, e.EventSource, e.Raw,
	)
	return err
}