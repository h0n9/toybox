package provider

type Provider interface {
	GetSecretValue(string) (string, error)
	GetAndSaveSecretValueToFile(string, string) error
}
