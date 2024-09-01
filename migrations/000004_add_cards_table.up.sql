create table if not exists cards (
    id bigserial primary key,
    title text not null,
    created_at timestamp default now()
);

alter table events add constraint fk_card foreign key(card_id) references cards(id)
