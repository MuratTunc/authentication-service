CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    mail_address VARCHAR(255) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('Admin', 'Sales Representative', 'Customer')),
    phone_number VARCHAR(20),
    language_preference VARCHAR(10) DEFAULT 'en',
    resetcode VARCHAR(20),
    reset_verified BOOLEAN NOT NULL DEFAULT false,
    authentication_code VARCHAR(20),
    activated BOOLEAN NOT NULL DEFAULT false,
    login_status BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
