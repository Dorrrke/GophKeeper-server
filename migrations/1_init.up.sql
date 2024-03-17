CREATE TABLE IF NOT EXISTS users (
    uId serial PRIMARY KEY,
    login character(50) NOT NULL,
    hash character(200) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_login ON users (login);

CREATE TABLE IF NOT EXISTS cards (
    cId serial PRIMARY KEY, 
    name character(25) NOT NULL, 
    number character(16) NOT NULL, 
    date character(5) NOT NULL, 
    cvv integer NOT NULL,
    uId integer NOT NULL,
    deleted boolean NOT NULL,
    last_update timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS logins (
    lId serial PRIMARY KEY, 
    name character(25) NOT NULL, 
    login character(50) NOT NULL, 
    password character(50) NOT NULL,
    uId integer NOT NULL,
    deleted boolean NOT NULL,
    last_update timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS text_data (
    tId serial PRIMARY KEY, 
    name character(25) NOT NULL, 
    data character(255) NOT NULL,
    uId integer NOT NULL,
    deleted boolean NOT NULL,
    last_update timestamp with time zone NOT NULL
);

CREATE TABLE IF NOT EXISTS binares_data (
    bId serial PRIMARY KEY, 
    name character(25), 
    data character(255),
    uId integer,
    deleted boolean,
    last_update timestamp with time zone
);