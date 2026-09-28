create table if not exists app_user (
    id uuid primary key default gen_random_uuid(),
    email text not null,
    password_hash text not null,
    phone_number text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create table if not exists file (
    id uuid primary key default gen_random_uuid(),
    url text not null,
    filename text not null,
    type text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists chat_role (
    id uuid primary key default gen_random_uuid(),
    type text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists profile (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    icon_id uuid,
    nickname text not null,
    first_name text,
    last_name text,
    bio text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists session (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    expires_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists user_contact (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    contact_user_id uuid not null,
    pseudonym text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists chat (
    id uuid primary key default gen_random_uuid(),
    icon_id uuid,
    name text,
    description text,
    type text not null,
    members_count integer not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create table if not exists user_in_chat (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    chat_id uuid not null,
    role_id uuid not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create table if not exists sticker (
    id uuid primary key default gen_random_uuid(),
    file_id uuid not null,
    emoji text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists message (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    chat_id uuid not null,
    reply_to_message_id uuid,
    sticker_id uuid,
    content text,
    is_edited boolean not null default false,
    type text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create table if not exists attachment (
    id uuid primary key default gen_random_uuid(),
    file_id uuid not null,
    message_id uuid not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create table if not exists message_edit (
    id uuid primary key default gen_random_uuid(),
    message_id uuid not null,
    content text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists message_reaction (
    id uuid primary key default gen_random_uuid(),
    message_id uuid not null,
    user_id uuid not null,
    emoji text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

---- create above / drop below ----

drop table if exists message_reaction;
drop table if exists message_edit;
drop table if exists attachment;
drop table if exists message;
drop table if exists sticker;
drop table if exists user_in_chat;
drop table if exists chat;
drop table if exists user_contact;
drop table if exists session;
drop table if exists profile;
drop table if exists chat_role;
drop table if exists file;
drop table if exists app_user;
