package config

import "fmt"

// ClientSetServerAddress sets the server address in the client config
func ClientSetServerAddress(args ...any) error {
	// Client == nil means that CLI is doing the config change, not the running daemon
	if Client == nil {
		InitClient()
	}

	if len(args) == 0 {
		return fmt.Errorf("missing address")
	} else if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	address := args[0].(string)

	Client.Lock()
	Client.ServerAddress = address
	Client.saveToFile("/etc/deepsentinel/client-config.json")
	Client.Unlock()
	return nil
}
