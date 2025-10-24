// Copyright (c) 2025 Asher Buk
// SPDX-License-Identifier: MIT

package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	defaultDialTimeout = 3 * time.Second
)

// SendRequest connects to the given socket path, sends the request, and waits for a response.
func SendRequest(socketPath string, req Request, timeout time.Duration) (Response, error) {
	if socketPath == "" {
		return Response{}, fmt.Errorf("ipc socket path not provided")
	}

	if timeout <= 0 {
		timeout = defaultDialTimeout
	}

	conn, err := net.DialTimeout("unix", socketPath, timeout)
	if err != nil {
		return Response{}, fmt.Errorf("failed to connect to ipc server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return Response{}, fmt.Errorf("failed to set ipc deadline: %w", err)
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return Response{}, fmt.Errorf("failed to encode request: %w", err)
	}
	payload = append(payload, '\n')

	if _, err := conn.Write(payload); err != nil {
		return Response{}, fmt.Errorf("failed to send request: %w", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return Response{}, fmt.Errorf("failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(line, &resp); err != nil {
		return Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if !resp.OK {
		if resp.Message != "" {
			return resp, fmt.Errorf("%s", resp.Message)
		}
		return resp, fmt.Errorf("command failed")
	}

	return resp, nil
}
