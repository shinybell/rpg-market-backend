-- ブランドデータの削除
DELETE FROM brands;

-- カテゴリデータの削除（子から順に）
DELETE FROM item_categories WHERE parent_id IS NOT NULL;
DELETE FROM item_categories WHERE parent_id IS NULL;
