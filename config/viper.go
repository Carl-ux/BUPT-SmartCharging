package config

import (
	"os"

	"github.com/spf13/viper"
)

func InitConfig() {
	workdir, _ := os.Getwd()
	viper.SetConfigName("application")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workdir + "/config/")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
