package requests

import (
	"log"
	"student-service/models"

	"github.com/labstack/echo/v4"
)

type StudentsRequestHandler interface {
	Bind(c echo.Context, request interface{}, model interface{}) error
}
type StudentRequest struct{}
type StudentCreateRequest struct {
	Name    string `json:"name" validate:"required"`
	Class   string `json:"class" validate:"checkValidGrade"`
	Gender  string `json:"gender" validate:"required"`
	RollNo  string `json:"roll_no" validate:"required"`
	PhoneNo string `json:"phone_no" validate:"required"`
}
type StudentUpdateRequest struct {
	Id      int    `json:"id" validate:"required"`
	Name    string `json:"name,omitempty"`
	Class   string `json:"class,omitempty" validate:"omitempty,checkValidGradeUpdate"`
	Gender  string `json:"gender,omitempty"`
	RollNo  string `json:"roll_no,omitempty"`
	PhoneNo string `json:"phone_no,omitempty"`
}

func (r StudentRequest) Bind(c echo.Context, req interface{}, model interface{}) error {
	var err error

	if err = c.Bind(req); err != nil {
		log.Println("Error in reading request", err.Error())
		return err
	}
	if err = c.Validate(req); err != nil {
		log.Println("error in validating request", err.Error())
		return err
	}
	switch v := req.(type) {
	case *StudentCreateRequest:
		data := models.Student{}
		data.Name = req.(*StudentCreateRequest).Name
		data.Class = req.(*StudentCreateRequest).Class
		data.Gender = req.(*StudentCreateRequest).Gender
		data.RollNumber = req.(*StudentCreateRequest).RollNo
		data.PhoneNo = req.(*StudentCreateRequest).PhoneNo
		modelptr := model.(*[]models.Student)
		*modelptr = append(*modelptr, data)
	case *StudentUpdateRequest:
		model.(*models.Student).Id = req.(*StudentUpdateRequest).Id
		model.(*models.Student).Name = req.(*StudentUpdateRequest).Name
		model.(*models.Student).Class = req.(*StudentUpdateRequest).Class
		model.(*models.Student).Gender = req.(*StudentUpdateRequest).Gender
		model.(*models.Student).RollNumber = req.(*StudentUpdateRequest).RollNo
		model.(*models.Student).PhoneNo = req.(*StudentUpdateRequest).PhoneNo

	default:
		log.Println("request type Unknown for transformation", v)
	}

	return nil
}

func NewVaccineDriveRequestHandler() StudentsRequestHandler {
	return StudentRequest{}
}
