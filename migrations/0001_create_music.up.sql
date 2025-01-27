create table groups
(
    group_id   serial
        constraint groups_id_pk
            primary key,
    group_name text
        unique
);

alter table groups
    owner to postgres;

create index group_name__index
    on groups (group_name);

create table songs
(
    song_id   serial
        constraint songs_pk
            primary key,
    song_name text,
    group_id  integer
        constraint songs_groups_group_id_fk
            references groups
);

alter table songs
    owner to postgres;

create table details
(
    song_id      integer not null
        constraint music_songs_song_id_fk
            references songs,
    release_date date    not null,
    text         text    not null,
    link         text    not null
);

alter table details
    owner to postgres;

create index music_song_id__index
    on details (song_id);

create index song_name__index
    on songs (song_name);

create unique index songs_song_name_group_id_uindex
    on songs (song_name, group_id);