# `api-kafka-websocket-pubsub` 
* This is an example based on a prototype from an NYTimes hack week. It mixes `gizmo/server.SimpleServer`, `gizmo/server.SimpleService`, `gizmo/pubsub.KafkaPublisher`, `gizmo/pubsub.KafkaSubscriber` and `gorilla/websocket` and was used to test out realtime, collaborative crossword games.
* The server offers 3 endpoints to allow users to:
  1. Create a new topic on Kafka (visit http://localhost:8080/svc/v1/create to get a 'stream ID')
  2. Upgrade a request to a websocket connection and expose the topic over it
  3. Serve an HTML page that demos the service.(visit http://localhost:8080/svc/v1/demo/{stream_id from 'create'})

### This demo requires Kafka and Zookeeper to be installed and running locally by default.
  * To install and run on OS X, run: `brew install kafka` and then `zookeeper-server-start.sh /usr/local/etc/kafka/zookeeper.properties` to run Zookeeper and `kafka-server-start.sh /usr/local/etc/kafka/server.properties` to start a Kafka broker.

### The config in this example is loaded via a local JSON file and the default `gizmo/config.Config` struct.
