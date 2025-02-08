# Pocket Realm

The backend for the Pocket Realm game.  A bunch of modules written in go.  Personal project to get more experience working with go.

## Status
Actively being worked on

## Main Modules

* `cron` - Responsible for doing updates
* `game` - Main module that handles connections from the client

## Support Modules

These are kept seperate so they can be shared with the main modules

* `models` - Contains all the models using Gorm for management
* `redis` - Abstracts out the connection to redis
* `utils` - Houses various shared utilities

## Running

Services should be run using the `docker-compose` files in infrastructure

To manually run a main module, enter the `cron` or `game` directory and run `go mod .`

## Requirements

* Go
