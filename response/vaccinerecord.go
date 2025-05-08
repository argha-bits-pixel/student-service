package response

import (
	"student-service/models"
	"student-service/requests"
)

type VaccineRecordResponseHandler interface {
	ProcessResponse(req interface{}, data interface{}) VaccineRecordConsolidatedResposne
}

type VaccineRecordResponse struct {
	Id        int         `json:"id,omitempty"`
	StudentId int         `json:"student_id"`
	DriveId   int         `json:"drive_id"`
	Links     interface{} `json:"_links,omitempty"`
}

type VaccineRecordConsolidatedResposne struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error,omitempty"`
	Total   int         `json:"total,omitempty"`
	Limit   int         `json:"limit,omitempty"`
	Offset  int         `json:"offset,omitempty"`
}

func (r VaccineRecordConsolidatedResposne) ProcessResponse(req interface{}, data interface{}) VaccineRecordConsolidatedResposne {
	resp := VaccineRecordConsolidatedResposne{}
	switch req.(type) {
	case *requests.VaccineRecordCreateRequest:
		v := VaccineRecordResponse{
			Id:        data.([]models.VaccineInsertionRecord)[0].Record.Id,
			StudentId: data.([]models.VaccineInsertionRecord)[0].Record.StudentId,
			DriveId:   data.([]models.VaccineInsertionRecord)[0].Record.DriveId,
		}
		if data.([]models.VaccineInsertionRecord)[0].Status {
			resp.Message = "Vaccination Record added Successfully"
		} else {
			resp.Message = "Vaccination Record addition failed"
			resp.Error = data.([]models.VaccineInsertionRecord)[0].ErrorReason
		}
		resp.Data = v
	case *requests.GetStudentVaccineRecordRequest:
		resp.Message = "student record fetched successfully"
		resp.Data = data
		resp.Limit = req.(*requests.GetStudentVaccineRecordRequest).Pagination.Limit
		resp.Offset = req.(*requests.GetStudentVaccineRecordRequest).Pagination.Offset
	}
	return resp
}
func NewVaccineRecordResponseHandler() VaccineRecordResponseHandler {
	return VaccineRecordConsolidatedResposne{}
}
