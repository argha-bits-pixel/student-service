package server

import (
	"log"
	"net/http"
	"os"
	"student-service/adapters/minio"
	"student-service/adapters/mysql"
	"student-service/adapters/rabbitmq"
	"student-service/controller/bulkuploadcontroller"
	"student-service/controller/studentcontroller"
	"student-service/controller/vaccinerecordcontroller"
	bulkuploadrp "student-service/repository/bulkupload"
	studentrp "student-service/repository/students"
	vrecordrp "student-service/repository/vaccinerecord"
	"student-service/requests"
	"student-service/response"
	bulkuploaduc "student-service/usecase/bulkupload"
	studentuc "student-service/usecase/student"
	vrecorduc "student-service/usecase/vaccinerecord"
	"student-service/utils/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func newRouter() *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Validator = validator.NewValidator()
	dbConn, err := mysql.GetMySQLConnect()
	if err != nil {
		log.Println("error in connecting to db", err.Error())
		os.Exit(1)
	}
	minIo, err := minio.GetMinIOClient()
	if err != nil {
		log.Fatalln("error creating min IO Client", err.Error())
	}
	rabb := rabbitmq.GetRabbitConn()

	studentRequest := requests.NewVaccineDriveRequestHandler()
	studentRepo := studentrp.NewStudentRepositoryHandler(dbConn)
	studentUsecase := studentuc.NewStudentUsecaseHandler(studentRepo)
	studentResponse := response.NewStudentResponseHandler()
	studentcontroller.NewStudentServiceController(e, studentRequest, studentUsecase, studentResponse)

	bulkUploadRepository := bulkuploadrp.NewBulkUploadRepositoryHandler(dbConn, minIo, rabb)

	vaccineRecordRequest := requests.NewVaccineRecordRequestHandler()
	vaccineRecordRepo := vrecordrp.NewVaccineRecordRepositoryHandler(dbConn)
	vaccineRecordUsecase := vrecorduc.NewVaccineRecordUsecaseHandler(vaccineRecordRepo, studentRepo, bulkUploadRepository)
	vaccineRecordResponse := response.NewVaccineRecordResponseHandler()
	vaccinerecordcontroller.NewVaccineRecordServiceController(e, vaccineRecordRequest, vaccineRecordUsecase, vaccineRecordResponse)

	bulkUploadRequest := requests.NewBulkUploadRequestHandler()

	bulkUploadUsecase := bulkuploaduc.NewBulkUploadRequestUsecaseHandler(bulkUploadRepository, nil, nil)
	bulkuploadcontroller.NewBulkUploadController(e, bulkUploadRequest, bulkUploadUsecase, studentResponse)

	return e
}
