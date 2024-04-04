package agent

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/equals215/deepsentinel/config"
)

func reportPanic() {}

func reportWatcherDied() {}

func reportUnregisterAgent() error {
	if config.Client.MachineName == "" {
		return fmt.Errorf("machine name not set")
	}
	rawURL := fmt.Sprintf("%s/probe/%s", config.Client.ServerAddress, config.Client.MachineName)
	// Parse the server address URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("error parsing server address: %v", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http://"
	}

	// Send a DELETE HTTP request to parsedURL
	req, err := http.NewRequest("DELETE", parsedURL.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating DELETE request: %v", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", config.Client.AuthToken)

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
