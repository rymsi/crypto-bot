-- Create stream for the BTC-USD ticker data
CREATE STREAM IF NOT EXISTS btc_usd_stream (
    type VARCHAR,
    sequence BIGINT,
    product_id VARCHAR,
    price VARCHAR,
    open_24h VARCHAR,
    volume_24h VARCHAR,
    low_24h VARCHAR,
    high_24h VARCHAR,
    volume_30d VARCHAR,
    best_bid VARCHAR,
    best_bid_size VARCHAR,
    best_ask VARCHAR,
    best_ask_size VARCHAR,
    side VARCHAR,
    time VARCHAR,
    trade_id BIGINT,
    last_size VARCHAR
) WITH (
    kafka_topic = 'btc_usd',
    value_format = 'JSON',
    partitions = 1,
    replicas = 1
);

-- Create stream for the signals data
CREATE STREAM IF NOT EXISTS btc_usd_signals_stream (
    timestamp VARCHAR,
    avg_price DOUBLE
) WITH (
    kafka_topic = 'btc_usd_signals',
    value_format = 'JSON',
    partitions = 1,
    replicas = 1
);


-- Create the unified stream
CREATE STREAM IF NOT EXISTS btc_usd_unified AS
SELECT
    a.time AS timestamp,
    CAST(a.volume_24h AS DOUBLE) AS volume_24h,
    b.avg_price
FROM btc_usd_stream a
JOIN btc_usd_signals_stream b
WITHIN 5 SECONDS
ON a.time = b.timestamp;