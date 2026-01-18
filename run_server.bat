@echo off
setlocal
rem Adjust paths if needed
set INSTANT_CLIENT=C:\Users\parth\OneDrive\Desktop\Parth\Code\instantclient_23_0
set GCC_BIN=C:\msys64\ucrt64\bin
set WALLET_DIR=C:\oracle\wallet

set PATH=%INSTANT_CLIENT%;%GCC_BIN%;%PATH%
set CGO_ENABLED=1
set TNS_ADMIN=%WALLET_DIR%
set DB_USER=ADMIN
set DB_PASSWORD=Amizara@2000
set DB_CONNECT_STRING=sf1qflnhz887u1f0_tp
set APP_ENV=production
set PORT=8080
set JWT_SECRET=change-me
set JWT_ISSUER=phsio-track
set JWT_EXPIRY_MIN=60

echo Using instant client at %INSTANT_CLIENT%
echo Using wallet at %WALLET_DIR%
echo Starting server on port %PORT%...

go run app/api/main.go

endlocal
