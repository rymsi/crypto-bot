package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type SignalerConfig struct {
	KafkaBrokers       []string `mapstructure:"kafka_brokers"`
	KafkaConsumerGroup string   `mapstructure:"kafka_consumer_group"`
}

func LoadSignalerConfig(sugar *zap.SugaredLogger) (*SignalerConfig, error) {
	viper.SetDefault("kafka_brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka_consumer_group", "bot-consumer-group")

	var cfg SignalerConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		sugar.Errorw("Failed to unmarshal config", "error", err)
		return nil, err
	}
	return &cfg, nil
}
