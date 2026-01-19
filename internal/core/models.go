package core

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID           string    `json:"id,omitempty"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedTime  time.Time `json:"created_time,omitempty"`
}

type Patient struct {
	ID             string    `json:"id"`
	FullName       string    `json:"full_name"`
	PhoneNumber    string    `json:"phone_number"`
	Age            int       `json:"age"`
	Gender         string    `json:"gender"`
	ChiefComplaint string    `json:"chief_complaint"`
	PresentHistory string    `json:"present_history"`
	MedicalHistory string    `json:"medical_history"`
	Observation    string    `json:"observation"`
	Palpation      string    `json:"palpation"`
	Examination    string    `json:"examination"`
	Rehab          string    `json:"rehab"`
	Diagnosis      string    `json:"diagnosis"`
	CreatedTime    time.Time `json:"created_time,omitempty"`
	UpdatedTime    time.Time `json:"updated_time,omitempty"`
	LastPaidAmount float64   `json:"last_paid_amount"`
	Status         string    `json:"status"`
	OwnerUsername  string    `json:"-"`
}

type PatientUpdate struct {
	FullName       *string  `json:"full_name,omitempty"`
	PhoneNumber    *string  `json:"phone_number,omitempty"`
	Age            *int     `json:"age,omitempty"`
	Gender         *string  `json:"gender,omitempty"`
	ChiefComplaint *string  `json:"chief_complaint,omitempty"`
	PresentHistory *string  `json:"present_history,omitempty"`
	MedicalHistory *string  `json:"medical_history,omitempty"`
	Observation    *string  `json:"observation,omitempty"`
	Palpation      *string  `json:"palpation,omitempty"`
	Examination    *string  `json:"examination,omitempty"`
	Rehab          *string  `json:"rehab,omitempty"`
	Diagnosis      *string  `json:"diagnosis,omitempty"`
	LastPaidAmount *float64 `json:"last_paid_amount,omitempty"`
	Status         *string  `json:"status,omitempty"`
}

type Payment struct {
	ID            string   `json:"id"`
	PatientID     string   `json:"patient_id"`
	Amount        float64  `json:"amount"`
	Mode          string   `json:"mode"`
	Date          JSONTime `json:"date"`
	OwnerUsername string   `json:"-"`
}

type PaymentUpdate struct {
	Amount *float64  `json:"amount,omitempty"`
	Mode   *string   `json:"mode,omitempty"`
	Date   *JSONTime `json:"date,omitempty"`
}

// JSONTime supports flexible JSON parsing (RFC3339 or "2006-01-02T15:04:05").
type JSONTime struct {
	time.Time
}

// NewJSONTime wraps a time.Time.
func NewJSONTime(t time.Time) JSONTime { return JSONTime{Time: t} }

// UnmarshalJSON parses JSON strings into time.Time with flexible layouts.
func (jt *JSONTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		jt.Time = time.Time{}
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05", // no zone
		"2006-01-02",          // date only
	}
	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.Parse(layout, s)
		if err == nil {
			jt.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("parse time: %w", err)
}

// MarshalJSON outputs RFC3339 (UTC).
func (jt JSONTime) MarshalJSON() ([]byte, error) {
	if jt.Time.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + jt.Time.UTC().Format(time.RFC3339) + `"`), nil
}

// Scan implements sql.Scanner.
func (jt *JSONTime) Scan(src interface{}) error {
	if src == nil {
		jt.Time = time.Time{}
		return nil
	}
	switch v := src.(type) {
	case time.Time:
		jt.Time = v
		return nil
	default:
		return fmt.Errorf("cannot scan %T into JSONTime", src)
	}
}

// Value implements driver.Valuer.
func (jt JSONTime) Value() (driver.Value, error) {
	if jt.Time.IsZero() {
		return nil, nil
	}
	return jt.Time, nil
}
