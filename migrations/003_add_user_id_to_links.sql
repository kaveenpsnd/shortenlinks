-- Add user_id column to short_links table
ALTER TABLE short_links 
ADD COLUMN user_id TEXT REFERENCES users(id);

-- Add index for faster user-based queries
CREATE INDEX idx_short_links_user_id ON short_links(user_id);
