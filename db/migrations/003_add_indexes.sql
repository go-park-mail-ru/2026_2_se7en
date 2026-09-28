create unique index if not exists uq_app_user_email_active
    on app_user (email)
    where deleted_at is null;

create unique index if not exists uq_user_in_chat_active
    on user_in_chat (user_id, chat_id)
    where deleted_at is null;

create index if not exists idx_profile_icon_id
    on profile (icon_id);

create index if not exists idx_chat_icon_id
    on chat (icon_id);

create index if not exists idx_session_user_id
    on session (user_id);

create index if not exists idx_session_expires_at
    on session (expires_at);

create index if not exists idx_user_contact_user_id
    on user_contact (user_id);

create index if not exists idx_user_contact_contact_user_id
    on user_contact (contact_user_id);

create index if not exists idx_user_in_chat_chat_id
    on user_in_chat (chat_id);

create index if not exists idx_user_in_chat_user_id
    on user_in_chat (user_id);

create index if not exists idx_user_in_chat_role_id
    on user_in_chat (role_id);

create index if not exists idx_message_chat_id
    on message (chat_id);

create index if not exists idx_message_user_id
    on message (user_id);

create index if not exists idx_message_reply_to
    on message (reply_to_message_id);

create index if not exists idx_message_sticker_id
    on message (sticker_id);

create index if not exists idx_message_edit_message_id
    on message_edit (message_id);

create index if not exists idx_message_reaction_user_id
    on message_reaction (user_id);

create index if not exists idx_attachment_message_id
    on attachment (message_id);

create index if not exists idx_attachment_file_id
    on attachment (file_id);

create index if not exists idx_sticker_file_id
    on sticker (file_id);

---- create above / drop below ----

drop index if exists idx_sticker_file_id;
drop index if exists idx_attachment_file_id;
drop index if exists idx_attachment_message_id;
drop index if exists idx_message_reaction_user_id;
drop index if exists idx_message_edit_message_id;
drop index if exists idx_message_sticker_id;
drop index if exists idx_message_reply_to;
drop index if exists idx_message_user_id;
drop index if exists idx_message_chat_id;
drop index if exists idx_user_in_chat_role_id;
drop index if exists idx_user_in_chat_user_id;
drop index if exists idx_user_in_chat_chat_id;
drop index if exists idx_user_contact_contact_user_id;
drop index if exists idx_user_contact_user_id;
drop index if exists idx_session_expires_at;
drop index if exists idx_session_user_id;
drop index if exists idx_profile_icon_id;
drop index if exists idx_chat_icon_id;
drop index if exists idx_profile_user_id;
drop index if exists idx_message_reaction_message_id;
drop index if exists uq_user_in_chat_active;
drop index if exists uq_user_email_active;
drop index if exists uq_app_user_email_active;
