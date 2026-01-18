$ErrorActionPreference = "Stop"

# Paths (adjust if needed)
$instantClient = "C:\Users\parth\OneDrive\Desktop\Parth\Code\instantclient_23_0"
$gccBin        = "C:\msys64\ucrt64\bin"
$walletDir     = "C:\oracle\wallet"

# App env
$env:Path          = "$instantClient;$gccBin;" + $env:Path
$env:CGO_ENABLED   = "1"
$env:TNS_ADMIN     = $walletDir
$env:DB_USER       = "ADMIN"
$env:DB_PASSWORD   = "Amizara@2000"
$env:DB_CONNECT_STRING = "sf1qflnhz887u1f0_tp"
$env:APP_ENV       = "production"
$env:PORT          = "8080"
$env:JWT_SECRET    = "change-me"
$env:JWT_ISSUER    = "phsio-track"
$env:JWT_EXPIRY_MIN = "60"

Write-Host "Using instant client at $instantClient"
Write-Host "Using wallet at $walletDir"
Write-Host "Starting server on port $($env:PORT)..."

go run app/api/main.go
