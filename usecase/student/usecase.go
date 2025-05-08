package student

import (
	"fmt"
	"log"
	"student-service/models"
	"student-service/repository"
	"student-service/usecase"
)

type StudentUsecase struct {
	studentRepo repository.StudentRepositoryHandler
}

func (u *StudentUsecase) CreateStudentRecords(records *[]models.Student) []models.InsertionRecord {
	return u.studentRepo.CreateStudentRecord(records)
}
func (u *StudentUsecase) UpdateStudentRecord(records models.Student) (models.Student, error) {
	var err error
	//check if Student with given entry exists
	studentData, err := u.studentRepo.GetStudents(fmt.Sprintf("id = %d", records.Id))
	if err != nil {
		log.Println(fmt.Sprintf("error in getting student with id %d", records.Id), err.Error())
		return records, err
	}
	if len(studentData) == 0 || studentData[0].Id != records.Id {
		return records, fmt.Errorf("student with id %d is not available", records.Id)
	}
	if err = u.studentRepo.UpdateStudents(records); err != nil {
		log.Println("could not update student record")
		return records, err
	}
	// fetch updated record
	studentData, err = u.studentRepo.GetStudents(fmt.Sprintf("id = %d", records.Id))
	return studentData[0], err
}
func NewStudentUsecaseHandler(studentRepo repository.StudentRepositoryHandler) usecase.StudentsUsecaseHandler {
	return &StudentUsecase{studentRepo: studentRepo}
}
