package handler

type SecretHandler func(string) (string, error)

func HandleSecretValue(secretValue string, handler SecretHandler) error {
	return nil
}