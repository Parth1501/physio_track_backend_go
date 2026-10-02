package repo

import (
	"context"
	"database/sql"
	"math"
	"strconv"
	"strings"

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
	if err := r.assertPatientOwner(ctx, owner, p.PatientID); err != nil {
		return err
	}
	p.Mode = strings.ToUpper(strings.TrimSpace(p.Mode))
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
	if err := r.setLastPaid(ctx, owner, p.PatientID, p.Amount); err != nil {
		return err
	}
	return nil
}

// Upsert inserts or updates a payment keyed by id.
func (r *PaymentRepo) Upsert(ctx context.Context, owner string, p *core.Payment) error {
	if err := r.assertPatientOwner(ctx, owner, p.PatientID); err != nil {
		return err
	}
	p.Mode = strings.ToUpper(strings.TrimSpace(p.Mode))
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
	if err := r.setLastPaid(ctx, owner, p.PatientID, p.Amount); err != nil {
		return err
	}
	return nil
}

func (r *PaymentRepo) List(ctx context.Context, owner, patientID string) ([]core.Payment, error) {
	var rows *sql.Rows
	var err error
	if patientID != "" && patientID != "ALL" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, patient_id, amount, payment_mode, paid_date, partial_of
			FROM payments
			WHERE patient_id=:1 AND owner_username=:2
			ORDER BY paid_date DESC
		`, patientID, owner)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, patient_id, amount, payment_mode, paid_date, partial_of
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
		var partialOf sql.NullString
		if err := rows.Scan(&p.ID, &p.PatientID, &p.Amount, &p.Mode, &paid, &partialOf); err != nil {
			return nil, err
		}
		p.PartialOf = partialOf.String
		if paid.Valid {
			p.Date = core.NewJSONTime(paid.Time)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PaymentRepo) Update(ctx context.Context, owner, id string, upd *core.PaymentUpdate) (core.Payment, error) {
	// Ensure payment belongs to a patient owned by requester
	current, err := r.GetByID(ctx, owner, id)
	if err != nil {
		if err == ErrNotFound {
			return core.Payment{}, ErrForbidden
		}
		return core.Payment{}, err
	}
	if err := r.assertPatientOwner(ctx, owner, current.PatientID); err != nil {
		return core.Payment{}, err
	}

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
		fields = append(fields, field{name: "paid_date", val: upd.Date.Time})
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
	updatedPayment, err := r.GetByID(ctx, owner, id)
	if err != nil {
		return core.Payment{}, err
	}
	if err := r.setLastPaid(ctx, owner, updatedPayment.PatientID, updatedPayment.Amount); err != nil {
		return core.Payment{}, err
	}
	return updatedPayment, nil
}

func (r *PaymentRepo) Delete(ctx context.Context, owner, id string) error {
	// Ensure the payment belongs to a patient owned by requester
	p, err := r.GetByID(ctx, owner, id)
	if err != nil {
		return err
	}
	if err := r.assertPatientOwner(ctx, owner, p.PatientID); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// A partial paid row whose pending row is deleted becomes a standalone paid row.
	if _, err := tx.ExecContext(ctx, `UPDATE payments SET partial_of=NULL WHERE partial_of=:1 AND owner_username=:2`, id, owner); err != nil {
		return err
	}
	cmd, err := tx.ExecContext(ctx, `DELETE FROM payments WHERE id=:1 AND owner_username=:2`, id, owner)
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
	return tx.Commit()
}

const amountEpsilon = 0.005

type pendingRow struct {
	id     string
	amount float64
	date   sql.NullTime
}

type partialRow struct {
	id     string
	amount float64
	mode   string
}

// Settle applies amount to the patient's pending payments oldest-first in one transaction.
// A session paid in parts keeps a single paid row, linked to its pending row via partial_of
// until the session is fully paid.
func (r *PaymentRepo) Settle(ctx context.Context, owner, patientID string, amount float64, mode string) error {
	if err := r.assertPatientOwner(ctx, owner, patientID); err != nil {
		return err
	}
	mode = strings.ToUpper(strings.TrimSpace(mode))

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	pending, err := loadPending(ctx, tx, owner, patientID)
	if err != nil {
		return err
	}
	partials, err := loadPartials(ctx, tx, owner, patientID)
	if err != nil {
		return err
	}

	total := 0.0
	for _, p := range pending {
		total += p.amount
	}
	if amount > total+amountEpsilon {
		return ErrExceedsPending
	}

	remaining := amount
	for _, p := range pending {
		if remaining <= amountEpsilon {
			break
		}
		pay := math.Min(remaining, p.amount)
		remaining -= pay
		left := p.amount - pay
		fullyPaid := left <= amountEpsilon

		// Paid side is written before the pending side for each session.
		if partner, ok := partials[p.id]; ok {
			var link interface{} = p.id
			if fullyPaid {
				link = nil
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE payments SET amount=:1, payment_mode=:2, partial_of=:3
				WHERE id=:4 AND owner_username=:5
			`, partner.amount+pay, mergedMode(partner, pay, mode), link, partner.id, owner); err != nil {
				return err
			}
			if fullyPaid {
				_, err = tx.ExecContext(ctx, `DELETE FROM payments WHERE id=:1 AND owner_username=:2`, p.id, owner)
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE payments SET amount=:1 WHERE id=:2 AND owner_username=:3`, left, p.id, owner)
			}
			if err != nil {
				return err
			}
		} else if fullyPaid {
			if _, err := tx.ExecContext(ctx, `
				UPDATE payments SET payment_mode=:1 WHERE id=:2 AND owner_username=:3
			`, mode, p.id, owner); err != nil {
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO payments (id, patient_id, amount, payment_mode, paid_date, owner_username, partial_of)
				VALUES (:1,:2,:3,:4,:5,:6,:7)
			`, uuid.NewString(), patientID, pay, mode, p.date, owner, p.id); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE payments SET amount=:1 WHERE id=:2 AND owner_username=:3
			`, left, p.id, owner); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// mergedMode picks the mode of the larger part of a session paid in parts; ONLINE on a tie.
func mergedMode(partner partialRow, pay float64, mode string) string {
	switch {
	case math.Abs(pay-partner.amount) <= amountEpsilon:
		return "ONLINE"
	case pay > partner.amount:
		return mode
	default:
		return partner.mode
	}
}

func loadPending(ctx context.Context, tx *sql.Tx, owner, patientID string) ([]pendingRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, amount, paid_date
		FROM payments
		WHERE patient_id=:1 AND owner_username=:2 AND payment_mode='PENDING'
		ORDER BY paid_date ASC, id ASC
		FOR UPDATE
	`, patientID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []pendingRow
	for rows.Next() {
		var p pendingRow
		if err := rows.Scan(&p.id, &p.amount, &p.date); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func loadPartials(ctx context.Context, tx *sql.Tx, owner, patientID string) (map[string]partialRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, amount, payment_mode, partial_of
		FROM payments
		WHERE patient_id=:1 AND owner_username=:2 AND partial_of IS NOT NULL
		FOR UPDATE
	`, patientID, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := map[string]partialRow{}
	for rows.Next() {
		var p partialRow
		var mode sql.NullString
		var pendingID string
		if err := rows.Scan(&p.id, &p.amount, &mode, &pendingID); err != nil {
			return nil, err
		}
		p.mode = strings.ToUpper(mode.String)
		items[pendingID] = p
	}
	return items, rows.Err()
}

func (r *PaymentRepo) GetByID(ctx context.Context, owner, id string) (core.Payment, error) {
	var p core.Payment
	var paid sql.NullTime
	var partialOf sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, patient_id, amount, payment_mode, paid_date, partial_of
		FROM payments
		WHERE id=:1 AND owner_username=:2
	`, id, owner).Scan(&p.ID, &p.PatientID, &p.Amount, &p.Mode, &paid, &partialOf)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrNotFound
		}
		return p, err
	}
	p.PartialOf = partialOf.String
	if paid.Valid {
		p.Date = core.NewJSONTime(paid.Time)
	}
	return p, nil
}

// assertPatientOwner ensures the patient belongs to the requesting owner.
func (r *PaymentRepo) assertPatientOwner(ctx context.Context, owner, patientID string) error {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM patients
		WHERE id=:1 AND owner_username=:2
	`, patientID, owner).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrForbidden
		}
		return err
	}
	return nil
}

// setLastPaid updates the patient's last_paid_amount.
func (r *PaymentRepo) setLastPaid(ctx context.Context, owner, patientID string, amount float64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE patients
		   SET last_paid_amount = :1
		 WHERE id = :2 AND owner_username = :3
	`, amount, patientID, owner)
	return err
}
