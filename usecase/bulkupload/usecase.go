package bulkupload

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"student-service/models"
	"student-service/repository"
	"student-service/requests"
	"student-service/usecase"
	"student-service/utils/validator"

	"github.com/xuri/excelize/v2"
)

type BulkUploadRequestUsecase struct {
	BulkUploadRepo       repository.BulkUploadRepositoryHandler
	StudentUsecase       usecase.StudentsUsecaseHandler
	VaccineRecordusecase usecase.VaccineRecordUsecaseHandler
}

func (b *BulkUploadRequestUsecase) GetBulkUploadDetails(requestId string, pagination requests.Pagination) (int, []models.BulkUploadModel, error) {
	var count int
	var result []models.BulkUploadModel
	var err error
	//get count
	count, err = b.BulkUploadRepo.GetBulkUploadCounts(requestId, pagination)
	//handle error
	if err != nil {
		log.Println("error in fetching count", err.Error())
		return count, result, err
	}
	//get data
	result, err = b.BulkUploadRepo.GetBulkUploads(requestId, pagination)
	//handle error
	if err != nil {
		log.Println("error in fetching count", err.Error())
		return count, result, err
	}
	return count, result, err
}
func (b *BulkUploadRequestUsecase) ProcessBulkStudentRecord(model *models.BulkUploadModel) error {
	//change status to processing
	b.BulkUploadRepo.UpdateFileUploadEntry(&models.BulkUploadModel{Id: model.Id, Status: "PROCESSING"})
	//Get the file
	fileLoc, err := b.BulkUploadRepo.GetFileFromStorage(os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), model.FilePath)
	if err != nil {
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	//validate the file
	file, err := os.Open(fileLoc)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	defer file.Close()
	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	if !(bytes.HasPrefix(header, []byte{0x50, 0x4B, 0x03, 0x04}) || bytes.Equal(header, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})) {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	f, err := excelize.OpenFile(fileLoc)
	if err != nil {
		log.Printf("failed to open Excel file: %v", err)
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	defer f.Close()
	sheets := f.GetSheetList()
	log.Printf("Sheets: %v\n", sheets)
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	//Header Adjusted
	model.TotalRecords = len(rows) - 1
	if err != nil {
		log.Fatalf("failed to read rows: %v", err)
	}
	studentSet := new([]models.Student)
	insertionRecords := []models.InsertionRecord{}
	validat := validator.NewValidator()
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 5 {
			log.Printf("failed to open Excel file: %v", err)
			model.ErrorMessage = "Missing Columns"
			model.Status = "FAILED"
			b.BulkUploadRepo.UpdateFileUploadEntry(model)
			return nil
		}
		sReq := requests.StudentCreateRequest{
			Name:    row[0],
			Class:   row[1],
			Gender:  row[2],
			RollNo:  row[3],
			PhoneNo: row[4],
		}
		sModel := models.Student{
			Name:       row[0],
			Class:      row[1],
			Gender:     row[2],
			RollNumber: row[3],
			PhoneNo:    row[4],
		}
		err := validat.Validate(sReq)
		if err != nil {
			insertionRecord := models.InsertionRecord{
				Record:      sModel,
				Status:      false,
				ErrorReason: err.Error(),
			}
			insertionRecords = append(insertionRecords, insertionRecord)
			continue
		}
		*studentSet = append(*studentSet, sModel)
	}
	result := b.StudentUsecase.CreateStudentRecords(studentSet)

	log.Println("Request Processing Complete", result)
	result = append(result, insertionRecords...)

	//Create report
	reportFile := excelize.NewFile()
	reportShheetName := "Report"
	index, _ := reportFile.NewSheet(reportShheetName)
	reportHeaders := []string{"Name", "Class", "Gender", "Roll Number", "Phone Number", "Status", "Remarks"}
	for col, header := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1) // (col+1, row=1)
		reportFile.SetCellValue(reportShheetName, cell, header)
	}
	//Keeping track Of Processed File
	totalProcessed := 0
	for i, insertion := range result {
		rowNum := i + 2
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("A%d", rowNum), insertion.Record.Name)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("B%d", rowNum), insertion.Record.Class)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), insertion.Record.Gender)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("D%d", rowNum), insertion.Record.RollNumber)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("E%d", rowNum), insertion.Record.PhoneNo)
		if !insertion.Status {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Rejected")
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("G%d", rowNum), insertion.ErrorReason)
		} else {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Accepted")
			totalProcessed++
		}
	}
	model.ProcessedRecords = totalProcessed
	reportFile.SetActiveSheet(index)
	reportFileName := "Report.xlsx"
	err = reportFile.SaveAs(reportFileName)
	if err != nil {
		log.Println("error creating report file", err.Error(), reportFileName)
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	log.Println("Report File Created", reportFileName)
	//Upload report
	uploadedReportFile, err := b.BulkUploadRepo.UploadFileToMinio(reportFileName, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "reports/", model.RequestId)

	if err != nil {
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	log.Println("Report File Uploaded", reportFileName)
	model.FilePath = fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), uploadedReportFile)
	//change status
	model.Status = "PROCESSED"
	//update db
	b.BulkUploadRepo.UpdateFileUploadEntry(model)
	log.Println("Processing Complete", model)
	return nil
}

func (b *BulkUploadRequestUsecase) ProcessBulkVaccineRecord(model *models.BulkUploadModel) error {
	b.BulkUploadRepo.UpdateFileUploadEntry(&models.BulkUploadModel{Id: model.Id, Status: "PROCESSING"})
	//Get the file
	fileLoc, err := b.BulkUploadRepo.GetFileFromStorage(os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), model.FilePath)
	if err != nil {
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	//validate the file
	file, err := os.Open(fileLoc)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	defer file.Close()
	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	if !(bytes.HasPrefix(header, []byte{0x50, 0x4B, 0x03, 0x04}) || bytes.Equal(header, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})) {
		model.ErrorMessage = "Inavlid File, Only .xlsx or .xls allowed"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	f, err := excelize.OpenFile(fileLoc)
	if err != nil {
		log.Printf("failed to open Excel file: %v", err)
		model.ErrorMessage = "Internal Server Error"
		model.Status = "FAILED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	defer f.Close()
	sheets := f.GetSheetList()
	log.Printf("Sheets: %v\n", sheets)
	sheetName := sheets[0]

	rows, err := f.GetRows(sheetName)
	//Header Adjusted
	model.TotalRecords = len(rows) - 1
	if err != nil {
		log.Fatalf("failed to read rows: %v", err)
	}
	vaccineRecord := new([]models.VaccineRecord)
	insertionRecords := []models.VaccineInsertionRecord{}
	validat := validator.NewValidator()
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) != 2 {
			log.Printf("failed to open Excel file: %v", err)
			model.ErrorMessage = "Missing Columns"
			model.Status = "FAILED"
			b.BulkUploadRepo.UpdateFileUploadEntry(model)
			return nil
		}
		insertionRecord := models.VaccineInsertionRecord{}
		studentId, err := strconv.Atoi(row[0])
		if err != nil {
			log.Printf("invalid insertion Record: %v", err.Error())
			model.ErrorMessage = fmt.Sprintf("invalid entry at row %d", i+1)
			model.Status = "FAILED"
			b.BulkUploadRepo.UpdateFileUploadEntry(model)
			return nil
		}
		driveId, err := strconv.Atoi(row[1])
		if err != nil {
			log.Printf("invalid insertion Record: %v", err.Error())
			model.ErrorMessage = fmt.Sprintf("invalid entry at row %d", i+1)
			model.Status = "FAILED"
			b.BulkUploadRepo.UpdateFileUploadEntry(model)
			return nil
		}

		sReq := requests.VaccineRecordCreateRequest{
			StudentId: studentId,
			DriveId:   driveId,
		}
		sModel := models.VaccineRecord{
			StudentId: studentId,
			DriveId:   driveId,
		}
		err = validat.Validate(sReq)
		if err != nil {
			insertionRecord.Record = sModel
			insertionRecord.Status = false
			insertionRecord.ErrorReason = "invalid input"
			insertionRecords = append(insertionRecords, insertionRecord)
			continue
		}
		*vaccineRecord = append(*vaccineRecord, sModel)
	}
	result := b.VaccineRecordusecase.CreateVaccinationRecords(vaccineRecord)

	log.Println("Request Processing Complete", result)
	result = append(result, insertionRecords...)

	//Create report
	reportFile := excelize.NewFile()
	reportShheetName := "Report"
	index, _ := reportFile.NewSheet(reportShheetName)
	reportHeaders := []string{"Student Id", "Drive Id", "Status", "Remarks"}
	for col, header := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1) // (col+1, row=1)
		reportFile.SetCellValue(reportShheetName, cell, header)
	}
	//Keeping track Of Processed File
	totalProcessed := 0
	for i, insertion := range result {
		rowNum := i + 2
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("A%d", rowNum), insertion.Record.StudentId)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("B%d", rowNum), insertion.Record.DriveId)
		if !insertion.Status {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), "Rejected")
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("D%d", rowNum), insertion.ErrorReason)
		} else {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), "Accepted")
			totalProcessed++
		}
	}
	model.ProcessedRecords = totalProcessed
	reportFile.SetActiveSheet(index)
	reportFileName := "Report.xlsx"
	err = reportFile.SaveAs(reportFileName)
	if err != nil {
		log.Println("error creating report file", err.Error(), reportFileName)
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	log.Println("Report File Created", reportFileName)
	//Upload report
	uploadedReportFile, err := b.BulkUploadRepo.UploadFileToMinio(reportFileName, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "reports/", model.RequestId)
	os.Remove(reportFileName)

	if err != nil {
		model.ErrorMessage = "Report File Not Genrated"
		model.Status = "PROCESSED"
		b.BulkUploadRepo.UpdateFileUploadEntry(model)
		return nil
	}
	log.Println("Report File Uploaded", reportFileName)
	model.FilePath = fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), uploadedReportFile)
	//change status
	model.Status = "PROCESSED"
	//update db
	b.BulkUploadRepo.UpdateFileUploadEntry(model)
	log.Println("Processing Complete", model)
	return nil
}

func (b *BulkUploadRequestUsecase) UploadBulkRequestFile(req *models.BulkUploadModel) error {
	var err error
	uploadLoc, err := b.BulkUploadRepo.UploadFileToMinio(req.FilePath, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "uploads/", req.RequestId)
	if err != nil {
		log.Printf("error in uploading to minIo %s", err.Error())
		return err
	}
	log.Println("Uploaded at: ", uploadLoc)
	req.FilePath = uploadLoc
	//Create Entry in DB
	if err = b.BulkUploadRepo.CreateFileUploadEntry(req); err != nil {
		return fmt.Errorf("error in creating bulk upload file entry %s", err.Error())
	}
	//dump in rmq to be picked by async worker
	return b.BulkUploadRepo.SubmitToRabbitMQ(req, "bulk_upload")
}
func NewBulkUploadRequestUsecaseHandler(bulkUploadRepo repository.BulkUploadRepositoryHandler, StudentUsecase usecase.StudentsUsecaseHandler, VaccineRecordusecase usecase.VaccineRecordUsecaseHandler) usecase.BulkUploadRequestHandler {
	return &BulkUploadRequestUsecase{
		BulkUploadRepo:       bulkUploadRepo,
		StudentUsecase:       StudentUsecase,
		VaccineRecordusecase: VaccineRecordusecase,
	}
}
