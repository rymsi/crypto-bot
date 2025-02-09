package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	CoinbaseWSURL string   `mapstructure:"coinbase_ws_url"`
	ProductIDs    []string `mapstructure:"product_ids"`
	KafkaBrokers  []string `mapstructure:"kafka_brokers"`
	KafkaTopic    string   `mapstructure:"kafka_topic"`
	LogLevel      string   `mapstructure:"log_level"`
}

func LoadConfig(sugar *zap.SugaredLogger) (*Config, error) {
	viper.SetDefault("coinbase_ws_url", "wss://ws-feed.exchange.coinbase.com")
	viper.SetDefault("product_ids", []string{"BTC-USD"})
	viper.SetDefault("kafka_topic", "market_data")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("kafka_brokers", []string{"localhost:9092"})

	// You can set config file & path if desired
	// viper.SetConfigFile(".env") // or "config.yaml", etc.

	viper.AutomaticEnv() // read env vars

	if err := viper.ReadInConfig(); err != nil {
		// it’s okay if no config file, handle if needed
		sugar.Infow("No config file found, proceeding with env vars only. Err: %v\n", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		sugar.Errorw("Failed to unmarshal config", "error", err)
		return nil, err
	}
	return &cfg, nil

}
