-- Control Plane Database Schema

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(64) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User RDB resources
CREATE TABLE IF NOT EXISTS user_rdbs (
    id SERIAL PRIMARY KEY,
    uid VARCHAR(64) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    rdb_type VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, uid)
);

-- User KV resources
CREATE TABLE IF NOT EXISTS user_kvs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    uid VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    kv_type VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, uid)
);

-- Verification codes table
CREATE TABLE IF NOT EXISTS verification_codes (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false
);

-- Config sync tasks table (async task queue)
CREATE TABLE IF NOT EXISTS config_tasks (
    id SERIAL PRIMARY KEY,
    user_uid VARCHAR(64) NOT NULL,
    task_type VARCHAR(32) NOT NULL,  -- 'config_update', 'pod_create', etc.
    status VARCHAR(16) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_rdbs_user_id ON user_rdbs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_kvs_user_id ON user_kvs(user_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_email ON verification_codes(email);
CREATE INDEX IF NOT EXISTS idx_config_tasks_status ON config_tasks(status);
CREATE INDEX IF NOT EXISTS idx_config_tasks_user ON config_tasks(user_uid);
