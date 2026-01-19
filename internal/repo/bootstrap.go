package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// BootstrapSchema ensures required tables and indexes exist in Oracle.
func BootstrapSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`BEGIN
		   EXECUTE IMMEDIATE 'CREATE TABLE users (
		     id VARCHAR2(36) PRIMARY KEY,
		     username VARCHAR2(255) UNIQUE NOT NULL,
		     password_hash VARCHAR2(255) NOT NULL,
		     created_time TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL
		   )';
		 EXCEPTION
		   WHEN OTHERS THEN
		     IF SQLCODE != -955 THEN RAISE; END IF; -- ORA-00955 name already used
		 END;`,
		`BEGIN
		   EXECUTE IMMEDIATE 'CREATE TABLE patients (
		     id VARCHAR2(36) PRIMARY KEY,
		     full_name VARCHAR2(255) NOT NULL,
		     phone_number VARCHAR2(64),
		     age NUMBER,
		     gender VARCHAR2(50),
		     chief_complaint VARCHAR2(4000),
		     present_history VARCHAR2(4000),
		     medical_history VARCHAR2(4000),
		     observation VARCHAR2(4000),
		     palpation VARCHAR2(4000),
		     examination VARCHAR2(4000),
		     rehab VARCHAR2(4000),
		     diagnosis VARCHAR2(4000),
		     created_time TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
		     updated_time TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
		     last_paid_amount NUMBER,
		     status VARCHAR2(100)
		   )';
		 EXCEPTION
		   WHEN OTHERS THEN
		     IF SQLCODE != -955 THEN RAISE; END IF;
		 END;`,
		`BEGIN
		   EXECUTE IMMEDIATE 'CREATE TABLE payments (
		     id VARCHAR2(36) PRIMARY KEY,
		     patient_id VARCHAR2(36) NOT NULL,
		     amount NUMBER NOT NULL,
		     payment_mode VARCHAR2(100),
		     paid_date DATE,
		     CONSTRAINT fk_payment_patient FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE
		   )';
		 EXCEPTION
		   WHEN OTHERS THEN
		     IF SQLCODE != -955 THEN RAISE; END IF;
		 END;`,
		`BEGIN EXECUTE IMMEDIATE 'ALTER TABLE patients ADD (owner_username VARCHAR2(255))';
		 EXCEPTION WHEN OTHERS THEN IF SQLCODE != -01430 THEN NULL; END IF; END;`,
		`BEGIN EXECUTE IMMEDIATE 'ALTER TABLE payments ADD (owner_username VARCHAR2(255))';
		 EXCEPTION WHEN OTHERS THEN IF SQLCODE != -01430 THEN NULL; END IF; END;`,
		`BEGIN EXECUTE IMMEDIATE 'UPDATE patients SET owner_username = ''dency'' WHERE owner_username IS NULL';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'UPDATE payments SET owner_username = ''dency'' WHERE owner_username IS NULL';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`CREATE INDEX idx_patients_phone ON patients(phone_number)`,
		`CREATE INDEX idx_payments_patient ON payments(patient_id)`,
		`BEGIN EXECUTE IMMEDIATE 'DROP INDEX idx_payments_unique_id';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'ALTER TABLE payments DROP COLUMN unique_payment_id';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'ALTER TABLE payments DROP COLUMN created_time';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'ALTER TABLE patients DROP COLUMN exercise_table_json';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
		`BEGIN EXECUTE IMMEDIATE 'ALTER TABLE patients DROP COLUMN exercise_table_raw';
		 EXCEPTION WHEN OTHERS THEN NULL; END;`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			// ignore ORA-00955 and similar "name already used" errors
			if !isNameExistsError(err) {
				return fmt.Errorf("bootstrap failed: %w", err)
			}
		}
	}
	return nil
}

func isNameExistsError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "ORA-00955") || strings.Contains(msg, "ORA-01408")
}
