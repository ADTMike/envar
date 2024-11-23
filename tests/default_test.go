package tests

import (
	"os"
	"testing"

	"github.com/ADTMike/envar"
)

var Cfg struct {
	PostgresUrl string `env:"POSTGRES_URL"`
}

func TestBind(t *testing.T) {
	if err := envar.Bind(&Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestAssertion(t *testing.T) {
	expected := os.Getenv("POSTGRES_URL")
	if Cfg.PostgresUrl != expected {
		t.Fatalf(`Cfg.PostgresUrl = %q, want %q`, Cfg.PostgresUrl, expected)
	}
}
