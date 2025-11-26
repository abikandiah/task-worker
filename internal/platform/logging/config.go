package logging

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"environment"`
	ServiceName string `mapstructure:"service_name"`
	Level       string `mapstructure:"level"`
	Version     string `mapstructure:"version"`
	Format      string `mapstructure:"format"`
}

func SetConfigDefaults(v *viper.Viper) {
	v.SetDefault("logger.level", "INFO")
	v.SetDefault("logger.format", "json")
}

func BindEnvironmentVariables(v *viper.Viper) {
	v.BindEnv("logger.level", "LOG_LEVEL")
	v.SetDefault("logger.format", "LOG_FORMAT")
}
