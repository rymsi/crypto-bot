import json
import numpy as np
import pandas as pd
from kafka import KafkaConsumer
from stable_baselines3 import DQN
import logging
import time

KAFKA_BROKER = "localhost:9092"
KAFKA_TOPIC = "BTC_USD_JOINED"
MODEL_PATH = "model"
TRADE_LOG_FILE = "trade_logs.txt"
INITIAL_BALANCE = 100000

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")

class TradingState:
    def __init__(self):
        self.balance = INITIAL_BALANCE
        self.position = 0  # 0 or 1 BTC
        self.entry_price = 0  # Price at which BTC was bought

trading_state = TradingState()

logging.info("Loading RL trading model...")
try:
    model = DQN.load(MODEL_PATH)
    logging.info("RL model loaded successfully.")
except Exception as e:
    logging.error(f"Error loading model: {e}")
    exit(1)


logging.info(f"Connecting to Kafka topic: {KAFKA_TOPIC}...")
consumer = KafkaConsumer(
    KAFKA_TOPIC,
    bootstrap_servers=KAFKA_BROKER,
    value_deserializer=lambda x: json.loads(x.decode("utf-8")),
    auto_offset_reset="latest",
)

logging.info(f"Connected to Kafka topic: {KAFKA_TOPIC}")


def preprocess_trade_data(trade):
    """
    Converts a raw Kafka JSON message into a NumPy array suitable for RL prediction.
    """
    try:
        features = np.array([
            trade["PRICE"],  # Current price
            1 if trade["SIDE"] == "BUY" else -1,  # Buy (+1) or Sell (-1)
            trade["VOLUME_100S"],  # Rolling volume over last 100 seconds
            trade["PRICE"] - trade["AVG_PRICE_100S"],  # Price momentum
            trade["VOLUME_100S"] - trade.get("prev_VOLUME_100S", trade["VOLUME_100S"])  # Volume change FIXME
        ], dtype=np.float32)

        return features
    except KeyError as e:
        logging.error(f"Missing key in trade data: {e}")
        return None


def execute_trade(action, trade):
    """
    Executes trades and updates balance and position
    """
    price = float(trade['PRICE'])
    action_map = {0: "HOLD", 1: "BUY", 2: "SELL"}
    action_str = action_map.get(int(action) if bool(action) else 0, "UNKNOWN")
    
    old_balance = trading_state.balance
    old_position = trading_state.position

    # Execute trade based on action
    if action_str == "BUY" and trading_state.position == 0:
        if trading_state.balance >= price:
            trading_state.balance -= price
            trading_state.position = 1
            trading_state.entry_price = price
    
    elif action_str == "SELL" and trading_state.position == 1:
        trading_state.balance += price
        trading_state.position = 0
        trading_state.entry_price = 0

    # Calculate PnL if position changed
    pnl = ""
    if old_position != trading_state.position:
        pnl = f", PnL: ${trading_state.balance - old_balance:.2f}"

    log_entry = (
        f"Time: {trade['TIME']}, Action: {action_str}, "
        f"Price: ${price:.2f}, Volume: {trade['VOLUME_100S']}, "
        f"Balance: ${trading_state.balance:.2f}, "
        f"Position: {trading_state.position} BTC{pnl}\n"
    )
    
    with open(TRADE_LOG_FILE, "a") as log_file:
        log_file.write(log_entry)
    
    logging.info(log_entry.strip())

logging.info("Starting Kafka consumer loop...")
logging.info(f"Initial Balance: ${trading_state.balance:.2f}, Position: {trading_state.position} BTC")

while True:
    message_batch = consumer.poll(timeout_ms=100)
    if not message_batch:
        continue
        
    for topic_partition, messages in message_batch.items():
        for message in messages:
            trade = message.value

            trade_features = preprocess_trade_data(trade)
            if trade_features is None:
                continue

            action, _states = model.predict(trade_features)
            execute_trade(action, trade)
