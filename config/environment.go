package config

import "github.com/spf13/viper"

type Environment struct {
	DBUsername string `mapstructure:"DB_USERNAME"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBName     string `mapstructure:"DB_NAME"`
}

func LoadEnv(path, envName string) (env Environment, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(envName)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&env)
	return
}
