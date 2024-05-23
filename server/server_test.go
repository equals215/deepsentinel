package server

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/dashboard"
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	var payloadTestChan = make(chan *monitoring.Payload)
	var dashboardTestChan = make(chan *dashboard.Data)

	var testTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var testClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: testTransport,
	}
	config.Server = &config.ServerConfig{
		AuthToken: "test-auth-token",
	}

	go func() {
		for payload := range payloadTestChan {
			_ = payload
		}
	}()
	s := newServer(payloadTestChan, dashboardTestChan)
	// Test if the server is created
	assert.NotNil(t, s, "newServer() returned nil")

	// Test if the server is created with the correct name
	assert.Equal(t, "DeepSentinel API", s.Config().AppName, "newServer() returned a server with incorrect name")

	go s.Listen("localhost:8486")
	defer s.Shutdown()

	// Test if the server is running
	resp, err := testClient.Get("http://localhost:8486/health")
	assert.Nil(t, err, "Failed to send request to server")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Server returned incorrect status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Server returned incorrect content type")

	// Test content of the response to /health
	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.Nil(t, err, "Failed to decode response body")
	assert.Equal(t, "pass", body["status"], "Server returned incorrect response")

	// Test POST /probe/:machine/report
	payload := &monitoring.Payload{
		MachineStatus: "pass",
		Services:      nil,
	}
	payloadBytes, err := json.Marshal(payload)
	assert.Nil(t, err, "Failed to marshal payload")
	assert.Equal(t, `{"machineStatus":"pass","services":null}`, string(payloadBytes), "Failed to marshal payload correctly")
	req, err := http.NewRequest("POST", "http://localhost:8486/probe/testmachine/report", bytes.NewBuffer(payloadBytes))
	assert.Nil(t, err, "Failed to create POST request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "test-auth-token")
	resp, err = testClient.Do(req)
	assert.Nil(t, err, "Failed to send POST request to server")
	assert.Equal(t, http.StatusAccepted, resp.StatusCode, "Server returned incorrect status code for POST /probe/:machine/report")

	// Test DELETE /probe/:machine
	req, _ = http.NewRequest("DELETE", "http://localhost:8486/probe/testmachine", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "test-auth-token")
	resp, err = testClient.Do(req)
	assert.Nil(t, err, "Failed to send DELETE request to server")
	assert.Equal(t, http.StatusAccepted, resp.StatusCode, "Server returned incorrect status code for DELETE /probe/:machine")
}
