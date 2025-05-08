package vaccinerecordcontroller

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
	req  requests.VaccineRecordsRequestHandler
	uc   usecase.VaccineRecordUsecaseHandler
	resp response.VaccineRecordResponseHandler
}

// we handle single create record through api , bulk record creation to be done via excel and will be handled via workers
func (v Controller) CreateVaccineRecord(c echo.Context) error {
	var err error
	req := new(requests.VaccineRecordCreateRequest)
	model := new([]models.VaccineRecord)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Println("request for adding vaccine record is", model)
	resp := v.uc.CreateVaccinationRecords(model)
	if resp[0].Status {
		return c.JSON(http.StatusCreated, v.resp.ProcessResponse(req, resp))
	}
	return c.JSON(http.StatusBadRequest, v.resp.ProcessResponse(req, resp))
}
func (v Controller) GetStudentVaccinationRecord(c echo.Context) error {
	var err error
	req := new(requests.GetStudentVaccineRecordRequest)
	model := new(models.Student)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	total, resp, err := v.uc.GetStudentVaccinationRecords(req)
	if err != nil {
		log.Println("error in getting vaccination Detail", err.Error())
	}
	log.Printf("resp generated received is %+v and total records are %d", resp, total)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	finalResp := v.resp.ProcessResponse(req, resp)
	finalResp.Total = total
	return c.JSON(http.StatusOK, finalResp)
	// log.Println("request for adding student record is", model)
	// resp, err := v.uc.UpdateStudentRecord(*model)
	// if err != nil {
	// 	log.Println("student record update failed", err.Error())
	// 	return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	// }
	// return c.JSON(http.StatusOK, v.resp.ProcessResponse(req, resp))
}
func (v Controller) GetVaccinationRecordDashBoard(c echo.Context) error {
	var err error
	req := new(requests.VaccineRecordRequest)
	model := new(models.Student)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	total, vaccinated, err := v.uc.GetStudentVaccinationDashBoard()
	if err != nil {
		log.Println("error in getting vaccination Detail", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, map[string]int{
		"total_students":      total,
		"vaccinated_students": vaccinated,
	})
}
func (v Controller) GenerateReport(c echo.Context) error {
	var err error
	req := new(requests.GenerateReportRequest)
	model := new(models.Student)
	if err = v.req.Bind(c, req, model); err != nil {
		log.Println("error in binding generate report request")
		return c.JSON(http.StatusBadRequest, response.ProcessErrorResponse(err))
	}
	log.Printf("request received is %+v", req)
	fileLoc, err := v.uc.GenerateVaccinationReport(req)
	if err != nil {
		log.Println("error in getting vaccination Detail", err.Error())
		return c.JSON(http.StatusInternalServerError, response.ProcessErrorResponse(err))
	}
	return c.JSON(http.StatusOK, map[string]string{
		"file": fileLoc,
	})
}

func NewVaccineRecordServiceController(e *echo.Echo, req requests.VaccineRecordsRequestHandler, uc usecase.VaccineRecordUsecaseHandler, resp response.VaccineRecordResponseHandler) controller.VaccineRecordController {
	vaccineRecordServiceController := Controller{
		req:  req,
		uc:   uc,
		resp: resp,
	}
	e.POST("/vaccine-records", vaccineRecordServiceController.CreateVaccineRecord)
	e.GET("/vaccine-records/students/:id", vaccineRecordServiceController.GetStudentVaccinationRecord)
	e.GET("/vaccine-records/students", vaccineRecordServiceController.GetStudentVaccinationRecord)
	e.GET("/vaccine-records/dashboard", vaccineRecordServiceController.GetVaccinationRecordDashBoard)
	e.GET("/vaccine-records/genrate-report", vaccineRecordServiceController.GenerateReport)
	return e
}
