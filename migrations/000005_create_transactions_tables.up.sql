-- transactions テーブル (取引)
CREATE TABLE transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    item_id BIGINT NOT NULL,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    price BIGINT NOT NULL CHECK (price >= 0),
    fee_amount BIGINT NOT NULL DEFAULT 0,
    profit_amount BIGINT NOT NULL,
    payment_status ENUM('pending', 'captured', 'failed', 'refunded', 'cancelled') DEFAULT 'pending',
    transaction_status ENUM('awaiting_pay', 'awaiting_ship', 'shipped', 'delivered', 'completed', 'cancelled') DEFAULT 'awaiting_pay',
    payment_method VARCHAR(50),
    payment_transaction_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL,

    FOREIGN KEY (item_id) REFERENCES items(id),
    FOREIGN KEY (buyer_id) REFERENCES users(id),
    FOREIGN KEY (seller_id) REFERENCES users(id),

    INDEX idx_buyer_id (buyer_id),
    INDEX idx_seller_id (seller_id),
    INDEX idx_item_id (item_id),
    INDEX idx_payment_status (payment_status),
    INDEX idx_transaction_status (transaction_status),
    INDEX idx_buyer_status_date (buyer_id, transaction_status, created_at),
    INDEX idx_seller_status_date (seller_id, transaction_status, created_at),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- shipments テーブル (配送情報)
CREATE TABLE shipments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    transaction_id BIGINT NOT NULL UNIQUE,
    tracking_number VARCHAR(255),
    carrier ENUM('yamato', 'japan_post', 'sagawa', 'other') NOT NULL,
    status ENUM('label_created', 'in_transit', 'out_for_delivery', 'delivered', 'failed') DEFAULT 'label_created',
    qr_code_url VARCHAR(500),
    shipped_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (transaction_id) REFERENCES transactions(id),
    INDEX idx_tracking_number (tracking_number),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- reviews テーブル (評価)
CREATE TABLE reviews (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    transaction_id BIGINT NOT NULL,
    reviewer_id BIGINT NOT NULL,
    reviewee_id BIGINT NOT NULL,
    rating ENUM('good', 'neutral', 'bad') NOT NULL,
    comment TEXT,
    helpful_count INT DEFAULT 0,
    unhelpful_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (transaction_id) REFERENCES transactions(id),
    FOREIGN KEY (reviewer_id) REFERENCES users(id),
    FOREIGN KEY (reviewee_id) REFERENCES users(id),

    UNIQUE KEY unique_review (transaction_id, reviewer_id),
    INDEX idx_reviewee_id (reviewee_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- transaction_messages テーブル (取引メッセージ - クローズド)
CREATE TABLE transaction_messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    transaction_id BIGINT NOT NULL,
    sender_id BIGINT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id),

    INDEX idx_transaction_id (transaction_id),
    INDEX idx_sender_id (sender_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
