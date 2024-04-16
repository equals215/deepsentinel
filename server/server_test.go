package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/equals215/deepsentinel/config"
	"github.com/equals215/deepsentinel/monitoring"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	testChan := make(chan *monitoring.Payload)
	config.Server = &config.ServerConfig{
		AuthToken: "test-auth-token",
	}

	s := newServer(testChan)

	// Test if the server is created
	assert.NotNil(t, s, "newServer() returned nil")

	// Test if the server is created with the correct name
	assert.Equal(t, "DeepSentinel API", s.Config().AppName, "newServer() returned a server with incorrect name")

	go s.Listen("localhost:8486")
	defer s.Shutdown()

	// Test if the server is running
	resp, err := http.Get("http://localhost:8486/health")
	assert.Nil(t, err, "Failed to send request to server")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Server returned incorrect status code")
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Server returned incorrect content type")

	// Test content of the response to /health
	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.Nil(t, err, "Failed to decode response body")
	assert.Equal(t, "pass", body["status"], "Server returned incorrect response")

	// Test POST /probe/:machine/report
	// payload := &monitoring.Payload{
	// 	MachineStatus: "pass",
	// }
	// payloadBytes, err := json.Marshal(payload)
	// assert.Nil(t, err, "Failed to marshal payload")
	// assert.Equal(t, `{"MachineStatus":"pass"}`, string(payloadBytes), "Failed to marshal payload correctly")
	// req, err := http.NewRequest("POST", "http://localhost:8486/probe/testmachine/report", bytes.NewBuffer(payloadBytes))
	// assert.Nil(t, err, "Failed to create POST request")
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Authorization", "test-auth-token")
	// resp, err = http.DefaultClient.Do(req)
	// assert.Nil(t, err, "Failed to send POST request to server")
	// assert.Equal(t, http.StatusAccepted, resp.StatusCode, "Server returned incorrect status code for POST /probe/:machine/report")

	// Test DELETE /probe/:machine
	// req, _ = http.NewRequest("DELETE", "http://localhost:8486/probe/test-machine", nil)
	// resp, err = http.DefaultClient.Do(req)
	// assert.Nil(t, err, "Failed to send DELETE request to server")
	// assert.Equal(t, http.StatusAccepted, resp.StatusCode, "Server returned incorrect status code for DELETE /probe/:machine")
}
