package requests

import (
	"log"
	"student-service/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type VaccineRecordsRequestHandler interface {
	Bind(c echo.Context, request interface{}, model interface{}) error
}
type VaccineRecordRequest struct{}

type VaccineRecordCreateRequest struct {
	StudentId int `json:"student_id" validate:"required"`
	DriveId   int `json:"drive_id" validate:"required"`
}
type GetStudentVaccineRecordRequest struct {
	Id          int    `param:"id"`
	RollNo      string `query:"roll_no"`
	VaccineName string `query:"vaccine_name"`
	Class       string `query:"class" validate:"omitempty,checkValidGradeUpdate"`
	Name        string `query:"name"`
	Pagination  Pagination
}

type GenerateReportRequest struct {
	Class       string `query:"class" validate:"omitempty,checkValidGradeUpdate"`
	VaccineName string `query:"vaccine_name"`
	RequestId   string
}

func (r VaccineRecordRequest) Bind(c echo.Context, req interface{}, model interface{}) error {
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
	case *VaccineRecordCreateRequest:
		data := models.VaccineRecord{}
		data.StudentId = req.(*VaccineRecordCreateRequest).StudentId
		data.DriveId = req.(*VaccineRecordCreateRequest).DriveId
		modelptr := model.(*[]models.VaccineRecord)
		*modelptr = append(*modelptr, data)

	case *GetStudentVaccineRecordRequest:
		req.(*GetStudentVaccineRecordRequest).Pagination = GetPagination(req.(*GetStudentVaccineRecordRequest).Pagination)
	case *GenerateReportRequest:
		req.(*GenerateReportRequest).RequestId = uuid.NewString()
	default:
		log.Println("request type Unknown for transformation", v)
	}
	return nil
}
func NewVaccineRecordRequestHandler() VaccineRecordsRequestHandler {
	return VaccineRecordRequest{}
}
