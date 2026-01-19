package repo

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"phsio_track_backend/internal/core"
)

type PatientRepo struct {
	db *sql.DB
}

func NewPatientRepo(db *sql.DB) *PatientRepo {
	return &PatientRepo{db: db}
}

func (r *PatientRepo) Create(ctx context.Context, owner string, p *core.Patient) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}

	created := p.CreatedTime
	if created.IsZero() {
		created = time.Now()
	}
	updated := p.UpdatedTime
	if updated.IsZero() {
		updated = created
	}
	p.Status = "ACTIVE"

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO patients (
			id, full_name, phone_number, age, gender, chief_complaint, present_history,
			medical_history, observation, palpation, examination, rehab, diagnosis, created_time, updated_time, last_paid_amount, status, owner_username
		) VALUES (
			:1,:2,:3,:4,:5,:6,:7,:8,:9,:10,:11,:12,:13,:14,:15,:16,:17,:18
		)
	`,
		p.ID, p.FullName, p.PhoneNumber, p.Age, p.Gender, p.ChiefComplaint, p.PresentHistory,
		p.MedicalHistory, p.Observation, p.Palpation, p.Examination, p.Rehab, p.Diagnosis, created, updated,
		p.LastPaidAmount, p.Status, owner,
	)
	if err != nil {
		return err
	}
	p.CreatedTime = created
	p.UpdatedTime = updated
	return nil
}

func (r *PatientRepo) List(ctx context.Context, owner string) ([]core.Patient, error) {
	var items []core.Patient
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, full_name, phone_number, age, gender, chief_complaint, present_history,
		       medical_history, observation, palpation, examination, rehab, diagnosis, created_time, updated_time, last_paid_amount, status, owner_username
		FROM patients
		WHERE owner_username=:1
		ORDER BY created_time DESC
	`, owner)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var p core.Patient
		var phone, gender, chief, present, medical, observation, palpation, examination, rehab, diagnosis, status, ownerName sql.NullString
		var age sql.NullInt64
		var lastPaid sql.NullFloat64
		if err := rows.Scan(
			&p.ID, &p.FullName, &phone, &age, &gender, &chief, &present,
			&medical, &observation, &palpation, &examination, &rehab, &diagnosis, &p.CreatedTime, &p.UpdatedTime, &lastPaid, &status, &ownerName,
		); err != nil {
			return make([]core.Patient, 0), err
		}
		p.PhoneNumber = nullStringToString(phone)
		p.Age = nullIntToInt(age)
		p.Gender = nullStringToString(gender)
		p.ChiefComplaint = nullStringToString(chief)
		p.PresentHistory = nullStringToString(present)
		p.MedicalHistory = nullStringToString(medical)
		p.Observation = nullStringToString(observation)
		p.Palpation = nullStringToString(palpation)
		p.Examination = nullStringToString(examination)
		p.Rehab = nullStringToString(rehab)
		p.Diagnosis = nullStringToString(diagnosis)
		p.LastPaidAmount = nullFloatToFloat(lastPaid)
		p.Status = nullStringToString(status)
		p.OwnerUsername = nullStringToString(ownerName)
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PatientRepo) GetByID(ctx context.Context, owner, id string) (core.Patient, error) {
	var p core.Patient
	var phone, gender, chief, present, medical, observation, palpation, examination, rehab, diagnosis, status, ownerName sql.NullString
	var age sql.NullInt64
	var lastPaid sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT id, full_name, phone_number, age, gender, chief_complaint, present_history,
		       medical_history, observation, palpation, examination, rehab, diagnosis, created_time, updated_time, last_paid_amount, status, owner_username
		FROM patients
		WHERE id=:1 AND owner_username=:2
	`, id, owner).Scan(
		&p.ID, &p.FullName, &phone, &age, &gender, &chief, &present,
		&medical, &observation, &palpation, &examination, &rehab, &diagnosis,
		&p.CreatedTime, &p.UpdatedTime, &lastPaid, &status, &ownerName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrNotFound
		}
		return p, err
	}
	p.PhoneNumber = nullStringToString(phone)
	p.Age = nullIntToInt(age)
	p.Gender = nullStringToString(gender)
	p.ChiefComplaint = nullStringToString(chief)
	p.PresentHistory = nullStringToString(present)
	p.MedicalHistory = nullStringToString(medical)
	p.Observation = nullStringToString(observation)
	p.Palpation = nullStringToString(palpation)
	p.Examination = nullStringToString(examination)
	p.Rehab = nullStringToString(rehab)
	p.Diagnosis = nullStringToString(diagnosis)
	p.LastPaidAmount = nullFloatToFloat(lastPaid)
	p.Status = nullStringToString(status)
	p.OwnerUsername = nullStringToString(ownerName)
	return p, nil
}

func (r *PatientRepo) Update(ctx context.Context, owner, id string, upd *core.PatientUpdate) (core.Patient, error) {
	sets := []string{}
	args := []interface{}{}

	add := func(cond bool, expr string, val interface{}) {
		if cond {
			sets = append(sets, expr)
			args = append(args, val)
		}
	}

	if upd.FullName != nil {
		add(true, "full_name=:%d", *upd.FullName)
	}
	if upd.PhoneNumber != nil {
		add(true, "phone_number=:%d", *upd.PhoneNumber)
	}
	if upd.Age != nil {
		add(true, "age=:%d", *upd.Age)
	}
	if upd.Gender != nil {
		add(true, "gender=:%d", *upd.Gender)
	}
	if upd.ChiefComplaint != nil {
		add(true, "chief_complaint=:%d", *upd.ChiefComplaint)
	}
	if upd.PresentHistory != nil {
		add(true, "present_history=:%d", *upd.PresentHistory)
	}
	if upd.MedicalHistory != nil {
		add(true, "medical_history=:%d", *upd.MedicalHistory)
	}
	if upd.Observation != nil {
		add(true, "observation=:%d", *upd.Observation)
	}
	if upd.Palpation != nil {
		add(true, "palpation=:%d", *upd.Palpation)
	}
	if upd.Examination != nil {
		add(true, "examination=:%d", *upd.Examination)
	}
	if upd.Rehab != nil {
		add(true, "rehab=:%d", *upd.Rehab)
	}
	if upd.Diagnosis != nil {
		add(true, "diagnosis=:%d", *upd.Diagnosis)
	}
	if upd.LastPaidAmount != nil {
		add(true, "last_paid_amount=:%d", *upd.LastPaidAmount)
	}
	if upd.Status != nil {
		add(true, "status=:%d", *upd.Status)
	}

	if len(sets) == 0 {
		// nothing to update
		return r.GetByID(ctx, owner, id)
	}

	// add updated_time
	sets = append(sets, "updated_time=SYSTIMESTAMP")

	// build query with parameter indexes
	for i := range sets {
		sets[i] = strings.Replace(sets[i], "%d", strconv.Itoa(i+1), 1)
	}
	args = append(args, id)
	args = append(args, owner)
	idPos := len(args) - 1
	ownerPos := len(args)
	q := "UPDATE patients SET " + strings.Join(sets, ", ") + " WHERE id=:" + strconv.Itoa(idPos) + " AND owner_username=:" + strconv.Itoa(ownerPos)
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return core.Patient{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return core.Patient{}, ErrNotFound
	}
	return r.GetByID(ctx, owner, id)
}

func nullableJSON(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return b
}

func nullableText(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullIntToInt(ni sql.NullInt64) int {
	if ni.Valid {
		return int(ni.Int64)
	}
	return 0
}

func nullFloatToFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}
