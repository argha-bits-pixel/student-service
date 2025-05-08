package usecase

import (
	"student-service/models"
	"student-service/requests"
)

type VaccineRecordUsecaseHandler interface {
	CreateVaccinationRecords(records *[]models.VaccineRecord) []models.VaccineInsertionRecord
	GetStudentVaccinationRecords(request *requests.GetStudentVaccineRecordRequest) (int, []models.GetStudentDetails, error)
	GetStudentVaccinationDashBoard() (int, int, error)
	GenerateVaccinationReport(request *requests.GenerateReportRequest) (string, error)
}
