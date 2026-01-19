package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"

	"phsio_track_backend/internal/core"
	"phsio_track_backend/internal/http/handlers"
	"phsio_track_backend/internal/repo"
)

// seed_from_sheet ingests CSV exports of the legacy Google Sheets.
// Usage:
// go run tools/seed_from_sheet.go --dsn "postgres://..." --details-csv details.csv --payments-csv payments.csv --admin-user admin --admin-pass secret
func main() {
	var (
		dbUser          string
		dbPass          string
		dbConnectString string
		tnsAdmin        string
		detailsPath     string
		paymentsPath    string
		detailsSheet    string
		paymentsSheet   string
		adminUser       string
		adminPass       string
		paymentsOnly    bool
		updateTimesOnly bool
		ownerUsername   string
	)

	flag.StringVar(&dbUser, "db-user", "", "oracle db user")
	flag.StringVar(&dbPass, "db-pass", "", "oracle db password")
	flag.StringVar(&dbConnectString, "db-connect-string", "", "oracle TNS alias (e.g., sf1qflnhz887u1f0_high)")
	flag.StringVar(&tnsAdmin, "tns-admin", "", "path to wallet directory")
	flag.StringVar(&detailsPath, "details-xlsx", "PATIENT_DETAILS.xlsx", "path to details XLSX")
	flag.StringVar(&paymentsPath, "payments-xlsx", "", "path to payments XLSX (optional)")
	flag.StringVar(&detailsSheet, "details-sheet", "details", "sheet name for patient details")
	flag.StringVar(&paymentsSheet, "payments-sheet", "payment", "sheet name for payments")
	flag.StringVar(&adminUser, "admin-user", "dency", "admin username")
	flag.StringVar(&adminPass, "admin-pass", "Dency@1121", "admin password")
	flag.BoolVar(&paymentsOnly, "payments-only", false, "import payments only (skip patients)")
	flag.BoolVar(&updateTimesOnly, "update-times-only", false, "update created_time/updated_time from details sheet only")
	flag.StringVar(&ownerUsername, "owner-username", "dency", "owner username to stamp on records")
	flag.Parse()

	if dbUser == "" || dbPass == "" || dbConnectString == "" || tnsAdmin == "" || detailsPath == "" {
		fmt.Println("db-user, db-pass, db-connect-string, tns-admin, details-xlsx are required")
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := repo.NewDB(ctx, repo.DBConfig{
		User:          dbUser,
		Password:      dbPass,
		ConnectString: dbConnectString,
		TNSAdmin:      tnsAdmin,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userRepo := repo.NewUserRepo(db)
	patientRepo := repo.NewPatientRepo(db)
	paymentRepo := repo.NewPaymentRepo(db)

	if err := seedAdmin(ctx, userRepo, adminUser, adminPass); err != nil {
		panic(err)
	}

	if updateTimesOnly {
		if err := updatePatientTimes(ctx, db, detailsPath, detailsSheet); err != nil {
			panic(err)
		}
		fmt.Println("Updated patient created_time/updated_time from sheet")
		return
	}

	if !paymentsOnly {
		if err := importDetails(ctx, patientRepo, ownerUsername, detailsPath, detailsSheet); err != nil {
			panic(err)
		}
	}

	if paymentsPath != "" {
		if err := importPayments(ctx, paymentRepo, ownerUsername, paymentsPath, paymentsSheet); err != nil {
			panic(err)
		}
	}

	fmt.Println("Import completed")
}

func seedAdmin(ctx context.Context, repo *repo.UserRepo, username, password string) error {
	return handlers.SeedUser(repo, username, password)
}

func importDetails(ctx context.Context, repo *repo.PatientRepo, owner string, path, sheet string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return fmt.Errorf("no data rows in %s", path)
	}

	for i, row := range rows[1:] {
		if len(row) < 17 {
			fmt.Printf("skipping row %d: not enough columns\n", i+2)
			continue
		}
		createdAt := parseSheetDate(get(row, 13))
		updatedAt := parseSheetDate(get(row, 14))
		p := core.Patient{
			ID:             uuidForString(get(row, 0)),
			FullName:       get(row, 1),
			PhoneNumber:    get(row, 2),
			Age:            atoi(get(row, 3)),
			Gender:         get(row, 4),
			ChiefComplaint: get(row, 5),
			PresentHistory: get(row, 6),
			MedicalHistory: get(row, 7),
			Observation:    get(row, 8),
			Palpation:      get(row, 9),
			Examination:    get(row, 10),
			Rehab:          get(row, 11),
			Diagnosis:      get(row, 12),
			CreatedTime:    createdAt,
			UpdatedTime:    updatedAt,
			LastPaidAmount: atof(get(row, 15)),
			Status:         get(row, 16),
		}
		if err := repo.Create(ctx, owner, &p); err != nil {
			fmt.Printf("error row %d: %v\n", i+2, err)
		}
	}
	return nil
}

func importPayments(ctx context.Context, paymentRepo *repo.PaymentRepo, owner string, path, sheet string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return nil
	}

	for i, row := range rows[1:] {
		if len(row) < 5 {
			fmt.Printf("skipping payment row %d: not enough columns\n", i+2)
			continue
		}
		p := core.Payment{
			ID:        uuidForString(get(row, 1)),
			PatientID: uuidForString(get(row, 0)),
			Amount:    atof(get(row, 2)),
			Mode:      get(row, 3),
			Date:      core.NewJSONTime(parseDate(get(row, 4))),
		}
		if p.PatientID == "" {
			fmt.Printf("skipping payment row %d: empty patient_id\n", i+2)
			continue
		}
		if err := paymentRepo.Upsert(ctx, owner, &p); err != nil {
			if strings.Contains(err.Error(), "ORA-02291") {
				fmt.Printf("skipping payment row %d: patient not found (patient_id=%s)\n", i+2, p.PatientID)
				continue
			}
			fmt.Printf("payment row %d error (patient_id=%s): %v\n", i+2, p.PatientID, err)
		}
	}
	return nil
}

func updatePatientTimes(ctx context.Context, db *sql.DB, path, sheet string) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return err
	}
	if len(rows) < 2 {
		return fmt.Errorf("no data rows in %s", path)
	}

	updated := 0
	for i, row := range rows[1:] {
		if len(row) < 15 {
			fmt.Printf("skipping row %d: not enough columns for times\n", i+2)
			continue
		}
		id := get(row, 0)
		if id == "" {
			continue
		}
		createdAt := parseSheetDate(get(row, 13))
		updatedAt := parseSheetDate(get(row, 14))
		if createdAt.IsZero() && updatedAt.IsZero() {
			continue
		}
		// If one is zero, fallback to the other
		if createdAt.IsZero() {
			createdAt = updatedAt
		}
		if updatedAt.IsZero() {
			updatedAt = createdAt
		}
		_, err := db.ExecContext(ctx, `
			UPDATE patients
			   SET created_time = :1,
			       updated_time = :2
			 WHERE id = :3
		`, createdAt, updatedAt, id)
		if err != nil {
			fmt.Printf("row %d update error: %v\n", i+2, err)
			continue
		}
		updated++
	}
	fmt.Printf("Updated timestamps for %d patients\n", updated)
	return nil
}

func uuidForString(s string) string {
	if s == "" {
		return ""
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(s)).String()
}

func get(row []string, idx int) string {
	if idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	// try YYYY-MM-DD HH:MM:SS
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t
	}
	// try YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	// fallback: parse Excel-like date (dd/mm/yyyy)
	if t, err := time.Parse("02/01/2006", s); err == nil {
		return t
	}
	return time.Time{}
}

// parseSheetDate handles strings like "24/02/2025 16:54DATE" or "24/02/2025 16:54"
func parseSheetDate(s string) time.Time {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "DATE")
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse("02/01/2006 15:04:05", s); err == nil {
		return t
	}
	if t, err := time.Parse("02/01/2006 15:04", s); err == nil {
		return t
	}
	return time.Time{}
}
