package main

import (
	"encoding/json"
	"errors"
	"math"
	"time"
)

var messageCache []EditMessage

type EditMessage struct {
	Timestamp  time.Time `json:"interval"`
	ChangeSize int `json:"change_size"`
}

func (e *EditMessage) UnmarshalJSON(jsonBytes []byte) error {
	var msg map[string]*json.RawMessage
	if err := json.Unmarshal(jsonBytes, &msg); err != nil {
		return err
	}
	var action string
	var changeSize int

	if msg["action"] == nil || msg["change_size"] == nil {
		return errors.New("keys 'action' and 'change_size' are required")
	}

	if err := json.Unmarshal(*msg["action"], &action); err != nil {
		return err
	}

	if action != "edit" {
		return errors.New("key 'action' is not of value 'edit'")
	}

	if err := json.Unmarshal(*msg["change_size"], &changeSize); err != nil {
		return err
	}

	*e = EditMessage{
		ChangeSize: int(math.Abs(float64(changeSize))),
		Timestamp:  time.Now(),
	}

	return nil
}
