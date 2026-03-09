package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("LP_PORT", "")
	t.Setenv("LP_DB_PATH", "")
	t.Setenv("LP_NAMESPACE", "")
	t.Setenv("LP_KUBECONFIG", "")
	t.Setenv("LP_KUBE_CLUSTERS", "")
	t.Setenv("LP_DEFAULT_CLUSTER", "")
	t.Setenv("LP_STATIC_DIR", "")
	t.Setenv("LP_DEV_MODE", "")
	t.Setenv("LP_JWT_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
	if cfg.DBPath != "lobsterpool.db" {
		t.Fatalf("expected default db path, got %q", cfg.DBPath)
	}
	if cfg.Namespace != "lobsterpool" {
		t.Fatalf("expected default namespace, got %q", cfg.Namespace)
	}
	if cfg.StaticDir != "./static" {
		t.Fatalf("expected default static dir, got %q", cfg.StaticDir)
	}
	if cfg.DevMode {
		t.Fatalf("expected default dev mode false")
	}
	if cfg.JWTSecret != "lobsterpool-dev-secret-change-me" {
		t.Fatalf("expected default jwt secret, got %q", cfg.JWTSecret)
	}
	if cfg.Kubeconfig == "" {
		t.Fatalf("expected default kubeconfig to be set")
	}
	if !cfg.LegacySingleCluster {
		t.Fatalf("expected legacy single cluster mode by default")
	}
	if cfg.DefaultCluster != defaultClusterName {
		t.Fatalf("expected default cluster %q, got %q", defaultClusterName, cfg.DefaultCluster)
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	t.Setenv("LP_PORT", "18080")
	t.Setenv("LP_DB_PATH", "/tmp/lp.db")
	t.Setenv("LP_NAMESPACE", "test-ns")
	t.Setenv("LP_KUBECONFIG", "/tmp/kubeconfig")
	t.Setenv("LP_STATIC_DIR", "/tmp/static")
	t.Setenv("LP_DEV_MODE", "true")
	t.Setenv("LP_JWT_SECRET", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Port != "18080" ||
		cfg.DBPath != "/tmp/lp.db" ||
		cfg.Namespace != "test-ns" ||
		cfg.Kubeconfig != "/tmp/kubeconfig" ||
		cfg.StaticDir != "/tmp/static" ||
		!cfg.DevMode ||
		cfg.JWTSecret != "secret" {
		t.Fatalf("unexpected config override result: %+v", cfg)
	}
}

func TestLoad_MultiClusterConfig(t *testing.T) {
	t.Setenv("LP_NAMESPACE", "fallback-ns")
	t.Setenv("LP_KUBECONFIG", "/tmp/default-kubeconfig")
	t.Setenv("LP_DEFAULT_CLUSTER", "kind-dev")
	t.Setenv("LP_KUBE_CLUSTERS", `[{"name":"kind-dev","display_name":"Kind Dev","namespace":"lobsterpool-local","kubeconfig":"/tmp/kind-config","context":"kind-lobsterpool-dev"},{"name":"remote-a","api_server":"https://10.0.0.10:6443","token":"token-a"}]`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.LegacySingleCluster {
		t.Fatalf("expected multi cluster mode")
	}
	if len(cfg.KubeClusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(cfg.KubeClusters))
	}
	if cfg.KubeClusters[1].Namespace != "fallback-ns" {
		t.Fatalf("expected namespace fallback, got %q", cfg.KubeClusters[1].Namespace)
	}
	if cfg.DefaultCluster != "kind-dev" {
		t.Fatalf("expected default cluster to be kind-dev, got %q", cfg.DefaultCluster)
	}
}
