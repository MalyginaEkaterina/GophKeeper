package server

type Config struct {
	Address             string `env:"RUN_ADDRESS"`
	DatabaseURI         string `env:"DATABASE_URI"`
	FileUserStoragePath string `env:"FILE_USER_STORAGE_PATH"`
	FileDataStoragePath string `env:"FILE_DATA_STORAGE_PATH"`
	SecretFilePath      string `env:"SECRET_FILE"`
}
