
DROP TABLE IF EXISTS users;
CREATE TABLE users (
    uid		    SERIAL PRIMARY KEY ,
    name		TEXT NOT NULL UNIQUE,
    password	TEXT,
    token	TEXT,
    refresh_token TEXT,
    user_type TEXT,
    verified BOOLEAN,
    email TEXT
);

DROP TABLE IF EXISTS chats;
CREATE TABLE chats (
    messageID SERIAL PRIMARY KEY,
    user_id_1 INTEGER REFERENCES users(uid) ON DELETE CASCADE,
    user_id_2 INTEGER REFERENCES users(uid) ON DELETE CASCADE, 
    body TEXT,
    messageTime TEXT
);



DROP TABLE IF EXISTS posts;
CREATE TABLE posts(
    id SERIAL PRIMARY KEY,
    field TEXT NOT NULL,
    uid INTEGER REFERENCES users(uid) ON DELETE CASCADE,
    intro TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    modified_at TIMESTAMP NOT NULL DEFAULT NOW(),
    participants INTEGER[],
    num_participants INTEGER
);

DROP TABLE IF EXISTS singleusers;
CREATE TABLE singleusers(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    commitment INTEGER,
    year INTEGER,
    location TEXT,
    faculty TEXT
);

DROP TABLE IF EXISTS pairedusers;
CREATE TABLE pairedusers(
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    partner TEXT NOT NULL UNIQUE
);
