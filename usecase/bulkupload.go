package usecase

import (
	"student-service/models"
	"student-service/requests"
)

type BulkUploadRequestHandler interface {
	UploadBulkRequestFile(req *models.BulkUploadModel) error
	ProcessBulkStudentRecord(model *models.BulkUploadModel) error
	ProcessBulkVaccineRecord(model *models.BulkUploadModel) error
	GetBulkUploadDetails(requestId string, pagination requests.Pagination) (int, []models.BulkUploadModel, error)
}
