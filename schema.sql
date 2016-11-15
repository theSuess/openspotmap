CREATE EXTENSION postgis;

CREATE TABLE spots (
       id SERIAL PRIMARY KEY,
       name TEXT NOT NULL,
       description TEXT NOT NULL,
       location GEOGRAPHY(POINT,4326) NOT NULL UNIQUE,
       images TEXT[]
)
