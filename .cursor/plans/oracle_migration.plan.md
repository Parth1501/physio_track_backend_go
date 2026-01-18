# Plan: Switch backend to Oracle Autonomous (wallet/TNS)

## Goals

- Connect to Autonomous Oracle using wallet/TNS on startup (godror).
- Replace Postgres schema/queries with Oracle equivalents.
- Auto-create tables if missing; seed user dency/Dency@1121.
- Migrate data from `PATIENT_DETAILS.xlsx` (exercise table JSON/raw preserved).
- Keep REST API surface unchanged.

## Steps

1) Dependencies & config

- Add `github.com/godror/godror` and `github.com/xuri/excelize/v2`.
- Config vars: `TNS_ADMIN` (wallet dir), `DB_USER`, `DB_PASSWORD`, `DB_CONNECT_STRING` (e.g., `sf1qflnhz887u1f0_high`), `JWT_*`, `PORT`.
- Update `config.Load` to read these and expose DSN pieces.
2) DB connect & bootstrap
- Replace `repo.NewDB` to use godror pool with TNS alias + wallet.
- On startup, run Oracle DDL to create tables/indexes if absent (`users`, `patients`, `payments`).
- Adjust ID storage to `VARCHAR2(36)` with Go-generated UUIDs; timestamps as `TIMESTAMP DEFAULT SYSTIMESTAMP`; JSON as `JSON`.
3) Repositories (Oracle SQL)
- Rewrite queries with `:1` binds and Oracle functions (e.g., `SYSTIMESTAMP`).
- Ensure JSON columns handled via `godror` (bind/read as `[]byte`).
- Keep behavior: patients/payments CRUD, soft requirements unchanged.
4) Importer (XLSX)
- Replace CSV reader with `excelize` to read `PATIENT_DETAILS.xlsx` (and payments sheet if present).
- Parse rehab string with `Č`/`Ɍ` into `exercise_table_json`; store raw.
- Seed user dency/Dency@1121 via repo call.
5) Deployment docs
- Update `DEPLOY.md` for Oracle: wallet placement, env vars, build/run, importer usage with XLSX.

## Notes

- Wallet alias options from `tnsnames.ora`: `sf1qflnhz887u1f0_high|medium|low|tp|tpurgent`.
- Empty exercise cells stored as `null`; rows/cols dynamic.