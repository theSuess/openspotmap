CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS keys (
       id UUID PRIMARY KEY,
       name TEXT NOT NULL UNIQUE,
       email TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS spots (
       id SERIAL PRIMARY KEY,
       name TEXT NOT NULL UNIQUE,
       description TEXT NOT NULL,
       location GEOGRAPHY(POINT,4326) NOT NULL UNIQUE,
       images TEXT[],
       submitter UUID references keys(id) NOT NULL
)
