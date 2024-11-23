CREATE TABLE IF NOT EXISTS music (
    "group" TEXT PRIMARY KEY,
    song TEXT NOT NULL,
    releaseDate TEXT NOT NULL,
    "text" TEXT NOT NULL,
    link TEXT NOT NULL
);