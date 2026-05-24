CREATE TABLE
    links (
        id integer primary key autoincrement,
        public_id text,
        data json,
        meta json,
        created_at text not null default (datetime('now'))
    );

CREATE INDEX links_public_id ON links (public_id);