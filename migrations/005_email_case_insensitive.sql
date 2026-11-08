create unique index if not exists uq_app_user_email_active_lower
    on app_user (lower(email))
    where deleted_at is null;

drop index if exists uq_app_user_email_active;

---- create above / drop below ----

create unique index if not exists uq_app_user_email_active
    on app_user (email)
    where deleted_at is null;

drop index if exists uq_app_user_email_active_lower;
