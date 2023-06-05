package server

type Config struct {
	Address        string `env:"RUN_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	SecretFilePath string `env:"SECRET_FILE"`
}
