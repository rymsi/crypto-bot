-- Create stream for the BTC-USD ticker data
CREATE STREAM IF NOT EXISTS BTC_USD_STREAM (
    product_id VARCHAR,
    price DOUBLE,
    side VARCHAR,
    `size` DOUBLE,
    time BIGINT,
    trade_id BIGINT
) WITH (
    kafka_topic = 'btc_usd',
    value_format = 'JSON',
    partitions = 1,
    replicas = 1,
    timestamp = 'time'
);

-- Create an enriched stream from btc_usd_stream
CREATE STREAM IF NOT EXISTS BTC_USD_STREAM_ENRICHED AS 
SELECT 
    product_id,
    price,
    side,
    `size`,
    time,
    CAST((time / 1000) * 1000 AS BIGINT) AS current_second,
    trade_id
FROM btc_usd_stream;



-- Create table for the signals data with volume, average price
CREATE TABLE IF NOT EXISTS BTC_USD_SIGNALS AS
SELECT 
    product_id,
    WINDOWSTART as window_start,
    WINDOWEND as window_end,
    SUM(`size`) AS volume_100s,
    AVG(price) AS avg_price_100s
FROM btc_usd_stream
    WINDOW HOPPING (SIZE 100 SECONDS, ADVANCE BY 1 SECOND)
GROUP BY product_id
EMIT CHANGES;

-- Create a stream from btc_usd_signals
CREATE STREAM IF NOT EXISTS BTC_USD_SIGNALS_STREAM (
    window_start BIGINT,
    window_end BIGINT,
    volume_100s DOUBLE,
    avg_price_100s DOUBLE
) WITH (
    kafka_topic = 'BTC_USD_SIGNALS',
    value_format = 'JSON',
    partitions = 1,
    replicas = 1
);

-- Join btc_usd_signals_stream with btc_usd_stream_enriched
CREATE STREAM IF NOT EXISTS BTC_USD_JOINED AS
SELECT 
    btc_usd_stream_enriched.product_id,
    btc_usd_stream_enriched.price,
    btc_usd_stream_enriched.current_second,
    btc_usd_stream_enriched.side,
    btc_usd_stream_enriched.`size`,
    btc_usd_stream_enriched.time,
    btc_usd_stream_enriched.trade_id,
    btc_usd_signals_stream.window_start,
    btc_usd_signals_stream.window_end,
    btc_usd_signals_stream.volume_100s,
    btc_usd_signals_stream.avg_price_100s
FROM btc_usd_stream_enriched 
JOIN btc_usd_signals_stream
WITHIN 5 SECONDS
ON btc_usd_stream_enriched.current_second = btc_usd_signals_stream.window_end;