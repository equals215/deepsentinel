package config

type KeepHQConfig struct {
}

func (k *KeepHQConfig) Type() AlertProviderType {
	return keepHQ
}
