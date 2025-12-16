-- 初期カテゴリデータの投入
INSERT INTO item_categories (name, parent_id, display_order, path) VALUES
-- トップレベルカテゴリ
('武器', NULL, 1, '/1'),
('防具', NULL, 2, '/2'),
('アイテム', NULL, 3, '/3'),
('魔法書', NULL, 4, '/4'),
('素材', NULL, 5, '/5');

-- 武器のサブカテゴリ
INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '剣', id, 1, CONCAT('/1/', id) FROM item_categories WHERE name = '武器';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '杖', id, 2, CONCAT('/1/', id) FROM item_categories WHERE name = '武器';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '弓', id, 3, CONCAT('/1/', id) FROM item_categories WHERE name = '武器';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '槍', id, 4, CONCAT('/1/', id) FROM item_categories WHERE name = '武器';

-- 防具のサブカテゴリ
INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '鎧', id, 1, CONCAT('/2/', id) FROM item_categories WHERE name = '防具';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '盾', id, 2, CONCAT('/2/', id) FROM item_categories WHERE name = '防具';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT 'ヘルメット', id, 3, CONCAT('/2/', id) FROM item_categories WHERE name = '防具';

-- アイテムのサブカテゴリ
INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT '回復薬', id, 1, CONCAT('/3/', id) FROM item_categories WHERE name = 'アイテム';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT 'バフアイテム', id, 2, CONCAT('/3/', id) FROM item_categories WHERE name = 'アイテム';

INSERT INTO item_categories (name, parent_id, display_order, path)
SELECT 'その他消耗品', id, 3, CONCAT('/3/', id) FROM item_categories WHERE name = 'アイテム';

-- 初期ブランドデータの投入
INSERT INTO brands (name) VALUES
('伝説の鍛冶屋'),
('王国御用達'),
('冒険者ギルド公認'),
('魔法学院製'),
('ドワーフの工房'),
('エルフの技巧'),
('古代遺物'),
('ノーブランド');
