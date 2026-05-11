package api

import (
	"fmt"

	"connectcli/internal/credentials"
)

// ShiftDeletionClient is a placeholder client for delete-shift wiring.
// The endpoint implementation is not available in this codebase yet.
type ShiftDeletionClient struct{}

type ShiftDeletionResponse struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	RequestID     string `json:"requestId"`
	ServerVersion string `json:"serverVersion"`
}

func NewShiftDeletionClient() *ShiftDeletionClient {
	return &ShiftDeletionClient{}
}

func (c *ShiftDeletionClient) DeleteShift(_ *credentials.Credentials, _ int, _ string) (*ShiftDeletionResponse, error) {
	return nil, fmt.Errorf("delete shift API is not implemented")
}
