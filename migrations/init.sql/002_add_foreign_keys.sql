-- Добавляем внешние ключи для таблицы user_inventory
ALTER TABLE user_inventory
    ADD CONSTRAINT fk_user_inventory_username
    FOREIGN KEY (username)
    REFERENCES users(username)
    ON DELETE CASCADE;

-- Добавляем внешний ключ для отправителя в таблице transactions
ALTER TABLE transactions
    ADD CONSTRAINT fk_transactions_sender
    FOREIGN KEY (sender_name)
    REFERENCES users(username)
    ON DELETE CASCADE; 