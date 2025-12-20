-- Add RPG-style name and description to items table
ALTER TABLE items
ADD COLUMN rpg_name VARCHAR(255) DEFAULT NULL COMMENT 'RPG風の商品名',
ADD COLUMN rpg_description TEXT DEFAULT NULL COMMENT 'RPG風の商品説明';
