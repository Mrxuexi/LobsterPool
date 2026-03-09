package main

import (
	"testing"

	"github.com/lobsterpool/lobsterpool/internal/config"
)

func TestBootstrapAdminPassword(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
		want string
	}{
		{
			name: "explicit bootstrap password wins",
			cfg: config.Config{
				DevMode:            true,
				BootstrapAdminPass: "custom-secret",
			},
			want: "custom-secret",
		},
		{
			name: "dev mode falls back to admin",
			cfg: config.Config{
				DevMode: true,
			},
			want: "admin",
		},
		{
			name: "non-dev mode has no implicit bootstrap password",
			cfg:  config.Config{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bootstrapAdminPassword(&tt.cfg); got != tt.want {
				t.Fatalf("bootstrapAdminPassword() = %q, want %q", got, tt.want)
			}
		})
	}
}
