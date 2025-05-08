package repository

import (
	"fmt"
	"student-service/adapters/mysql"
	"student-service/models"
	"student-service/repository"
)

type StudentRepository struct {
	DB *mysql.MysqlConnect
}

func (r *StudentRepository) CreateStudentRecord(record *[]models.Student) []models.InsertionRecord {
	insertionDetails := []models.InsertionRecord{}
	for _, j := range *record {
		insertionDetail := models.InsertionRecord{}
		err := r.DB.Table("students").Create(&j).Error
		insertionDetail.Record = j
		insertionDetail.Status = true
		if err != nil {
			insertionDetail.Status = false
			insertionDetail.ErrorReason = err.Error()
		}
		insertionDetails = append(insertionDetails, insertionDetail)
	}
	return insertionDetails
}
func (r *StudentRepository) GetStudents(filter string) ([]models.Student, error) {
	result := []models.Student{}
	var err error
	if filter == "" {
		err = r.DB.Table("students").Find(&result).Error
		return result, err
	} else {
		err = r.DB.Table("students").Where(filter).Find(&result).Error
		return result, err
	}
}
func (r *StudentRepository) UpdateStudents(student models.Student) error {
	updateArray := map[string]interface{}{}

	if student.Name != "" {
		updateArray["name"] = student.Name
	}
	if student.RollNumber != "" {
		updateArray["roll_number"] = student.RollNumber
	}
	if student.Class != "" {
		updateArray["class"] = student.Class
	}
	if student.Gender != "" {
		updateArray["gender"] = student.Gender
	}
	if student.PhoneNo != "" {
		updateArray["phone_no"] = student.PhoneNo
	}
	return r.DB.Table("students").Where(fmt.Sprintf("id = %d", student.Id)).Updates(updateArray).Error
}

func NewStudentRepositoryHandler(DB *mysql.MysqlConnect) repository.StudentRepositoryHandler {
	return &StudentRepository{
		DB: DB,
	}
}
