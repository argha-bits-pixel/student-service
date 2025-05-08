package models

import "time"

type VaccineRecord struct {
	Id        int        `json:"id"`
	StudentId int        `json:"student_id"`
	DriveId   int        `json:"drive_id"`
	CreatedAt *time.Time `json:"created_at"`
}
type VaccineInsertionRecord struct {
	Record      VaccineRecord `json:"record"`
	Status      bool          `json:"status"`
	ErrorReason string        `json:"error_reason"`
}

type StudentVaccinationDetail struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Class      string `json:"class"`
	Gender     string `json:"gender"`
	RollNumber string `json:"roll_number"`
	PhoneNo    string `json:"phone_no"`
	DriveId    int    `json:"drive_id"`
}
