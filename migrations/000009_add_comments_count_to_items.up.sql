-- Add comments_count column to items table
ALTER TABLE items ADD COLUMN comments_count INT DEFAULT 0 AFTER likes_count;
