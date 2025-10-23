// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package ipc

// Request represents a command sent over the IPC channel.
type Request struct {
	Command string            `json:"command"`
	Params  map[string]string `json:"params,omitempty"`
}

// Response represents the result of handling an IPC command.
type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Handler defines the signature for IPC command handlers.
type Handler func(req Request) (Response, error)

// NewSuccessResponse constructs a successful response payload.
func NewSuccessResponse(message string, data any) Response {
	return Response{
		OK:      true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse constructs an error response payload.
func NewErrorResponse(message string) Response {
	return Response{
		OK:      false,
		Message: message,
	}
}
