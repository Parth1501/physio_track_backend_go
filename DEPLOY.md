# Deploying to OCI Autonomous (Oracle, wallet/TNS)

1) **Wallet and env**
   - Place wallet dir on server, e.g. `/opt/wallet/Wallet_SF1QFLNHZ887U1F0`.
   - Export:
     ```
     TNS_ADMIN=/opt/wallet/Wallet_SF1QFLNHZ887U1F0
     DB_USER=<oracle_user>
     DB_PASSWORD=<oracle_password>
     DB_CONNECT_STRING=sf1qflnhz887u1f0_high   # from tnsnames.ora
     APP_ENV=production
     PORT=8080
     JWT_SECRET=change-me
     JWT_ISSUER=phsio-track
     JWT_EXPIRY_MIN=60
     ```

2) **Build (Ampere 1 OCPU / 1 GB)**
   - godror requires cgo + a C toolchain.
   ```bash
   CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o api cmd/api/main.go
   ```

3) **Run**
   - Copy binary and wallet to the VM.
   - Systemd example:
     ```
     [Unit]
     Description=phsio-track-api
     After=network.target

     [Service]
     WorkingDirectory=/opt/phsio
     ExecStart=/opt/phsio/api
     Environment=TNS_ADMIN=/opt/wallet/Wallet_SF1QFLNHZ887U1F0
     Environment=DB_USER=<user>
     Environment=DB_PASSWORD=<pass>
     Environment=DB_CONNECT_STRING=sf1qflnhz887u1f0_high
     Environment=APP_ENV=production
     Environment=PORT=8080
     Environment=JWT_SECRET=change-me
     Environment=JWT_ISSUER=phsio-track
     Environment=JWT_EXPIRY_MIN=60
     Restart=on-failure
     User=ubuntu

     [Install]
     WantedBy=multi-user.target
     ```

4) **Startup bootstrap**
   - App connects to Oracle via wallet, auto-creates tables (`users`, `patients`, `payments`), and keeps indexes.
   - Seeds user `dency / Dency@1121` unless it already exists.

5) **Import legacy data (XLSX)**
   ```bash
   go run tools/seed_from_sheet.go \
     --db-user "$DB_USER" \
     --db-pass "$DB_PASSWORD" \
     --db-connect-string "$DB_CONNECT_STRING" \
     --tns-admin "$TNS_ADMIN" \
     --details-xlsx PATIENT_DETAILS.xlsx \
     --payments-xlsx PAYMENTS.xlsx \  # optional if you have payments sheet
     --details-sheet details \
     --payments-sheet payment \
     --admin-user dency \
     --admin-pass "Dency@1121"
   ```
   - Rehab string parsed with `Č` (columns) and `Ɍ` (rows) into `exercise_table_json`; raw string also stored.

6) **API endpoints (Bearer token required except login)**
   - `POST /auth/login`
   - `POST /patients`, `GET /patients`, `GET /patients/:id`, `PATCH /patients/:id`
   - `POST /payments`, `GET /payments?patient_id=...|ALL`, `PATCH /payments/:id`, `DELETE /payments/:id`
   - `POST /auth/login` → `{token}` (use admin creds or seeded user)
   - `POST /patients`, `GET /patients`, `GET /patients/:id`, `PATCH /patients/:id`
   - `POST /payments`, `GET /payments?patient_id=...|ALL`, `PATCH /payments/:id`, `DELETE /payments/:id`

8) **Android client usage**
   - Login once, cache token, send `Authorization: Bearer <token>` header.
   - Send `exercise_table_json` as an array of rows (strings/nulls), no special header row enforced.
