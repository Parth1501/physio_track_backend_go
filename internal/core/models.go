package core

import "time"

type User struct {
	ID           string    `json:"id,omitempty"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedTime  time.Time `json:"created_time,omitempty"`
}

type Patient struct {
	ID                string      `json:"id"`
	FullName          string      `json:"full_name"`
	PhoneNumber       string      `json:"phone_number"`
	Age               int         `json:"age"`
	Gender            string      `json:"gender"`
	ChiefComplaint    string      `json:"chief_complaint"`
	PresentHistory    string      `json:"present_history"`
	MedicalHistory    string      `json:"medical_history"`
	Observation       string      `json:"observation"`
	Palpation         string      `json:"palpation"`
	Examination       string      `json:"examination"`
	Rehab             string      `json:"rehab"`
	Diagnosis         string      `json:"diagnosis"`
	CreatedTime       time.Time   `json:"created_time,omitempty"`
	UpdatedTime       time.Time   `json:"updated_time,omitempty"`
	LastPaidAmount    float64     `json:"last_paid_amount"`
	Status            string      `json:"status"`
	OwnerUsername     string      `json:"-"`
}

type PatientUpdate struct {
	FullName          *string      `json:"full_name,omitempty"`
	PhoneNumber       *string      `json:"phone_number,omitempty"`
	Age               *int         `json:"age,omitempty"`
	Gender            *string      `json:"gender,omitempty"`
	ChiefComplaint    *string      `json:"chief_complaint,omitempty"`
	PresentHistory    *string      `json:"present_history,omitempty"`
	MedicalHistory    *string      `json:"medical_history,omitempty"`
	Observation       *string      `json:"observation,omitempty"`
	Palpation         *string      `json:"palpation,omitempty"`
	Examination       *string      `json:"examination,omitempty"`
	Rehab             *string      `json:"rehab,omitempty"`
	Diagnosis         *string      `json:"diagnosis,omitempty"`
	LastPaidAmount    *float64     `json:"last_paid_amount,omitempty"`
	Status            *string      `json:"status,omitempty"`
}

type Payment struct {
	ID              string    `json:"id"`
	PatientID       string    `json:"patient_id"`
	Amount          float64   `json:"amount"`
	Mode            string    `json:"mode"`
	Date            time.Time `json:"date"`
	OwnerUsername   string    `json:"-"`
}

type PaymentUpdate struct {
	Amount *float64   `json:"amount,omitempty"`
	Mode   *string    `json:"mode,omitempty"`
	Date   *time.Time `json:"date,omitempty"`
}
