package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type IngestorConfig struct {
	CoinbaseWSURL string   `mapstructure:"coinbase_ws_url"`
	KafkaBrokers  []string `mapstructure:"kafka_brokers"`
}

func LoadIngestorConfig(sugar *zap.SugaredLogger) (*IngestorConfig, error) {
	viper.SetDefault("coinbase_ws_url", "wss://advanced-trade-ws.coinbase.com")
	viper.SetDefault("kafka_brokers", []string{"localhost:9092"})

	// You can set config file & path if desired
	// viper.SetConfigFile(".env") // or "config.yaml", etc.

	viper.AutomaticEnv() // read env vars

	if err := viper.ReadInConfig(); err != nil {
		// it’s okay if no config file, handle if needed
		sugar.Infow("No config file found, proceeding with env vars only. Err: %v\n", err)
	}

	var cfg IngestorConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		sugar.Errorw("Failed to unmarshal config", "error", err)
		return nil, err
	}
	return &cfg, nil

}
