-- 004_seed_data.sql

insert into chat_role (type) values
('creator'),
('admin'),
('member');

insert into user (email, password_hash, phone_number) values
('admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '+79998061092'),
('user@example.com',  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '+79998061093');

insert into profile (user_id, nickname, first_name, last_name)
select id, 'admin', 'Admin', 'Root' from user where email = 'admin@example.com';

insert into profile (user_id, nickname, first_name, last_name)
select id, 'user', 'Test', 'User' from user where email = 'user@example.com';

---- create above / drop below ----

delete from profile;
delete from user;
delete from chat_role;