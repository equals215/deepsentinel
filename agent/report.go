package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/equals215/deepsentinel/config/v1"
)

func reportPanic() {}

func reportWatcherDied() {}

func reportUnregisterAgent() error {
	if config.Agent.MachineName == "" {
		return fmt.Errorf("machine name not set")
	}
	rawURL := fmt.Sprintf("%s/probe/%s", config.Agent.ServerAddress, config.Agent.MachineName)
	// Parse the server address URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("error parsing server address: %v", err)
	}

	// Send a DELETE HTTP request to parsedURL
	req, err := http.NewRequest("DELETE", parsedURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating DELETE request: %v", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", config.Agent.AuthToken)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending DELETE request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	return nil
}

func reportAlive() error {
	if config.Agent.MachineName == "" {
		return fmt.Errorf("machine name not set")
	}
	rawURL := fmt.Sprintf("%s/probe/%s/report", config.Agent.ServerAddress, config.Agent.MachineName)
	// Parse the server address URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("error parsing server address: %v", err)
	}

	// Send a GET HTTP request to parsedURL
	req, err := http.NewRequest("POST", parsedURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating POST request: %v", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", config.Agent.AuthToken)

	// Add JSON body
	body := []byte(`{"machineStatus":"pass"}`)
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	return nil
}
