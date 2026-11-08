alter table message
    add column if not exists is_read boolean not null default false;

---- create above / drop below ----

alter table message
    drop column if exists is_read;
