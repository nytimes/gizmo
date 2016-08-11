/*
Package config contains a handful of useful functions to load to configuration structs from JSON files, JSON blobs in Consul k/v or environment variables.

The subpackages contain structs meant for managing common configuration options and credentials. There are currently configs for:

* Go Kit Metrics
* MySQL
* MongoDB
* Oracle
* PostgreSQL
* AWS (S3, DynamoDB, ElastiCache)
* GCP
* Gorilla's `securecookie`

The package also has a generic `Config` type in the `config/combined` package that contains all of the above types. It's meant to be a 'catch all' convenience struct that many applications should be able to use.
*/
package config
