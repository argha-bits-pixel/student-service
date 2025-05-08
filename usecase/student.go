package usecase

import (
	"student-service/models"
)

type StudentsUsecaseHandler interface {
	CreateStudentRecords(records *[]models.Student) []models.InsertionRecord
	UpdateStudentRecord(records models.Student) (models.Student, error)
}
