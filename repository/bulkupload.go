package repository

import (
	"student-service/models"
	"student-service/requests"

	"github.com/streadway/amqp"
)

type BulkUploadRepositoryHandler interface {
	UploadFileToMinio(filePath, bucketName, root, uniqueId string) (string, error)
	CreateFileUploadEntry(model *models.BulkUploadModel) error
	SubmitToRabbitMQ(rmqData *models.BulkUploadModel, queueName string) error
	UpdateFileUploadEntry(model *models.BulkUploadModel) error
	SubscribeToQueue(queueName string) (<-chan amqp.Delivery, error)
	GetFileFromStorage(bucketName, fileLocation string) (string, error)
	GetBulkUploads(requestId string, pagination requests.Pagination) ([]models.BulkUploadModel, error)
	GetBulkUploadCounts(requestId string, pagination requests.Pagination) (int, error)
}
