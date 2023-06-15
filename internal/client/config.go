package client

// Config is the client configuration.
type Config struct {
	ServerAddress       string `json:"server_address"`
	GrpcTimeout         int    `json:"timeout_sec"`
	FlushIntoFileAfter  int    `json:"flush_into_file_sec"`
	SyncWithServerAfter int    `json:"sync_with_server_sec"`
	LogFilePath         string `json:"log_file"`
	CacheFilePath       string `json:"cache_file"`
	PutsFilePath        string `json:"puts_file"`
	LoggedUserFilePath  string `json:"logged_user_file"`
}
