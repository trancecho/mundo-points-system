package config

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

// InitConfig 初始化配置
func InitConfig() {
	// 分不同模式加载配置文件
	mode := flag.String("mode", "dev", "运行模式")
	flag.Parse()
	if *mode == "prod" {
		viper.SetConfigName("config.prod")
	} else if *mode == "docker" {
		viper.SetConfigName("config.docker")
	} else if *mode == "dev" {
		viper.SetConfigName("config.dev")
	} else {
		log.Fatalf("Invalid mode: %s", *mode)
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}
