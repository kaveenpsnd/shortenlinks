CREATE TABLE IF NOT EXISTS short_links (
    id BIGINT PRIMARY KEY,              -- Our Snowflake ID (NOT auto-increment)
    original_url TEXT NOT NULL,         -- The long URL (e.g. https://google.com/...)
    short_code VARCHAR(10) NOT NULL,    -- The Base62 string (e.g. "h7K9")
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,               -- Optional: For links that self-destruct
    clicks BIGINT DEFAULT 0             -- Track number of clicks
);

-- Indexing for speed:
-- We search by 'short_code' constantly (User visits /h7K9 -> DB lookup).
-- Without an index, Postgres has to scan the whole table. With an index, it's instant.
CREATE UNIQUE INDEX idx_short_code ON short_links(short_code);