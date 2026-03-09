package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultClusterName = "default"

type KubernetesCluster struct {
	Name                  string `json:"name"`
	DisplayName           string `json:"display_name"`
	Namespace             string `json:"namespace"`
	Kubeconfig            string `json:"kubeconfig"`
	Context               string `json:"context"`
	APIServer             string `json:"api_server"`
	Token                 string `json:"token"`
	CAFile                string `json:"ca_file"`
	InsecureSkipTLSVerify bool   `json:"insecure_skip_tls_verify"`
}

type Config struct {
	Port                string
	DBPath              string
	Namespace           string
	Kubeconfig          string
	KubeClusters        []KubernetesCluster
	DefaultCluster      string
	LegacySingleCluster bool
	StaticDir           string
	DevMode             bool
	JWTSecret           string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:       getEnv("LP_PORT", "8080"),
		DBPath:     getEnv("LP_DB_PATH", "lobsterpool.db"),
		Namespace:  getEnv("LP_NAMESPACE", "lobsterpool"),
		Kubeconfig: getEnv("LP_KUBECONFIG", defaultKubeconfig()),
		StaticDir:  getEnv("LP_STATIC_DIR", "./static"),
		DevMode:    getEnv("LP_DEV_MODE", "false") == "true",
		JWTSecret:  getEnv("LP_JWT_SECRET", "lobsterpool-dev-secret-change-me"),
	}

	rawClusters := strings.TrimSpace(os.Getenv("LP_KUBE_CLUSTERS"))
	if rawClusters == "" {
		cfg.DefaultCluster = getEnv("LP_DEFAULT_CLUSTER", defaultClusterName)
		cfg.LegacySingleCluster = true
		cfg.KubeClusters = []KubernetesCluster{
			{
				Name:        cfg.DefaultCluster,
				DisplayName: cfg.DefaultCluster,
				Namespace:   cfg.Namespace,
				Kubeconfig:  cfg.Kubeconfig,
			},
		}
		return cfg, nil
	}

	var clusters []KubernetesCluster
	if err := json.Unmarshal([]byte(rawClusters), &clusters); err != nil {
		return nil, fmt.Errorf("parse LP_KUBE_CLUSTERS: %w", err)
	}
	if len(clusters) == 0 {
		return nil, fmt.Errorf("LP_KUBE_CLUSTERS must contain at least one cluster")
	}

	cfg.DefaultCluster = getEnv("LP_DEFAULT_CLUSTER", strings.TrimSpace(clusters[0].Name))
	seen := make(map[string]struct{}, len(clusters))
	for i := range clusters {
		clusters[i].Name = strings.TrimSpace(clusters[i].Name)
		clusters[i].DisplayName = strings.TrimSpace(clusters[i].DisplayName)
		clusters[i].Namespace = strings.TrimSpace(clusters[i].Namespace)
		clusters[i].Kubeconfig = strings.TrimSpace(clusters[i].Kubeconfig)
		clusters[i].Context = strings.TrimSpace(clusters[i].Context)
		clusters[i].APIServer = strings.TrimSpace(clusters[i].APIServer)
		clusters[i].Token = strings.TrimSpace(clusters[i].Token)
		clusters[i].CAFile = strings.TrimSpace(clusters[i].CAFile)

		if clusters[i].Name == "" {
			return nil, fmt.Errorf("cluster %d has empty name", i)
		}
		if _, exists := seen[clusters[i].Name]; exists {
			return nil, fmt.Errorf("duplicate cluster name %q", clusters[i].Name)
		}
		seen[clusters[i].Name] = struct{}{}

		if clusters[i].DisplayName == "" {
			clusters[i].DisplayName = clusters[i].Name
		}
		if clusters[i].Namespace == "" {
			clusters[i].Namespace = cfg.Namespace
		}
		if clusters[i].APIServer == "" && clusters[i].Kubeconfig == "" {
			clusters[i].Kubeconfig = cfg.Kubeconfig
		}
	}

	if _, ok := seen[cfg.DefaultCluster]; !ok {
		return nil, fmt.Errorf("LP_DEFAULT_CLUSTER %q is not present in LP_KUBE_CLUSTERS", cfg.DefaultCluster)
	}

	cfg.KubeClusters = clusters
	return cfg, nil
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
