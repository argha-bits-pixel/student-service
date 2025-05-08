package models

import "time"

type Student struct {
	Id         int        `json:"id"`
	Name       string     `json:"name"`
	Class      string     `json:"class"`
	Gender     string     `json:"gender"`
	RollNumber string     `json:"roll_number"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"update_at"`
	PhoneNo    string     `json:"phone_no"`
}
type InsertionRecord struct {
	Record      Student `json:"record"`
	Status      bool    `json:"status"`
	ErrorReason string  `json:"error_reason"`
}

type GetStudentDetails struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Class       string `json:"class"`
	Gender      string `json:"gender"`
	RollNo      string `json:"roll_no"`
	PhoneNo     string `json:"phone_no"`
	Vaccination bool   `json:"vaccination"`
	VaccineName string `json:"vaccine_name,omitempty"`
	VaccineDate string `json:"vaccine_date,omitempty"`
}
