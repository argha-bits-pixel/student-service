package consumers

type BulkUploadConsumers interface {
	SubscribeToBulkUploadQueue(queueName string)
}
