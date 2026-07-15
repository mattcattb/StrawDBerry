package main

import (
	"os"
	"testing"

	"go-redis/redis"
)

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"REDIS_BIND",
		"REDIS_PORT",
		"REDIS_APPENDONLY",
		"REDIS_APPENDFILENAME",
		"REDIS_APPENDFSYNC",
	} {
		value, ok := os.LookupEnv(name)
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(name, value)
			} else {
				_ = os.Unsetenv(name)
			}
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.server.Bind != "0.0.0.0" || cfg.server.Port != 6479 || !cfg.aof.Enabled || cfg.aof.FSyncPolicy != redis.FsAlways {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("REDIS_BIND", "127.0.0.1")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_APPENDONLY", "no")
	t.Setenv("REDIS_APPENDFILENAME", "/tmp/redis.aof")
	t.Setenv("REDIS_APPENDFSYNC", "everysec")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.server.Bind != "127.0.0.1" || cfg.server.Port != 6379 {
		t.Fatalf("unexpected server config: %#v", cfg.server)
	}
	if cfg.aof.Enabled || cfg.aof.FilePath != "/tmp/redis.aof" || cfg.aof.FSyncPolicy != redis.FsEverySecond {
		t.Fatalf("unexpected AOF config: %#v", cfg.aof)
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	clearConfigEnv(t)
	for name, value := range map[string]string{
		"REDIS_PORT":        "70000",
		"REDIS_APPENDONLY":  "sometimes",
		"REDIS_APPENDFSYNC": "daily",
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, value)
			if _, err := loadConfig(); err == nil {
				t.Fatalf("expected %s=%q to fail", name, value)
			}
		})
	}
}
