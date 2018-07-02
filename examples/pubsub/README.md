# Example on how to use AWS SNS + SQS for pub/sub communication

We're going to set up two services to follow the pub/sub pattern using AWS SNS and AWS SQS.  One service is api-sns-pub and the other service is sqs-sub.  They can both be found at https://github.com/NYTimes/gizmo/tree/master/examples/pubsub. For more information about pub/sub pattern, SNS, and SQS check out http://www.infoq.com/articles/AmazonPubSub. 

### High level overview

1. Create SNS Topic 
2. Create SQS Queue
3. Create GCP Pubsub Project
4. Get the SQS Queue to subscribe to the SNS Topic
5. Set up api-sns-pub example service to send messages to the SNS Topic
6. Set up sqs-sub example service to listen to the SQS Queue

### Set up a topic with AWS SNS console
1. Login to your AWS account
2. Go to https://console.aws.amazon.com/sns
3. Click on "Topics" on the left hand side
4. Click "Create new topic" 
5. Topic Name = TestTopic  (or whatever you want)
6. TestTopic = TestTopic  (or whatever you want)

### Set up an AWS SQS Queue with AWS console 
1. Go to https://console.aws.amazon.com/sqs
*  Click on Create New Queue 
*  Queue Name = TestQueue (or whatever you want)
*  Click on "TestQueue" and then click "Queue Actions" dropdown and chooose "Subscribe Queue to SNS Topic"
*  Choose the topic we created earlier (TestTopic) from the correct region and then hit "Subscribe"

### Set up a GCP Pubsub Project with GCP console
1. Go to https://console.cloud.google.com
*  Select an Existing project/Click on Create New Project
*  Go to https://console.cloud.google.com/cloudpubsub
*  Click on New Topic
*  Click on your created topic options (three dots)
*  Click on New Subscription

### Change SNS Subscription to Raw Message Delivery
1. Go to https://console.aws.amazon.com/sns
* Click on the blue ARN text next to "TestTopic" to take you to the "Topic Details: TestTopic" page
* Click on the checkmark besides the ARN Subscription ID, which is from your SQS "TestQueue"
* Click "Edit Subscription Attributes" from the dropdown menu titled "Other Subscription Actions"
* Turn on "Raw Message Delivery"

### Set up logging files
1. gizmo logs to the directory specified in config.json.  The example project is pointing to `/var/nyt/logs/cats-publisher/` for the api-sns-pub service, `/var/nyt/logs/cats-subscriber/` for the sqs-sub service, `/var/nyt/logs/cats-pub/` for the cli-sns-pub service and `/var/nyt/logs/dogs-pubsub/` for the gcp-pubsub service
* You must create this directory and add access.log and app.log inside it.
* You might not have permission to run the following commands in which case you'll have to put sudo before them.
* In your terminal run `mkdir -p /var/nyt/logs/`
* run `cd /var/nyt/logs/`
* run `mkdir cats-publisher`
* run `mkdir cats-subscriber`
* run `mkdir cats-pub`
* run `mkdir dogs-pubsub`
* run `touch cats-publisher/access.log`
* run `touch cats-publisher/app.log`
* run `touch cats-subscriber/app.log`
* run `touch cats-pub/app.log`
* run `touch dogs-pubsub/app.log`
* To allow the api-sns-pub server to access these files run the following commands:
* `chmod -R 777 cats-publisher/`
* `chmod -R 777 cats-subscriber/`
* `chmod -R 777 cats-pub/`
* `chmod -R 777 dogs-pubsub/`

### Set up api-sns-pub example server
1. run `go get github.com/NYTimes/gizmo` inside your $GOPATH folder
* run `cd $GOPATH/src/github.com/NYTimes/gizmo/examples/pubsub/api-sns-pub`
* run `go get ./...`
* update the config.json file to point to your newly created ARN for "Topic".  Example: `arn:aws:sns:us-east-1:123456789:TestTopic`
* also update the config.json file to contain your AccessKey and SecretKey
* ensure that the region in your config.json file is the same as the region you created your SNS Topic and SQS Queue 
* run `go install ./...`
* run `api-sns-pub` (this should be in your $GOPATH/bin folder)
* Your api-sns-pub server should now be running! 

### Push a message to your topic
1. Call the server using a PUT request to `localhost:8080/svc/nyt/cats` with a body like the following json:
```
{
	"url":"http://www.nytiems.com/cats-article",
	"title":"cats cats cats"
}
```
2. CURL example: `curl -X PUT -H "Content-Type: application/json" -d '{"url":"http://www.nytiems.com/cats-article","title":"cats cats cats"}' localhost:8080/svc/nyt/cats`
3. You should see `{"status":"success!"}`
4. If you check your SQS Queue, you should now have one message in it

### Read a message from your SQS Queue 
1. run `cd $GOPATH/src/github.com/NYTimes/gizmo/examples/pubsub/sqs-sub`
*  run `go get ./...`
*  update the config.json file to contain your AccessKey and SecretKey
*  also make sure it points to the correct region 
*  run `go install ./...`
*  run `sqs-sub`
*  Your sqs-sub server should now be running!  
*  You should have seen the message(s) you sent earlier that were stacked on the SQS Queue
*  Now, when you run your `PUT /svc/nyt/cats` request, it will update the SNS, which SQS is subscribed to, and your sqs-sub service is polling the SQS Queue.  If there is a message on the queue, it will be deleted from the queue and read by sqs-sub.

<br />
<h4> Congrats on setting up pub/sub communication with AWS SNS + SQS! </h4>
<br />

### Troubleshooting
* Make sure the region your created your SNS topic and SQS Queue matches the region in your config.json files.  
* Make sure the Access Key and Secret Access Key you passed into the config.json have appropriate credentials for communicating with SNS and SQS
