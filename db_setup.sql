/* SETUP TABLE */
CREATE TABLE links (
    id uuid UNIQUE PRIMARY KEY,
    url text UNIQUE NOT NULL,
    short_url varchar (16) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unique_visits integer NOT NULL DEFAULT 1
);

/* Create indexes on most searched-by columns */
CREATE INDEX url ON links(url);
CREATE INDEX short_url ON links(short_url);

/* Function to update last_accessed whenever row is accessed */
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.last_accessed = NOW();
RETURN NEW;
END;

CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

/* EXPECTED QUERIES
Get short_url by normal url:
UPDATE links
    SET unique_visits = unique_visits + 1
    WHERE url = ""
    RETURNING (url, short_url);

Get url by short_url:
UPDATE links
    SET unique_visits = unique_visits + 1
    WHERE short_url = ""
    RETURNING (url, short_url);

Create new short_url
INSERT INTO links (url, short_url)
    VALUES ("", "");
*/