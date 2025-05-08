package repository

import (
	"student-service/models"
	"student-service/requests"
)

type VaccineRecordRepositoryHandler interface {
	CreateVaccinationRecord(record *[]models.VaccineRecord) []models.VaccineInsertionRecord
	GetStudentVaccinationRecord(filter string, pagination requests.Pagination) ([]models.StudentVaccinationDetail, error)
	GetStudentVaccinationRecordCount(filter, join string) (int, error)
}
