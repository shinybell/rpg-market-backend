-- wallet_transactions テーブルの type カラムから 'charge' を削除
ALTER TABLE wallet_transactions 
MODIFY COLUMN type ENUM(
    'sales_deposit', 
    'purchase', 
    'withdrawal', 
    'point_grant', 
    'point_expire', 
    'refund'
) NOT NULL;
