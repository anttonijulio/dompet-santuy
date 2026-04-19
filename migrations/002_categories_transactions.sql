CREATE TABLE IF NOT EXISTS categories (
    id          VARCHAR(36)              NOT NULL PRIMARY KEY,
    user_id     VARCHAR(36)              NOT NULL,
    name        VARCHAR(100)             NOT NULL,
    icon        VARCHAR(50),
    color       VARCHAR(20),
    type        ENUM('income','expense') NOT NULL,
    created_at  DATETIME                 NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_categories_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
    id          VARCHAR(36)              NOT NULL PRIMARY KEY,
    user_id     VARCHAR(36)              NOT NULL,
    category_id VARCHAR(36)              NOT NULL,
    amount      BIGINT                   NOT NULL,
    type        ENUM('income','expense') NOT NULL,
    note        TEXT,
    date        DATETIME                 NOT NULL,
    created_at  DATETIME                 NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_transactions_user     FOREIGN KEY (user_id)     REFERENCES users(id)       ON DELETE CASCADE,
    CONSTRAINT fk_transactions_category FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE INDEX idx_transactions_user_date ON transactions(user_id, date);
CREATE INDEX idx_transactions_category  ON transactions(category_id);
