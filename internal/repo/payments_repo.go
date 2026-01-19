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

type PaymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(ctx context.Context, owner string, p *core.Payment) error {
	p.Mode = strings.ToUpper(strings.TrimSpace(p.Mode))
	// Normalize date: if provided without zone, assume local and convert to UTC for storage consistency
	if !p.Date.Time.IsZero() {
		p.Date = core.NewJSONTime(ensureUTC(p.Date.Time))
	}
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO payments (id, patient_id, amount, payment_mode, paid_date, owner_username)
		VALUES (:1,:2,:3,:4,:5,:6)
	`, p.ID, p.PatientID, p.Amount, p.Mode, p.Date, owner)
	if err != nil {
		return err
	}
	return nil
}

// Upsert inserts or updates a payment keyed by id.
func (r *PaymentRepo) Upsert(ctx context.Context, owner string, p *core.Payment) error {
	p.Mode = strings.ToUpper(strings.TrimSpace(p.Mode))
	if !p.Date.Time.IsZero() {
		p.Date = core.NewJSONTime(ensureUTC(p.Date.Time))
	}
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx, `
		MERGE INTO payments t
		USING (SELECT :1 AS id,
		              :2 AS patient_id,
		              :3 AS amount,
		              :4 AS payment_mode,
		              :5 AS paid_date,
		              :6 AS owner_username
		       FROM dual) s
		ON (t.id = s.id AND t.owner_username = s.owner_username)
		WHEN MATCHED THEN
		  UPDATE SET t.amount = s.amount,
		             t.payment_mode = s.payment_mode,
		             t.paid_date = s.paid_date,
		             t.patient_id = s.patient_id
		WHEN NOT MATCHED THEN
		  INSERT (id, patient_id, amount, payment_mode, paid_date, owner_username)
		  VALUES (s.id, s.patient_id, s.amount, s.payment_mode, s.paid_date, s.owner_username)
	`, p.ID, p.PatientID, p.Amount, p.Mode, p.Date, owner)
	if err != nil {
		return err
	}
	return nil
}

func (r *PaymentRepo) List(ctx context.Context, owner, patientID string) ([]core.Payment, error) {
	var rows *sql.Rows
	var err error
	if patientID != "" && patientID != "ALL" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, patient_id, amount, payment_mode, paid_date
			FROM payments
			WHERE patient_id=:1 AND owner_username=:2
			ORDER BY paid_date DESC
		`, patientID, owner)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, patient_id, amount, payment_mode, paid_date
			FROM payments
			WHERE owner_username=:1
			ORDER BY paid_date DESC
		`, owner)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []core.Payment
	for rows.Next() {
		var p core.Payment
		var paid sql.NullTime
		if err := rows.Scan(&p.ID, &p.PatientID, &p.Amount, &p.Mode, &paid); err != nil {
			return nil, err
		}
		if paid.Valid {
			p.Date = core.NewJSONTime(paid.Time)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PaymentRepo) Update(ctx context.Context, owner, id string, upd *core.PaymentUpdate) (core.Payment, error) {
	// Build update set
	type field struct {
		name string
		val  interface{}
	}
	fields := []field{}
	if upd.Amount != nil {
		fields = append(fields, field{name: "amount", val: *upd.Amount})
	}
	if upd.Mode != nil {
		mode := strings.ToUpper(strings.TrimSpace(*upd.Mode))
		fields = append(fields, field{name: "payment_mode", val: mode})
	}
	if upd.Date != nil {
		fields = append(fields, field{name: "paid_date", val: ensureUTC(upd.Date.Time)})
	}
	if len(fields) == 0 {
		return r.GetByID(ctx, owner, id)
	}

	args := []interface{}{}
	setClauses := ""
	for i, f := range fields {
		if i > 0 {
			setClauses += ", "
		}
		setClauses += f.name + "=:" + strconv.Itoa(i+1)
		args = append(args, f.val)
	}
	args = append(args, id, owner)

	q := "UPDATE payments SET " + setClauses + " WHERE id=:" + strconv.Itoa(len(args)-1) + " AND owner_username=:" + strconv.Itoa(len(args))
	res, err := r.db.ExecContext(ctx, q, args...)
	if err != nil {
		return core.Payment{}, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return core.Payment{}, ErrNotFound
	}
	return r.GetByID(ctx, owner, id)
}

func (r *PaymentRepo) Delete(ctx context.Context, owner, id string) error {
	cmd, err := r.db.ExecContext(ctx, `DELETE FROM payments WHERE id=:1 AND owner_username=:2`, id, owner)
	if err != nil {
		return err
	}
	rows, err := cmd.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PaymentRepo) GetByID(ctx context.Context, owner, id string) (core.Payment, error) {
	var p core.Payment
	var paid sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, patient_id, amount, payment_mode, paid_date
		FROM payments
		WHERE id=:1 AND owner_username=:2
	`, id, owner).Scan(&p.ID, &p.PatientID, &p.Amount, &p.Mode, &paid)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrNotFound
		}
		return p, err
	}
	if paid.Valid {
		p.Date = core.NewJSONTime(paid.Time)
	}
	return p, nil
}

// ensureUTC normalizes times to UTC if they have no location.
func ensureUTC(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	if t.Location() == time.UTC {
		return t
	}
	if t.Location() == time.Local || t.Location() == nil {
		return t.UTC()
	}
	return t.In(time.UTC)
}
