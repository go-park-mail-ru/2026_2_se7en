alter table profile
    add constraint fk_profile_app_user
        foreign key (user_id) references app_user (id) on delete cascade,
    add constraint fk_profile_icon
        foreign key (icon_id) references file (id) on delete set null;

alter table session
    add constraint fk_session_app_user
        foreign key (user_id) references app_user (id) on delete cascade;

alter table user_contact
    add constraint fk_user_contact_owner_app_user
        foreign key (user_id) references app_user (id) on delete cascade,
    add constraint fk_user_contact_contact_app_user
        foreign key (contact_user_id) references app_user (id) on delete cascade;

alter table chat
    add constraint fk_chat_icon
        foreign key (icon_id) references file (id) on delete set null;

alter table user_in_chat
    add constraint fk_user_in_chat_app_user
        foreign key (user_id) references app_user (id) on delete cascade,
    add constraint fk_user_in_chat_chat
        foreign key (chat_id) references chat (id) on delete cascade,
    add constraint fk_user_in_chat_role
        foreign key (role_id) references chat_role (id) on delete restrict;

alter table sticker
    add constraint fk_sticker_file
        foreign key (file_id) references file (id) on delete cascade;

alter table message
    add constraint fk_message_app_user
        foreign key (user_id) references app_user (id) on delete cascade,
    add constraint fk_message_chat
        foreign key (chat_id) references chat (id) on delete cascade,
    add constraint fk_message_reply_to
        foreign key (reply_to_message_id) references message (id) on delete set null,
    add constraint fk_message_sticker
        foreign key (sticker_id) references sticker (id) on delete set null;

alter table attachment
    add constraint fk_attachment_file
        foreign key (file_id) references file (id) on delete cascade,
    add constraint fk_attachment_message
        foreign key (message_id) references message (id) on delete cascade;

alter table message_edit
    add constraint fk_message_edit_message
        foreign key (message_id) references message (id) on delete cascade;

alter table message_reaction
    add constraint fk_message_reaction_message
        foreign key (message_id) references message (id) on delete cascade,
    add constraint fk_message_reaction_app_user
        foreign key (user_id) references app_user (id) on delete cascade;

alter table profile
    add constraint uq_profile_user_id unique (user_id),
    add constraint uq_profile_nickname unique (nickname);

alter table chat_role
    add constraint uq_chat_role_type unique (type);

alter table app_user
    add constraint uq_app_user_phone_number unique (phone_number);

alter table file
    add constraint uq_file_url unique (url);

alter table user_contact
    add constraint uq_user_contact_pair unique (user_id, contact_user_id);

alter table message_reaction
    add constraint uq_message_reaction unique (message_id, user_id, emoji);

alter table app_user
    add constraint chk_app_user_email_format
        check (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    add constraint chk_app_user_email_length
        check (char_length(email) >= 5 and char_length(email) <= 254),
    add constraint chk_app_user_password_hash_length
        check (char_length(password_hash) >= 60),
    add constraint chk_app_user_phone_format
        check (phone_number is null or phone_number ~ '^\+?[0-9]{10,15}$'),
    add constraint chk_app_user_deleted_at
        check (deleted_at is null or deleted_at >= created_at);

alter table file
    add constraint chk_file_url_not_empty
        check (char_length(url) > 0),
    add constraint chk_file_filename_not_empty
        check (char_length(filename) > 0),
    add constraint chk_file_type
        check (type ~* '^[a-z0-9][a-z0-9!#$%&^_.+*-]*/[a-z0-9][a-z0-9!#$%&^_.+*-]*$');

alter table chat_role
    add constraint chk_chat_role_type
        check (type in ('creator', 'admin', 'member'));

alter table profile
    add constraint chk_profile_nickname_format
        check (nickname ~ '^[A-Za-z0-9_]{3,32}$'),
    add constraint chk_profile_first_name_length
        check (char_length(first_name) <= 64),
    add constraint chk_profile_last_name_length
        check (last_name is null or char_length(last_name) <= 64),
    add constraint chk_profile_bio_length
        check (bio is null or char_length(bio) <= 1000);

alter table session
    add constraint chk_session_expires
        check (expires_at > created_at);

alter table user_contact
    add constraint chk_user_contact_no_self
        check (user_id <> contact_user_id),
    add constraint chk_user_contact_pseudonym_length
        check (pseudonym is null or char_length(pseudonym) <= 64);

alter table chat
    add constraint chk_chat_type
        check (type in ('dialog', 'group', 'channel')),
    add constraint chk_chat_members_count
        check (members_count >= 0),
    add constraint chk_chat_name_length
        check (char_length(name) <= 128),
    add constraint chk_chat_description_length
        check (description is null or char_length(description) <= 500),
    add constraint chk_chat_deleted_at
        check (deleted_at is null or deleted_at >= created_at);

alter table user_in_chat
    add constraint chk_user_in_chat_deleted_at
        check (deleted_at is null or deleted_at >= created_at);

alter table sticker
    add constraint chk_sticker_emoji_length
        check (char_length(emoji) between 1 and 16);

alter table message
    add constraint chk_message_type
        check (type in ('text', 'sticker')),
    add constraint chk_message_content_by_type
        check (
            (type = 'text' and content is not null and char_length(content) <= 4096 and sticker_id is null)
            or (type = 'sticker' and sticker_id is not null and content is null)
        ),
    add constraint chk_message_no_self_reply
        check (reply_to_message_id is null or reply_to_message_id <> id),
    add constraint chk_message_deleted_at
        check (deleted_at is null or deleted_at >= created_at);

alter table attachment
    add constraint chk_attachment_deleted_at
        check (deleted_at is null or deleted_at >= created_at);

alter table message_edit
    add constraint chk_message_edit_content_not_empty
        check (char_length(content) > 0);

alter table message_reaction
    add constraint chk_message_reaction_emoji_length
        check (char_length(emoji) between 1 and 16);

---- create above / drop below ----

alter table message_reaction drop constraint if exists chk_message_reaction_emoji_not_empty;
alter table message_reaction drop constraint if exists chk_message_reaction_emoji_length;
alter table message_reaction drop constraint if exists chk_message_reaction_emoji_format;
alter table sticker drop constraint if exists chk_sticker_emoji_not_empty;
alter table sticker drop constraint if exists chk_sticker_emoji_length;
alter table sticker drop constraint if exists chk_sticker_emoji_format;
alter table message_edit drop constraint if exists chk_message_edit_content_not_empty;
alter table attachment drop constraint if exists chk_attachment_deleted_at;
alter table message
    drop constraint if exists chk_message_deleted_at,
    drop constraint if exists chk_message_no_self_reply,
    drop constraint if exists chk_message_content_by_type,
    drop constraint if exists chk_message_type;
alter table user_in_chat drop constraint if exists chk_user_in_chat_deleted_at;
alter table chat
    drop constraint if exists chk_chat_deleted_at,
    drop constraint if exists chk_chat_dialog_name,
    drop constraint if exists chk_chat_description_length,
    drop constraint if exists chk_chat_name_length,
    drop constraint if exists chk_chat_members_count,
    drop constraint if exists chk_chat_type;
alter table user_contact
    drop constraint if exists chk_user_contact_pseudonym_length,
    drop constraint if exists chk_user_contact_no_self,
    drop constraint if exists uq_user_contact_pair,
    drop constraint if exists fk_user_contact_contact_app_user,
    drop constraint if exists fk_user_contact_owner_app_user;
alter table session
    drop constraint if exists chk_session_expires,
    drop constraint if exists fk_session_app_user;
alter table profile
    drop constraint if exists chk_profile_bio_length,
    drop constraint if exists chk_profile_last_name_length,
    drop constraint if exists chk_profile_first_name_length,
    drop constraint if exists chk_profile_nickname_format,
    drop constraint if exists uq_profile_nickname,
    drop constraint if exists uq_profile_user_id,
    drop constraint if exists fk_profile_icon,
    drop constraint if exists fk_profile_app_user;
alter table chat_role drop constraint if exists chk_chat_role_type;
alter table chat_role drop constraint if exists uq_chat_role_type;
alter table file
    drop constraint if exists chk_file_type,
    drop constraint if exists chk_file_filename_not_empty,
    drop constraint if exists chk_file_url_not_empty,
    drop constraint if exists uq_file_url;
alter table app_user
    drop constraint if exists uq_app_user_phone_number,
    drop constraint if exists chk_app_user_deleted_at,
    drop constraint if exists chk_app_user_phone_format,
    drop constraint if exists chk_app_user_password_hash_length,
    drop constraint if exists chk_app_user_email_length,
    drop constraint if exists chk_app_user_email_format;
alter table message_reaction
    drop constraint if exists uq_message_reaction,
    drop constraint if exists fk_message_reaction_app_user,
    drop constraint if exists fk_message_reaction_message;
alter table message_edit
    drop constraint if exists fk_message_edit_message;
alter table attachment
    drop constraint if exists fk_attachment_message,
    drop constraint if exists fk_attachment_file;
alter table message
    drop constraint if exists fk_message_sticker,
    drop constraint if exists fk_message_reply_to,
    drop constraint if exists fk_message_chat,
    drop constraint if exists fk_message_app_user;
alter table sticker drop constraint if exists fk_sticker_file;
alter table user_in_chat
    drop constraint if exists fk_user_in_chat_role,
    drop constraint if exists fk_user_in_chat_chat,
    drop constraint if exists fk_user_in_chat_app_user;
alter table chat
    drop constraint if exists fk_chat_icon;
