
create index if not exists events_title_idx on events using gin(to_tsvector('web', title));

-- create index if not exists events_date_idx on events using gin(date);