-- 004_seed_data.sql

insert into chat_role (id, type) values
('00000000-0000-4000-8000-000000000001', 'creator'),
('00000000-0000-4000-8000-000000000002', 'admin'),
('00000000-0000-4000-8000-000000000003', 'member');

insert into app_user (id, email, password_hash, phone_number) values
('00000000-0000-4000-8000-000000000011', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '+79998061092'),
('00000000-0000-4000-8000-000000000012', 'user@example.com',  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '+79998061093');

insert into profile (id, user_id, nickname, first_name, last_name) values
('00000000-0000-4000-8000-000000000021', '00000000-0000-4000-8000-000000000011', 'admin', 'Admin', 'Root'),
('00000000-0000-4000-8000-000000000022', '00000000-0000-4000-8000-000000000012', 'user', 'Test', 'User');

---- create above / drop below ----

delete from profile
 where id in (
     '00000000-0000-4000-8000-000000000021',
     '00000000-0000-4000-8000-000000000022'
 );

delete from app_user
 where id in (
     '00000000-0000-4000-8000-000000000011',
     '00000000-0000-4000-8000-000000000012'
 );

delete from chat_role
 where id in (
     '00000000-0000-4000-8000-000000000001',
     '00000000-0000-4000-8000-000000000002',
     '00000000-0000-4000-8000-000000000003'
 );
