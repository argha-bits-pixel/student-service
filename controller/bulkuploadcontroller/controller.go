package bulkuploadcontroller

import (
	"log"
	"net/http"
	"student-service/controller"
	"student-service/models"
	"student-service/requests"
	"student-service/response"
	"student-service/usecase"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	req  requests.BulkUploadRequestHandler
	uc   usecase.BulkUploadRequestHandler
	resp response.StudentResponseHandler
}

const (
	BULK_STUDENT_RECORD = "STUDENT_RECORD"
	BULK_VACCINE_RECORD = "VACCINE_RECORD"
)

func (v Controller) CreateStudentRecordBulk(c echo.Context) error {
	var err error
	req := new(requests.BulkUploadRequest)
	model := new(models.BulkUploadModel)
	model.RequestType = BULK_STUDENT_RECORD
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	if err = v.uc.UploadBulkRequestFile(model); err != nil {
		log.Println("error in processing student record update Request")
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message": "Request Accepted! Please check after sometime",
		"data": map[string]string{
			"request_id": model.RequestId,
			"status":     model.Status,
		},
	})
}
func (v Controller) CreateVaccinationRecordBulk(c echo.Context) error {
	var err error
	req := new(requests.BulkUploadRequest)
	model := new(models.BulkUploadModel)
	model.RequestType = BULK_VACCINE_RECORD
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	if err = v.uc.UploadBulkRequestFile(model); err != nil {
		log.Println("error in processing student record update Request")
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message": "Request Accepted! Please check after sometime",
		"data": map[string]string{
			"request_id": model.RequestId,
			"status":     model.Status,
		},
	})
}
func (v Controller) GetBulkUploadStatus(c echo.Context) error {
	var err error
	req := new(requests.GetBulkUploadRequest)
	model := new(models.BulkUploadModel)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	total, resp, err := v.uc.GetBulkUploadDetails(req.RequestId, req.Pagination)
	if err != nil {
		log.Println("error in getting bulkUpload details Detail", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "bulk_upload details fetched successfully",
		"data":    resp,
		"limit":   req.Pagination.Limit,
		"offset":  req.Pagination.Offset,
		"total":   total,
	})
	// log.Println("request for adding student record is", model)
	// resp, err := v.uc.UpdateStudentRecord(*model)
	// if err != nil {
	// 	log.Println("student record update failed", err.Error())
	// 	return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	// }
	// return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
}

func NewBulkUploadController(e *echo.Echo, req requests.BulkUploadRequestHandler, uc usecase.BulkUploadRequestHandler, resp response.StudentResponseHandler) controller.StudentController {
	studentServiceController := Controller{
		req:  req,
		uc:   uc,
		resp: resp,
	}
	e.POST("/bulk-upload/students", studentServiceController.CreateStudentRecordBulk)
	e.POST("/bulk-upload/vaccine-records", studentServiceController.CreateVaccinationRecordBulk)
	e.GET("/bulk-upload/:request_id", studentServiceController.GetBulkUploadStatus)
	e.GET("/bulk-upload", studentServiceController.GetBulkUploadStatus)
	return e
}
