create table if not exists roles
(
    id   bigserial primary key,
    role text unique not null
);

create table if not exists roles_permissions
(
    role_id       bigint not null references roles (id) on delete cascade,
    permission_id bigint not null references permissions (id) on delete cascade,
    primary key (role_id, permission_id)
);

create table if not exists users_roles
(
    user_id bigint not null references users (id) on delete cascade,
    role_id bigint not null references roles (id) on delete cascade,
    primary key (user_id, role_id)
);


insert into roles (role)
values ('admin'),
       ('manager'),
       ('visitor'),
       ('user');

insert into roles_permissions (role_id, permission_id)
values (4, 1),
       (4, 2),
       (4, 3),
       (4, 4),
       (4, 5),
       (4, 6),
       (4, 7),
       (4, 8)
