CREATE TABLE
    links (
        id integer primary key autoincrement,
        public_id text,
        urlparams text,
        data json
    );

CREATE INDEX links_public_id ON links (public_id);