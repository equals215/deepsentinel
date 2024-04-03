package config

func ClientSetServerAddress(Address string) {
	Client.Lock()
	Client.ServerAddress = Address
	Client.Unlock()
}
