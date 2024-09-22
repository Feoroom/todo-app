

create table if not exists permissions (
    id bigserial primary key,
    permission text not null unique
);

create table if not exists users_permissions (
  user_id bigint references users on delete cascade,
  permission_id bigint references permissions on delete cascade,
  primary key (user_id, permission_id)
);

insert into permissions (permission)
values
    ('cards:create'),
    ('cards:read'),
    ('cards:update'),
    ('cards:delete'),

    ('events:create'),
    ('events:read'),
    ('events:update'),
    ('events:delete'),

    ('users:create'),
    ('users:read'),
    ('users:update'),
    ('users:delete');