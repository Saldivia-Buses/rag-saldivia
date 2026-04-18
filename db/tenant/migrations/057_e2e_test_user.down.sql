-- 053_e2e_test_user.down.sql
DELETE FROM user_roles WHERE user_id = 'u-e2e-test';
DELETE FROM users WHERE id = 'u-e2e-test';
