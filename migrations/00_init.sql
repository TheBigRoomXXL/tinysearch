PRAGMA journal_mode=WAL;

CREATE TABLE pages (
    id         INTEGER    NOT NULL    PRIMARY KEY,
    url        TEXT       NOT NULL    UNIQUE,
    content    TEXT       NOT NULL
) STRICT;
