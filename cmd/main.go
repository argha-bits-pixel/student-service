package main

import (
	"flag"
	"log"
	"os"
	"student-service/adapters/minio"
	"student-service/adapters/mysql"
	"student-service/adapters/rabbitmq"
	bulkConsumer "student-service/consumers/bulkuploadconsumer"
	bulkuploadrp "student-service/repository/bulkupload"
	studentrp "student-service/repository/students"
	vrecordrp "student-service/repository/vaccinerecord"
	"student-service/server"

	bulkuploaduc "student-service/usecase/bulkupload"
	studentuc "student-service/usecase/student"
	vrecorduc "student-service/usecase/vaccinerecord"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading env file", err.Error())
	}
}

func StartBulkProcessor(queueName string) {
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
	studentRepo := studentrp.NewStudentRepositoryHandler(dbConn)
	studentUsecase := studentuc.NewStudentUsecaseHandler(studentRepo)
	bulkUploadRepository := bulkuploadrp.NewBulkUploadRepositoryHandler(dbConn, minIo, rabb)
	vaccineRecordRepo := vrecordrp.NewVaccineRecordRepositoryHandler(dbConn)
	vaccineRecordUsecase := vrecorduc.NewVaccineRecordUsecaseHandler(vaccineRecordRepo, studentRepo, bulkUploadRepository)
	bulkUploadUsecase := bulkuploaduc.NewBulkUploadRequestUsecaseHandler(bulkUploadRepository, studentUsecase, vaccineRecordUsecase)
	consumer := bulkConsumer.NewBulkUploadProcessor(bulkUploadUsecase, rabb)
	consumer.SubscribeToBulkUploadQueue(queueName)

}

func main() {
	service := flag.String("service", "", "Service Being Requested: server, bulkProcessor")
	flag.Parse()
	switch *service {
	case "server":
		server.Start()
	default:
		log.Println("Starting Bulk processor")
		StartBulkProcessor("bulk-upload")
	}
}
