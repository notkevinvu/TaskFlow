@echo off
REM ============================================
REM TaskFlow - Stop Development Servers
REM ============================================

echo.
echo ====================================
echo   Stopping TaskFlow Services
echo ====================================
echo.

echo Stopping Backend (Go on port 8080)...
for /f "tokens=5" %%a in ('netstat -aon ^| find ":8080" ^| find "LISTENING"') do taskkill /F /PID %%a 2>nul

echo Stopping Frontend (Next.js on port 3000)...
for /f "tokens=5" %%a in ('netstat -aon ^| find ":3000" ^| find "LISTENING"') do taskkill /F /PID %%a 2>nul

echo.
echo All services stopped!
echo.
