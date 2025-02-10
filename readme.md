# Crypto Price Streaming Service

This project implements a real-time cryptocurrency price streaming service that ingests price data from Coinbase's WebSocket feed, processes it through Kafka and ksqlDB, and makes it available for analysis and automated trading signals.

## Features

- Real-time BTC-USD price data ingestion from Coinbase
- Scalable data processing pipeline using Kafka
- Stream processing and signal generation with ksqlDB
- Automated trading signal generation
- Modular microservices architecture
- Docker-based deployment

## Architecture

The service consists of several microservices and components:

- **Ingestor Service**: Connects to Coinbase WebSocket API to receive real-time BTC-USD ticker data and publishes it to Kafka
- **Signaler Service**: Processes raw price data and generates trading signals
- **Kafka**: Message broker for reliable and scalable data streaming
- **ksqlDB**: Stream processing engine for real-time data analysis and signal generation
- **Bot Service**: Consumes processed data and signals from Kafka for executing trading strategies

### System Diagram
```
[Coinbase WebSocket] → [Ingestor Service] → [Kafka] → [ksqlDB] → [Bot Service]
                                                   ↓
                                          [Signaler Service]
```

## Data Flow

1. Ingestor service subscribes to Coinbase WebSocket feed for BTC-USD ticker data
2. Raw ticker data is published to `btc_usd` Kafka topic
3. ksqlDB processes the data and creates derived streams:
   - `btc_usd_signals` for price signals and indicators
   - `btc_usd_unified` combining ticker data with signals
4. Signaler service processes the unified data to generate trading signals
5. Bot service consumes the processed data and signals for trade execution

## Setup

### Prerequisites

- Docker and Docker Compose v2.x or higher
- Go 1.21 or higher


### Running the Services

1. Start the infrastructure services:
```bash
docker-compose up
```

2. Start the application services:
```bash
# In separate terminals:
go run cmd/ingestor/main.go
go run cmd/signaler/main.go
go run cmd/bot/main.go
```

### Verification

1. Check if services are running:
```bash
docker-compose ps
```

2. View logs:
```bash
docker-compose logs -f
```

## Configuration

The services can be configured through environment variables or configuration files:

- `config/ingestor_config.go`: Ingestor service configuration
- `config/signaler_config.go`: Signaler service configuration
- `config/bot_config.go`: Bot service configuration

## Development

### Project Structure
```
.
├── cmd/                  # Application entry points
├── internal/             # Private application code
│   ├── bot/              # Bot service implementation
│   ├── config/           # Configuration
│   ├── ingestor/         # Ingestor service implementation
│   ├── kafka/            # Kafka producers and consumers
│   ├── signaler/         # Signaler service implementation
│   └── websocket/        # WebSocket client implementation
├── ksqldb.sql            # ksqlDB stream definitions
└── deps-docker-compose.yml # Infrastructure services
```

### Running Tests
```bash
go test ./...
```

## Troubleshooting

Common issues and solutions:

1. **Kafka Connection Issues**
   - Ensure Kafka is running: `docker-compose ps`
   - Check Kafka logs: `docker-compose logs kafka`






