CREATE TABLE users
(
    id       serial PRIMARY KEY,
    login    varchar NOT NULL,
    password varchar NOT NULL,
    UNIQUE (login)
);

CREATE TABLE data
(
    user_id  integer NOT NULL,
    key      varchar NOT NULL,
    data     bytea   NOT NULL,
    metadata varchar NULL,
    version  integer NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id),
    PRIMARY KEY (user_id, key)
);