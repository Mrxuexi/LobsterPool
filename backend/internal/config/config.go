package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port       string
	DBPath     string
	Namespace  string
	Kubeconfig string
	StaticDir  string
	DevMode    bool
	JWTSecret  string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("LP_PORT", "8080"),
		DBPath:     getEnv("LP_DB_PATH", "lobsterpool.db"),
		Namespace:  getEnv("LP_NAMESPACE", "lobsterpool"),
		Kubeconfig: getEnv("LP_KUBECONFIG", defaultKubeconfig()),
		StaticDir:  getEnv("LP_STATIC_DIR", "./static"),
		DevMode:    getEnv("LP_DEV_MODE", "false") == "true",
		JWTSecret:  getEnv("LP_JWT_SECRET", "lobsterpool-dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultKubeconfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
