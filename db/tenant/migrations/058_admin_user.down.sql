-- 058_admin_user.down.sql
DELETE FROM user_roles WHERE user_id = 'u-admin';
DELETE FROM users WHERE id = 'u-admin';
