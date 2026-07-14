package app

import (
	"net/url"
	"testing"
)

func TestDatabaseURLFromPostgresEnvironment(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_HOST", "postgres")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "saaskit_test")
	t.Setenv("POSTGRES_USER", "owner")
	t.Setenv("POSTGRES_PASSWORD", "p@ss:word/with#symbols")
	t.Setenv("POSTGRES_SSLMODE", "require")

	parsed, err := url.Parse(databaseURL())
	if err != nil {
		t.Fatal(err)
	}
	password, _ := parsed.User.Password()
	if parsed.Host != "postgres:5433" || parsed.Path != "/saaskit_test" || parsed.User.Username() != "owner" || password != "p@ss:word/with#symbols" || parsed.Query().Get("sslmode") != "require" {
		t.Fatalf("unexpected database URL: %s", parsed.String())
	}
}

func TestExplicitDatabaseURLOverridesPostgresEnvironment(t *testing.T) {
	want := "postgres://cloud-user:secret@db.example.com:5432/cloud?sslmode=verify-full"
	t.Setenv("DATABASE_URL", want)
	if got := databaseURL(); got != want {
		t.Fatalf("want explicit DATABASE_URL %q, got %q", want, got)
	}
}
