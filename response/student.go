package response

import (
	"fmt"
	"student-service/models"
	"student-service/requests"
	"time"
)

type StudentResponseHandler interface {
	ProcessResponse(req interface{}, data interface{}) interface{}
}
type StudentResponse struct {
	Id        int         `json:"id,omitempty"`
	Name      string      `json:"name"`
	Class     string      `json:"class"`
	Gender    string      `json:"gender"`
	RollNo    string      `json:"roll_no"`
	CreatedAt *time.Time  `json:"created_at,omitempty"`
	UpdatedAt *time.Time  `json:"update_at,omitempty"`
	PhoneNo   string      `json:"phone_no"`
	Links     interface{} `json:"_links,omitempty"`
}

type StudentConsolidatedResposne struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error,omitempty"`
}

func (r StudentConsolidatedResposne) ProcessResponse(req interface{}, result interface{}) interface{} {
	resp := StudentConsolidatedResposne{}
	switch req.(type) {
	case *requests.StudentCreateRequest:
		data := StudentResponse{}
		data.Id = result.([]models.InsertionRecord)[0].Record.Id
		data.Name = result.([]models.InsertionRecord)[0].Record.Name
		data.Class = result.([]models.InsertionRecord)[0].Record.Class
		data.Gender = result.([]models.InsertionRecord)[0].Record.Gender
		data.RollNo = result.([]models.InsertionRecord)[0].Record.RollNumber
		data.CreatedAt = result.([]models.InsertionRecord)[0].Record.CreatedAt
		data.UpdatedAt = result.([]models.InsertionRecord)[0].Record.UpdatedAt
		data.PhoneNo = result.([]models.InsertionRecord)[0].Record.PhoneNo
		if result.([]models.InsertionRecord)[0].Status {
			data.Links = geneRateHateOasForStudent(result.([]models.InsertionRecord)[0])
			resp.Message = "Student Successfully Onboarded"
		} else {
			resp.Error = result.([]models.InsertionRecord)[0].ErrorReason
			resp.Message = "Student Not Onboarded"
		}
		resp.Data = data
	case *requests.StudentUpdateRequest:
		data := StudentResponse{}
		data.Id = result.(models.Student).Id
		data.Name = result.(models.Student).Name
		data.Class = result.(models.Student).Class
		data.Gender = result.(models.Student).Gender
		data.RollNo = result.(models.Student).RollNumber
		data.CreatedAt = result.(models.Student).CreatedAt
		data.UpdatedAt = result.(models.Student).UpdatedAt
		data.PhoneNo = result.(models.Student).PhoneNo
		data.Links = geneRateHateOasForStudent(models.InsertionRecord{Record: result.(models.Student)})
		resp.Message = "Student Successfully Updated"
		resp.Data = data

	}

	return resp
}
func geneRateHateOasForStudent(data models.InsertionRecord) interface{} {
	hateOas := map[string]interface{}{}
	hateOas["self"] = map[string]string{
		"href":   fmt.Sprintf("http://localhost:8080/students/%d", data.Record.Id),
		"method": "GET",
	}
	hateOas["edit"] = map[string]string{
		"href":   fmt.Sprintf("http://localhost:8080/students/%d", data.Record.Id),
		"method": "PATCH",
	}

	return hateOas
}

func NewStudentResponseHandler() StudentResponseHandler {
	return StudentConsolidatedResposne{}
}
