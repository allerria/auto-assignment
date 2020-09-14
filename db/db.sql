create table urls
(
    short_url text not null
        constraint urls_pk
            primary key,
    long_url  text not null
);

create unique index urls_short_url_uindex
    on urls (short_url);
