-- Remove RPG-style fields from items table
ALTER TABLE items
DROP COLUMN rpg_name,
DROP COLUMN rpg_description;
