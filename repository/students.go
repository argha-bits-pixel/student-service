package repository

import (
	"student-service/models"
)

type StudentRepositoryHandler interface {
	CreateStudentRecord(record *[]models.Student) []models.InsertionRecord
	GetStudents(filter string) ([]models.Student, error)
	UpdateStudents(student models.Student) error
}
