-- some constraints

alter table events add constraint events_title_check unique (title);

alter table events add constraint events_text_blocks_check check ( array_length(text_blocks, 1) between 0 and 10);