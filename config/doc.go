/*
Package config contains a handful of structs meant for managing common configuration options and credentials. There are currently configs for:

    * MySQL
    * MongoDB
    * Oracle
    * AWS (SNS, SQS, S3, DynamoDB)
    * Kafka
    * Gorilla's `securecookie`
    * Gizmo Servers

The package also has a generic `Config` type that contains all of the above types. It's meant to be a 'catch all' struct that most applications should be able to use.

This package also contains functions to load these config structs from JSON files, JSON blobs in Consul k/v or environment variables.
*/
package config
