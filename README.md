# openspotmap

A public API for parkour spots

This application is built for Heroku but can be run anywhere else with little effort

## API Documentation

[Documentation at Swaggerhub](https://app.swaggerhub.com/api/theSuess/openspotmap/v0)

## Running openspotmap

You need a PostgreSQL database with the PostGIS extension to run this application
The database schema is in the `schema.sql` file.

The following environment variables have to be set:

* PORT: specifying the port on which the application should run
* DATABASE_URL: connection url in the following format: `postgres://user:password@host:port/database_name`
