package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type BotConfig struct {
	KafkaBrokers       []string `mapstructure:"kafka_brokers"`
	KafkaConsumerGroup string   `mapstructure:"kafka_consumer_group"`
}

func LoadBotConfig(sugar *zap.SugaredLogger) (*BotConfig, error) {
	viper.SetDefault("kafka_brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka_consumer_group", "bot-consumer-group")

	var cfg BotConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		sugar.Errorw("Failed to unmarshal config", "error", err)
		return nil, err
	}
	return &cfg, nil
}
