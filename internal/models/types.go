package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap is a map that can be stored as JSONB in PostgreSQL
type JSONMap map[string]interface{}

// Value implements driver.Valuer for database storage
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner for database retrieval
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// JSONStringArray is a string array that can be stored as JSONB in PostgreSQL
type JSONStringArray []string

// Value implements driver.Valuer for database storage
func (j JSONStringArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner for database retrieval
func (j *JSONStringArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// AIStatus represents the status of AI processing
type AIStatus string

const (
	AIStatusPending AIStatus = "pending"
	AIStatusRunning AIStatus = "running"
	AIStatusDone    AIStatus = "done"
	AIStatusError   AIStatus = "error"
)

// ActionType represents the type of next action
type ActionType string

const (
	ActionTypeDialog ActionType = "dialog"
	ActionTypeBooks  ActionType = "books"
	ActionTypeVideos ActionType = "videos"
)
