package config

// AlertProviderConfig is the interface for alert providers
type AlertProviderConfig interface {
	Type() AlertProviderType
}

type EmptyProvider struct{}

func (k *EmptyProvider) Type() AlertProviderType {
	return EmptyProviderType
}

// AlertProviderType is an iota type for different alert provider types
type AlertProviderType int

const (
	pagerDuty AlertProviderType = iota
	keepHQ
	EmptyProviderType
)

func (a AlertProviderType) String() string {
	return [...]string{"pagerduty", "keephq"}[a]
}
