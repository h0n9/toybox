package provider

type Provider interface {
	GetSecretValue(string) (string, error)
}
