package config

import "test-task-one/internal/config/db"

type Config struct {
	DatabasePg db.Database `yaml:"postgres" env-required:"true"`
	DatabaseCh db.Database `yaml:"clickhouse" env-required:"true"`
}
