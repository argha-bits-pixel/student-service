package studentcontroller

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
	req  requests.StudentsRequestHandler
	uc   usecase.StudentsUsecaseHandler
	resp response.StudentResponseHandler
}

func (v Controller) CreateStudentRecord(c echo.Context) error {
	var err error
	req := new(requests.StudentCreateRequest)
	model := new([]models.Student)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding student record is", model)
	resp := v.uc.CreateStudentRecords(model)
	if resp[0].Status {
		return c.JSON(http.StatusCreated, v.resp.ProcessResponse(req, resp))
	}
	return c.JSON(http.StatusBadRequest, v.resp.ProcessResponse(req, resp))
}
func (v Controller) CreateStudentRecordBulk(c echo.Context) error {
	var err error
	req := new(requests.StudentUpdateRequest)
	model := new(models.Student)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding student record is", model)
	resp, err := v.uc.UpdateStudentRecord(*model)
	if err != nil {
		log.Println("student record update failed", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
}
func (v Controller) EditStudentRecord(c echo.Context) error {
	var err error
	req := new(requests.StudentUpdateRequest)
	model := new(models.Student)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding student record is", model)
	resp, err := v.uc.UpdateStudentRecord(*model)
	if err != nil {
		log.Println("student record update failed", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
}

/*
	func (v Controller) GetStudentRecord(c echo.Context) error {
		var err error
		req := new(requests.StudentGetRequest)
		model := new(models.Student)
		if err = v.req.Bind(c, req, model); err != nil {
			log.Println("error in binding request")
			return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
		}
		log.Printf("request received is %+v", req)
		return c.JSON(http.StatusOK, map[string]string{
			"boo": "haaaa",
		})
		// log.Println("request for adding student record is", model)
		// resp, err := v.uc.UpdateStudentRecord(*model)
		// if err != nil {
		// 	log.Println("student record update failed", err.Error())
		// 	return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
		// }
		// return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
	}
*/
func NewStudentServiceController(e *echo.Echo, req requests.StudentsRequestHandler, uc usecase.StudentsUsecaseHandler, resp response.StudentResponseHandler) controller.StudentController {
	studentServiceController := Controller{
		req:  req,
		uc:   uc,
		resp: resp,
	}
	e.POST("/students/bulk-upload", studentServiceController.EditStudentRecord)
	e.POST("/students", studentServiceController.CreateStudentRecord)
	e.PATCH("/students", studentServiceController.EditStudentRecord)
	return e
}
