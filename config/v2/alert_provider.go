package config

// AlertProviderConfig is the interface for alert providers
type AlertProviderConfig interface {
	Type() AlertProviderType
}

// AlertProviderType is an iota type for different alert provider types
type AlertProviderType int

const (
	pagerDuty AlertProviderType = iota
	keepHQ
)

func (a AlertProviderType) String() string {
	return [...]string{"pagerduty", "keephq"}[a]
}
