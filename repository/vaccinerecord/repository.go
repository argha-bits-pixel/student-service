package vaccinerecord

import (
	"student-service/adapters/mysql"
	"student-service/models"
	"student-service/repository"
	"student-service/requests"
)

type VaccineRecord struct {
	DB *mysql.MysqlConnect
}

func (r *VaccineRecord) CreateVaccinationRecord(record *[]models.VaccineRecord) []models.VaccineInsertionRecord {
	insertionDetails := []models.VaccineInsertionRecord{}
	for _, j := range *record {
		insertionDetail := models.VaccineInsertionRecord{}
		err := r.DB.Table("vaccination_records").Create(&j).Error
		insertionDetail.Record = j
		insertionDetail.Status = true
		if err != nil {
			insertionDetail.Status = false
			insertionDetail.ErrorReason = err.Error()
		}
		insertionDetails = append(insertionDetails, insertionDetail)
	}
	return insertionDetails
}
func (r *VaccineRecord) GetStudentVaccinationRecord(filter string, pagination requests.Pagination) ([]models.StudentVaccinationDetail, error) {
	insertionDetails := []models.StudentVaccinationDetail{}
	if filter == "" {
		if pagination.Limit == 0 {
			return insertionDetails, r.DB.Table("students s").
				Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
				Joins("LEFT JOIN vaccination_records v ON s.id = v.student_id").
				Find(&insertionDetails).Error
		}
		return insertionDetails, r.DB.Table("students s").
			Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
			Order("id ASC").
			Joins("LEFT JOIN vaccination_records v ON s.id = v.student_id").
			Limit(pagination.Limit).
			Offset(pagination.Offset).
			Find(&insertionDetails).Error
	}
	if pagination.Limit == 0 {
		return insertionDetails, r.DB.Table("students s").
			Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
			Joins("LEFT JOIN vaccination_records v ON s.id = v.student_id").
			Where(filter).
			Find(&insertionDetails).Error
	}
	return insertionDetails, r.DB.Table("students s").
		Select("s.id AS id, s.name, s.class, s.roll_number as roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
		Order("id ASC").
		Joins("LEFT JOIN vaccination_records v ON s.id = v.student_id").
		Where(filter).
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&insertionDetails).Error
}
func (r *VaccineRecord) GetStudentVaccinationRecordCount(filter, join string) (int, error) {
	insertionDetails := 0
	if filter == "" {
		return insertionDetails, r.DB.Table("students s").
			Select("s.id AS id, s.name, s.class, s.roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
			Joins(join).
			Count(&insertionDetails).Error
	}
	return insertionDetails, r.DB.Table("students s").
		Select("s.id AS id, s.name, s.class, s.roll_number,s.gender,s.phone_no, v.drive_id as drive_id").
		Joins(join).
		Where(filter).
		Count(&insertionDetails).Error
}
func NewVaccineRecordRepositoryHandler(DB *mysql.MysqlConnect) repository.VaccineRecordRepositoryHandler {
	return &VaccineRecord{
		DB: DB,
	}
}
