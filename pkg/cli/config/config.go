package config

type Config struct {
	PublicKey  string
	PrivateKey string
	Images     []string
}

func Read() (*Config, error) {

	return nil, nil
}

func Create(conf *Config) error {
	return nil
}
