-- wallet_transactions テーブルの type カラムに 'charge' を追加
ALTER TABLE wallet_transactions 
MODIFY COLUMN type ENUM(
    'sales_deposit', 
    'purchase', 
    'withdrawal', 
    'point_grant', 
    'point_expire', 
    'refund',
    'charge'
) NOT NULL;
