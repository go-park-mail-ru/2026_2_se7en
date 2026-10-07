-- 004_seed_data.sql

insert into app_user (id, email, password_hash, phone_number, created_at, updated_at) values
('00000000-0000-4000-8000-000000000011', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '+79998061092', now() - interval '2 hours', now() - interval '2 hours'),
('00000000-0000-4000-8000-000000000012', 'user@example.com',  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '+79998061093', now() - interval '2 hours', now() - interval '2 hours');

insert into profile (id, user_id, nickname, first_name, last_name, created_at, updated_at) values
('00000000-0000-4000-8000-000000000021', '00000000-0000-4000-8000-000000000011', 'admin', 'Admin', 'Root', now() - interval '110 minutes', now() - interval '110 minutes'),
('00000000-0000-4000-8000-000000000022', '00000000-0000-4000-8000-000000000012', 'user', 'Test', 'User', now() - interval '110 minutes', now() - interval '110 minutes');

insert into chat (id, name, type, members_count, created_at, updated_at) values
('00000000-0000-4000-8000-000000000031', 'Тестовый чат', 'group', 2, now() - interval '1 hour', now() - interval '1 hour');

insert into chat_role (id, type, created_at, updated_at) values
('00000000-0000-4000-8000-000000000001', 'creator', now() - interval '3 hours', now() - interval '3 hours'),
('00000000-0000-4000-8000-000000000002', 'admin', now() - interval '3 hours', now() - interval '3 hours'),
('00000000-0000-4000-8000-000000000003', 'member', now() - interval '3 hours', now() - interval '3 hours');

insert into user_in_chat (id, user_id, chat_id, role_id, created_at, updated_at) values
('00000000-0000-4000-8000-000000000041', '00000000-0000-4000-8000-000000000011', '00000000-0000-4000-8000-000000000031', '00000000-0000-4000-8000-000000000001', now() - interval '50 minutes', now() - interval '50 minutes'),
('00000000-0000-4000-8000-000000000042', '00000000-0000-4000-8000-000000000012', '00000000-0000-4000-8000-000000000031', '00000000-0000-4000-8000-000000000003', now() - interval '50 minutes', now() - interval '50 minutes');

insert into message (id, user_id, chat_id, content, type, created_at, updated_at) values
('00000000-0000-4000-8000-000000000051', '00000000-0000-4000-8000-000000000011', '00000000-0000-4000-8000-000000000031', 'Привет! Как дела?', 'text', now() - interval '30 minutes', now() - interval '30 minutes'),
('00000000-0000-4000-8000-000000000052', '00000000-0000-4000-8000-000000000012', '00000000-0000-4000-8000-000000000031', 'Всё хорошо, спасибо!', 'text', now() - interval '20 minutes', now() - interval '20 minutes'),
('00000000-0000-4000-8000-000000000053', '00000000-0000-4000-8000-000000000011', '00000000-0000-4000-8000-000000000031', 'Рад это слышать.', 'text', now() - interval '10 minutes', now() - interval '10 minutes');

---- create above / drop below ----

delete from message
 where id in (
     '00000000-0000-4000-8000-000000000051',
     '00000000-0000-4000-8000-000000000052',
     '00000000-0000-4000-8000-000000000053'
 );

delete from user_in_chat
 where id in (
     '00000000-0000-4000-8000-000000000041',
     '00000000-0000-4000-8000-000000000042'
 );

delete from chat
 where id = '00000000-0000-4000-8000-000000000031';

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
