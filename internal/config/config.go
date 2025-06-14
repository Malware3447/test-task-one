package config

import (
	"github.com/Malware3447/configo"
	"test-task-one/internal/config/redis"
)

type Config struct {
	App        configo.App       `yaml:"app" env-required:"true"`
	DatabasePg configo.Database  `yaml:"postgres" env-required:"true"`
	DatabaseCh configo.Database  `yaml:"clickhouse" env-required:"true"`
	Redis      redis.RedisConfig `yaml:"redis" env-required:"true"`
}

func (c Config) Env() string {
	return c.App.Env
}
