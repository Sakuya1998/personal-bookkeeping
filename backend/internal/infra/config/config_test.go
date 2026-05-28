package config

import (
	"testing"
)

func TestDSN(t *testing.T) {
	c := &Config{
		DB: DBConfig{
			Host:     "myhost",
			Port:     "5432",
			User:     "myuser",
			Password: "mypass",
			Name:     "mydb",
			SSLMode:  "require",
		},
	}
	dsn := c.DSN()
	want := "host=myhost port=5432 user=myuser password=mypass dbname=mydb sslmode=require"
	if dsn != want {
		t.Fatalf("DSN: got %q, want %q", dsn, want)
	}
}

func TestL1Duration_Explicit(t *testing.T) {
	c := &CacheConfig{
		L1TTL: 60,
		TTL:   300,
	}
	if d := c.L1Duration(); d != 60 {
		t.Fatalf("expected 60, got %d", d)
	}
}

func TestL1Duration_FromTTL(t *testing.T) {
	c := &CacheConfig{
		L1TTL: 0,
		TTL:   600,
	}
	if d := c.L1Duration(); d != 60 {
		t.Fatalf("expected 60 (600/10), got %d", d)
	}
}

func TestL1Duration_SmallTTL(t *testing.T) {
	c := &CacheConfig{
		L1TTL: 0,
		TTL:   30,
	}
	if d := c.L1Duration(); d != 30 {
		t.Fatalf("expected 30 (min default), got %d", d)
	}
}

func TestL1Duration_ZeroTTL(t *testing.T) {
	c := &CacheConfig{
		L1TTL: 0,
		TTL:   0,
	}
	if d := c.L1Duration(); d != 30 {
		t.Fatalf("expected 30 (default), got %d", d)
	}
}

func TestDSN_Defaults(t *testing.T) {
	c := &Config{}
	dsn := c.DSN()
	// Should not panic, should produce some valid-ish string
	if dsn == "" {
		t.Fatal("DSN should not be empty")
	}
}
