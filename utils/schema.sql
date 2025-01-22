CREATE TABLE IF NOT EXISTS users_tbl (
    id INTEGER PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    role INTEGER,
    active BOOLEAN
);

CREATE TABLE IF NOT EXISTS projects_tbl (
    id INTEGER PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    active BOOLEAN,
    createdAt TEXT,
    updatedAt TEXT
);

CREATE TABLE IF NOT EXISTS repositories_tbl (
    id INTEGER PRIMARY KEY NOT NULL,
    projectId INTEGER NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    branch TEXT NOT NULL,
    buildTool TEXT NOT NULL,
    depFileName TEXT NOT NULL,
    tags TEXT,
    user TEXT NOT NULL,
    token TEXT NOT NULL,
    active BOOLEAN,
    createdAt TEXT,
    updatedAt TEXT
);