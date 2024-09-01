create table if not exists events (
    id bigserial primary key,
    created_at timestamp(0) with time zone not null default now(),
    card_id bigint not null,
    title text not null,
    description text not null,
    text_blocks text[] not null,
    date date not null,
    version integer not null default 1
)