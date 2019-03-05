/*
Package pubsub contains two generic interfaces for publishing data to queues and subscribing and consuming data from those queues.

    // Publisher is a generic interface to encapsulate how we want our publishers
    // to behave. Until we find reason to change, we're forcing all publishers
    // to emit protobufs.
    type Publisher interface {
        // Publish will publish a message.
        Publish(ctx context.Context, key string, msg proto.Message) error
        // Publish will publish a []byte message.
        PublishRaw(ctx context.Context, key string, msg []byte) error
    }

    // Subscriber is a generic interface to encapsulate how we want our subscribers
    // to behave. For now the system will auto stop if it encounters any errors. If
    // a user encounters a closed channel, they should check the Err() method to see
    // what happened.
    type Subscriber interface {
        // Start will return a channel of raw messages
        Start() <-chan SubscriberMessage
        // Err will contain any errors returned from the consumer connection.
        Err() error
        // Stop will initiate a graceful shutdown of the subscriber connection
        Stop() error
    }

Where a `SubscriberMessage` is an interface that gives implementations a hook for acknowledging/delete messages. Take a look at the docs for each implementation in `pubsub` to see how they behave.

There are currently 3 implementations of each type of `pubsub` interfaces:

For pubsub via Amazon's SNS/SQS, you can use the `pubsub/aws` package.

For pubsub via Google's Pubsub, you can use the `pubsub/gcp` package.

For pubsub via Kafka topics, you can use the `pubsub/kafka` package.

For publishing via HTTP, you can use the `pubsub/http` package.
*/
package pubsub // import "github.com/NYTimes/gizmo/pubsub"
