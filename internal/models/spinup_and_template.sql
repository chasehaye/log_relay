-- create database

CREATE DATABASE log_relay;

CREATE TABLE receiver (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    password TEXT NOT NULL,
    token TEXT NOT NULL,
    receiver_email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_receiver_email UNIQUE(receiver_email),
    CONSTRAINT unique_receiver_token UNIQUE(token)
)

CREATE TABLE sender (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    origin_email VARCHAR(255),
    origin_phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
)

CREATE TABLE message (
    id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES sender(id) ON DELETE SET NULL,
    receiver_id INT NOT NULL REFERENCES receiver(id) ON DELETE CASCADE,
    header VARCHAR(255),
    body TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL,
    message_status VARCHAR(20) DEFAULT 'pending',
    importance VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
    CONSTRAINT chk_message_status CHECK (message_status IN ('pending','sent','error')),
    CONSTRAINT chk_message_type CHECK (message_type IN ('bug_report','inquiry','forward')),
    CONSTRAINT chk_importance CHECK (importance IN ('low','medium','high'))
)