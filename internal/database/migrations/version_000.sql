CREATE TABLE settings
(
    token      TEXT,
    last_index TEXT,
    version    INTEGER NOT NULL
);

CREATE TABLE media_items
(
    uuid           TEXT    NOT NULL CONSTRAINT media_items_pk PRIMARY KEY,
    remote_id      TEXT    NOT NULL,
    base_url       TEXT    NOT NULL,
    mime_type      TEXT,
    filename       TEXT,
    description    TEXT,
    downloaded     INTEGER DEFAULT 0 NOT NULL,
    local_path     TEXT    NOT NULL,
    local_filename TEXT    NOT NULL,
    file_size      INTEGER,
    created_at     TEXT    NOT NULL,
    modified_at    TEXT,
    synced_at      TEXT,
    CONSTRAINT media_items_pk_2 UNIQUE (local_path, local_filename)
);

INSERT INTO settings VALUES(NULL, NULL, 1);
