CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    "group" TEXT NOT NULL,
    song TEXT NOT NULL
);

INSERT INTO songs ("group", song) VALUES
    ('The Beatles', 'Hey Jude'),
    ('Queen', 'Bohemian Rhapsody'),
    ('The Beatles', 'Let It Be'),
    ('Pink Floyd', 'Wish You Were Here'),
    ('Queen', 'Another One Bites the Dust');