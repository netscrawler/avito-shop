CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(255) UNIQUE NOT NULL,
  password BYTEA NOT NULL,
  coins BIGINT NOT NULL CHECK (coins >= 0) DEFAULT 1000
);

CREATE INDEX idx_users_username ON users(username);

CREATE TABLE user_inventory (
    username VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL CHECK (quantity >= 0),
    PRIMARY KEY (username, item_name)
);

CREATE TABLE merch (
  name VARCHAR(255) PRIMARY KEY,
  price INT NOT NULL CHECK (price >= 0)
);

CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  sender_name VARCHAR(255) NOT NULL,
  receiver_name VARCHAR(255),
  transfer_type VARCHAR(255) NOT NULL,
  amount BIGINT NOT NULL CHECK (amount >= 0),
  timestamp TIMESTAMP NOT NULL
);

CREATE INDEX idx_transactions_sender ON transactions(sender_name);
CREATE INDEX idx_transactions_type ON transactions(transfer_type);

INSERT INTO merch (name, price) VALUES
  ('t-shirt', 80),
  ('cup', 20),
  ('book', 50),
  ('pen', 10),
  ('powerbank', 200),
  ('hoody', 300),
  ('umbrella', 200),
  ('socks', 10),
  ('wallet', 50),
  ('pink-hoody', 500);

