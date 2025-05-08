package vaccinerecord

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"student-service/models"
	"student-service/repository"
	"student-service/requests"
	"student-service/usecase"
	"student-service/utils"

	"github.com/xuri/excelize/v2"
)

type VaccineRecordUsecase struct {
	vaccineRecordRepo repository.VaccineRecordRepositoryHandler
	studentRepo       repository.StudentRepositoryHandler
	bulkUploadrepo    repository.BulkUploadRepositoryHandler
}
type VaccineDriveGetResponse struct {
	Id        int         `json:"id"`
	Vaccine   string      `json:"vaccine_name"`
	DriveDate string      `json:"drive_date"`
	Doses     int         `json:"doses"`
	Classes   string      `json:"classes"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at,omitempty"`
	Links     interface{} `json:"_links,omitempty"`
}
type VaccineDriveResponse struct {
	Data json.RawMessage `json:"data"`
}

func (v *VaccineRecordUsecase) CreateVaccinationRecords(records *[]models.VaccineRecord) []models.VaccineInsertionRecord {
	//verify if drive actually exists
	validRecords := new([]models.VaccineRecord)
	inValidRecords := []models.VaccineInsertionRecord{}
	for _, j := range *records {
		//checking if drive exists
		driveData, err := verifyDriveExists(j.DriveId, "")
		log.Println("drive Data ", driveData, err == nil)
		if err != nil || len(driveData) == 0 {
			invalid := models.VaccineInsertionRecord{
				Record:      j,
				Status:      false,
				ErrorReason: fmt.Sprintf("no drive exists with drive_id : %d", j.DriveId),
			}
			log.Println("invalid due to wrong drive id", invalid)
			inValidRecords = append(inValidRecords, invalid)
			continue
		}
		//check if student is valid
		resp, _ := v.studentRepo.GetStudents(fmt.Sprintf("id = %d", j.StudentId))
		if len(resp) != 1 {
			invalid := models.VaccineInsertionRecord{
				Record:      j,
				Status:      false,
				ErrorReason: fmt.Sprintf("no student exists with student_id : %d", j.StudentId),
			}
			log.Println("invalid due to wrong student id", invalid)
			inValidRecords = append(inValidRecords, invalid)
			continue
		}
		*validRecords = append(*validRecords, j)
	}
	//proceed for insertion
	log.Println("Valid Records", validRecords, "invalidRecords", inValidRecords)
	resp := v.vaccineRecordRepo.CreateVaccinationRecord(validRecords)
	return append(resp, inValidRecords...)
}

func verifyDriveExists(id int, name string) ([]VaccineDriveGetResponse, error) {
	dataSet := VaccineDriveResponse{}
	var reqUrl string
	if name == "" {
		reqUrl = fmt.Sprintf("%s/vaccine/drives/%d", os.Getenv("VACCINE_SERVICE"), id)
	} else {
		reqUrl = fmt.Sprintf("%s/vaccine/drives?vaccine_name=%s", os.Getenv("VACCINE_SERVICE"), url.QueryEscape(name))
	}
	log.Println("Resource Being Requested at", reqUrl)
	code, resp, err := utils.MakeAPICall(http.MethodGet, reqUrl, map[string]string{
		"Content-Type": "application/json",
	}, nil)
	if err != nil {
		log.Println("error in verifying drive api call", err.Error())
		return []VaccineDriveGetResponse{}, err
	}
	log.Println("Response is", code, string(resp), err == nil)
	if code != http.StatusOK {
		log.Println("drive doesn't exist", code, string(resp))
		return []VaccineDriveGetResponse{}, err
	}

	if err := json.Unmarshal(resp, &dataSet); err != nil {
		log.Println("error in unmarshalling drive information", err.Error())
		return []VaccineDriveGetResponse{}, err
	}
	// return dataSet, err
	var drives []VaccineDriveGetResponse
	if err := json.Unmarshal(dataSet.Data, &drives); err == nil {
		return drives, nil
	}
	var drive VaccineDriveGetResponse
	if err := json.Unmarshal(dataSet.Data, &drive); err == nil {
		return []VaccineDriveGetResponse{drive}, nil
	}
	return []VaccineDriveGetResponse{}, errors.New("inavlid drive")
}

func (v *VaccineRecordUsecase) GenerateVaccinationReport(request *requests.GenerateReportRequest) (string, error) {
	var studentDetails []models.GetStudentDetails
	var vaccinationDetails []models.StudentVaccinationDetail
	var err error
	driveRegister := make(map[int]VaccineDriveGetResponse)
	queryString := ""

	if request.VaccineName != "" {
		drive, err := verifyDriveExists(0, request.VaccineName)
		if err != nil {
			log.Println("error in getting vaccine name", err.Error())
			return "", err
		}
		if len(drive) == 0 {
			return "", fmt.Errorf("no data for vaccine name %s", request.VaccineName)
		}
		for _, j := range drive {
			driveRegister[j.Id] = j
		}
		log.Printf("Drive info is as below %+v", driveRegister)
		driveIds := []int{}
		for i, _ := range driveRegister {
			driveIds = append(driveIds, i)
		}
		placeholders := make([]string, len(driveIds))
		fmt.Println(driveIds)
		for i, id := range driveIds {
			placeholders[i] = fmt.Sprintf("%d", id)
		}
		inClause := strings.Join(placeholders, ", ")
		queryString = fmt.Sprintf("v.drive_id IN (%s)", inClause)
	}
	if request.Class != "" {
		query := fmt.Sprintf("s.class = '%s'", request.Class)
		if queryString != "" {
			queryString += " AND " + query
		} else {
			queryString = query
		}
	}
	log.Println("query being used", queryString)
	vaccinationDetails, err = v.vaccineRecordRepo.GetStudentVaccinationRecord(queryString, requests.Pagination{})
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return "", err
	}
	log.Printf("details fetched %+v", vaccinationDetails)
	for _, j := range vaccinationDetails {
		studentDetail := models.GetStudentDetails{}
		studentDetail.Id = j.Id
		studentDetail.Name = j.Name
		studentDetail.Gender = j.Gender
		studentDetail.Class = j.Class
		studentDetail.RollNo = j.RollNumber
		studentDetail.PhoneNo = j.PhoneNo
		if j.DriveId == 0 {
			studentDetail.Vaccination = false
		} else {
			studentDetail.Vaccination = true
			drive, ok := driveRegister[j.Id]
			if !ok {
				var driveInfo []VaccineDriveGetResponse
				driveInfo, err = verifyDriveExists(j.DriveId, "")
				if err != nil {
					log.Println("error fetching vaccination record", err.Error())
					continue
				}
				log.Printf("response from vaccine service %+v", driveInfo)
				driveRegister[j.Id] = driveInfo[0]
				drive = driveInfo[0]
				log.Printf(" drive is %+v drive info is %+v", drive, driveInfo)
			}
			log.Printf(" drive is %+v", drive)
			studentDetail.VaccineName = drive.Vaccine
			studentDetail.VaccineDate = drive.DriveDate
		}
		studentDetails = append(studentDetails, studentDetail)
	}
	//Create report
	reportFile := excelize.NewFile()
	reportShheetName := "Report"
	index, _ := reportFile.NewSheet(reportShheetName)
	reportHeaders := []string{"Name", "Class", "Gender", "Roll Number", "Phone Number", "Vaccination Status", "Vaccine Name", "Vaccination Date"}
	for col, header := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1) // (col+1, row=1)
		reportFile.SetCellValue(reportShheetName, cell, header)
	}
	for i, student := range studentDetails {
		rowNum := i + 2
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("A%d", rowNum), student.Name)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("B%d", rowNum), student.Class)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("C%d", rowNum), student.Gender)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("D%d", rowNum), student.RollNo)
		reportFile.SetCellValue(reportShheetName, fmt.Sprintf("E%d", rowNum), student.PhoneNo)
		if !student.Vaccination {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Non Vaccinated")
		} else {
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("F%d", rowNum), "Vaccinated")
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("G%d", rowNum), student.VaccineName)
			reportFile.SetCellValue(reportShheetName, fmt.Sprintf("H%d", rowNum), student.VaccineDate)
		}
	}
	reportFile.SetActiveSheet(index)
	reportFileName := "Report.xlsx"
	err = reportFile.SaveAs(reportFileName)
	if err != nil {
		log.Println("Unable to save report File Locally", err.Error())
		return "", errors.New("Internal server Error")
	}
	uploadedReportFile, err := v.bulkUploadrepo.UploadFileToMinio(reportFileName, os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), "reports/", request.RequestId)
	if err != nil {
		log.Println("error in uploading file to minio", err.Error())
		return "", err
	}
	filePath := fmt.Sprintf("http://%s:%s/%s/%s", os.Getenv("MINIO_SERVER"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BULK_UPLOAD_BUCKET"), uploadedReportFile)

	return filePath, nil
}

func (v *VaccineRecordUsecase) GetStudentVaccinationRecords(request *requests.GetStudentVaccineRecordRequest) (int, []models.GetStudentDetails, error) {
	var studentDetails []models.GetStudentDetails
	var vaccinationDetails []models.StudentVaccinationDetail
	var err error
	var total int
	driveRegister := make(map[int]VaccineDriveGetResponse)

	joinCondtion := "LEFT JOIN vaccination_records v ON s.id = v.student_id"
	queryString := ""

	//if id is given
	if request.Id != 0 {
		queryString = fmt.Sprintf("s.id = '%d'", request.Id)
	}
	if request.RollNo != "" {
		if queryString == "" {
			queryString = fmt.Sprintf("s.roll_number = '%s'", request.RollNo)
		} else {
			queryString += " AND " + fmt.Sprintf("s.roll_number = '%s'", request.RollNo)
		}
	}
	if request.Class != "" {
		if queryString == "" {
			queryString = fmt.Sprintf("s.class = '%s'", request.Class)
		} else {
			queryString += " AND " + fmt.Sprintf("s.class = '%s'", request.Class)
		}
	}
	if request.Name != "" {
		if queryString == "" {
			queryString = fmt.Sprintf("s.name LIKE '%%%s%%'", request.Name)
		} else {
			queryString += " AND " + fmt.Sprintf("s.name LIKE '%%%s%%'", request.Name)
		}
	}
	if request.VaccineName != "" {
		//get drive info or id by vaccine name
		// var driveInfo VaccineDriveResponse
		drive, err := verifyDriveExists(0, request.VaccineName)
		if err != nil {
			log.Println("error fetching vaccination record", err.Error())
			return total, studentDetails, err
		}
		for _, j := range drive {
			driveRegister[j.Id] = j
		}
		log.Printf("Drive info is as below %+v", driveRegister)
		if len(driveRegister) == 0 {
			return total, studentDetails, fmt.Errorf("no vaccination drive with vaccine : %s", request.VaccineName)
		}
		driveIds := []int{}
		for i, _ := range driveRegister {
			driveIds = append(driveIds, i)
		}
		placeholders := make([]string, len(driveIds))
		for i, id := range driveIds {
			placeholders[i] = fmt.Sprintf("%d", id)
		}
		inClause := strings.Join(placeholders, ", ")
		log.Printf("vaccination_record.drive_id IN (%s)", inClause)
		//get count
		if queryString == "" {
			queryString = fmt.Sprintf("v.drive_id IN (%s)", inClause)
		} else {
			queryString += " AND " + fmt.Sprintf("v.drive_id IN (%s)", inClause)

		}
		//all record scenario
	}
	log.Println("query string being used", queryString)
	total, err = v.vaccineRecordRepo.GetStudentVaccinationRecordCount(queryString, joinCondtion)
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return total, studentDetails, err
	}
	vaccinationDetails, err = v.vaccineRecordRepo.GetStudentVaccinationRecord(queryString, request.Pagination)
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return total, studentDetails, err
	}

	//genrate consolidated Response
	for _, j := range vaccinationDetails {
		studentDetail := models.GetStudentDetails{}
		studentDetail.Id = j.Id
		studentDetail.Name = j.Name
		studentDetail.Gender = j.Gender
		studentDetail.Class = j.Class
		studentDetail.RollNo = j.RollNumber
		studentDetail.PhoneNo = j.PhoneNo
		if j.DriveId == 0 {
			studentDetail.Vaccination = false
		} else {
			studentDetail.Vaccination = true
			drive, ok := driveRegister[j.Id]
			if !ok {
				var driveInfo []VaccineDriveGetResponse
				driveInfo, err = verifyDriveExists(j.DriveId, "")
				if err != nil {
					log.Println("error fetching vaccination record", err.Error())
					continue
				}
				log.Printf("response from vaccine service %+v", driveInfo)
				driveRegister[j.Id] = driveInfo[0]
				drive = driveInfo[0]
				log.Printf(" drive is %+v drive info is %+v", drive, driveInfo)
			}
			log.Printf(" drive is %+v", drive)
			studentDetail.VaccineName = drive.Vaccine
			studentDetail.VaccineDate = drive.DriveDate
		}
		studentDetails = append(studentDetails, studentDetail)
	}
	return total, studentDetails, err
}

func (v *VaccineRecordUsecase) GetStudentVaccinationDashBoard() (int, int, error) {
	totalStudents, err := v.vaccineRecordRepo.GetStudentVaccinationRecordCount("", "LEFT JOIN vaccination_records v ON s.id = v.student_id")
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
		return totalStudents, 0, err
	}
	vaccnatedStudents, err := v.vaccineRecordRepo.GetStudentVaccinationRecordCount("", "INNER JOIN vaccination_records v ON s.id = v.student_id")
	if err != nil {
		log.Println("error fetching vaccination record", err.Error())
	}
	return totalStudents, vaccnatedStudents, err

}

// func createDriveRegist

func NewVaccineRecordUsecaseHandler(vaccRepo repository.VaccineRecordRepositoryHandler, studentRepo repository.StudentRepositoryHandler, bulkUploadRepo repository.BulkUploadRepositoryHandler) usecase.VaccineRecordUsecaseHandler {
	return &VaccineRecordUsecase{vaccineRecordRepo: vaccRepo, studentRepo: studentRepo, bulkUploadrepo: bulkUploadRepo}
}
