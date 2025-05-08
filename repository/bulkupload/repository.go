package bulkupload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"student-service/adapters/mysql"
	"student-service/adapters/rabbitmq"
	"student-service/models"
	"student-service/repository"
	"student-service/requests"

	"github.com/minio/minio-go/v7"
	"github.com/streadway/amqp"
)

type BulkUploadRepository struct {
	MinIoConn *minio.Client
	DB        *mysql.MysqlConnect
	Rabbit    *rabbitmq.RabbitChannel
}

func (b *BulkUploadRepository) SubscribeToQueue(queueName string) (<-chan amqp.Delivery, error) {
	log.Println("initiating consumption on ", queueName)
	return b.Rabbit.Consume(queueName, "", false, false, false, false, amqp.Table{})
}

func (b *BulkUploadRepository) UploadFileToMinio(filePath, bucketName, root, uniqueId string) (string, error) {
	uploadInfo, err := b.MinIoConn.FPutObject(context.Background(), bucketName, fmt.Sprintf("%s%s/%s", root, uniqueId, filepath.Base(filePath)), filePath, minio.PutObjectOptions{})
	os.Remove(filePath)
	return uploadInfo.Key, err
}
func (b *BulkUploadRepository) GetFileFromStorage(bucketName, fileLocation string) (string, error) {
	object, err := b.MinIoConn.GetObject(context.Background(), bucketName, fileLocation, minio.GetObjectOptions{})
	if err != nil {
		log.Println("Error in fetching object from object storage", err.Error())
		return "", err
	}
	defer object.Close()
	//create localFile
	tempFile, err := os.CreateTemp("", "school-vaccine-bulk-*")
	defer tempFile.Close()
	if err != nil {
		log.Println("error Creating temporary file for processing bulk request", err.Error())
		return "", err
	}
	_, err = io.Copy(tempFile, object)
	if err != nil {
		log.Println("error copying  temporary file for processing bulk request", err.Error())
		return "", err
	}
	return tempFile.Name(), nil
}

func (b *BulkUploadRepository) CreateFileUploadEntry(model *models.BulkUploadModel) error {
	return b.DB.Table("bulk_upload_jobs").Create(model).Error
}
func (b *BulkUploadRepository) UpdateFileUploadEntry(model *models.BulkUploadModel) error {
	updates := map[string]interface{}{}
	if model.Status != "" {
		updates["status"] = model.Status
	}
	if model.FilePath != "" {
		updates["file_path"] = model.FilePath
	}
	if model.TotalRecords != 0 {
		updates["total_records"] = model.TotalRecords
	}
	if model.ProcessedRecords != 0 {
		updates["processed_records"] = model.ProcessedRecords
	}
	if model.ErrorMessage != "" {
		updates["error_message"] = model.ErrorMessage
	}
	return b.DB.Table("bulk_upload_jobs").Updates(updates).Where("id = ?", model.Id).Error
}
func (b *BulkUploadRepository) GetBulkUploads(requestId string, pagination requests.Pagination) ([]models.BulkUploadModel, error) {
	result := []models.BulkUploadModel{}
	var err error
	if requestId == "" {
		log.Println("fetching bulkupload status requestes with no requestId")
		err = b.DB.Table("bulk_upload_jobs").
			Order("id ASC").
			Limit(pagination.Limit).
			Offset(pagination.Offset).
			Find(&result).Error
	} else {
		log.Println("fetching bulkupload status requestes with requestId", requestId)
		err = b.DB.Table("bulk_upload_jobs").
			Order("id ASC").
			Where("request_id =?", requestId).
			Limit(pagination.Limit).
			Offset(pagination.Offset).
			Find(&result).Error
	}
	return result, err
}
func (b *BulkUploadRepository) GetBulkUploadCounts(requestId string, pagination requests.Pagination) (int, error) {
	var result int
	var err error
	if requestId == "" {
		log.Println("fetching bulkupload status request count with no requestId")
		err = b.DB.Table("bulk_upload_jobs").
			Count(&result).Error
	} else {
		log.Println("fetching bulkupload status request count with no requestId")

		err = b.DB.Table("bulk_upload_jobs").
			Where("request_id = ?", requestId).
			Count(&result).Error
	}
	return result, err
}

func (b *BulkUploadRepository) SubmitToRabbitMQ(rmqData *models.BulkUploadModel, queueName string) error {
	body, _ := json.Marshal(rmqData)
	return b.Rabbit.Publish(
		"", "bulk-upload", false, false, amqp.Publishing{
			DeliveryMode: 2,
			ContentType:  "text/plain",
			Body:         body,
		},
	)
}

func NewBulkUploadRepositoryHandler(DB *mysql.MysqlConnect, MinIO *minio.Client, rabbit *rabbitmq.RabbitChannel) repository.BulkUploadRepositoryHandler {
	return &BulkUploadRepository{
		DB:        DB,
		MinIoConn: MinIO,
		Rabbit:    rabbit,
	}
}
