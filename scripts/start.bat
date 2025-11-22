@echo off
REM ============================================
REM TaskFlow - Start Local Development
REM Backend: Go + Supabase
REM Frontend: Next.js
REM ============================================

REM Change to project root directory
cd /d "%~dp0\.."

echo.
echo ====================================
echo   TaskFlow - Starting Local Dev
echo ====================================
echo.

REM Check if backend .env exists
if not exist "backend\.env" (
    echo ERROR: backend\.env not found!
    echo Please copy backend\.env.example to backend\.env and configure your Supabase connection.
    exit /b 1
)

echo [1/3] Starting Backend (Go + Supabase)...
start "TaskFlow Backend" cmd /k "cd /d %CD%\backend && go run cmd/server/main.go"

echo Waiting for backend to start...
timeout /t 3 /nobreak >nul

echo.
echo [2/3] Starting Frontend (Next.js)...
start "TaskFlow Frontend" cmd /k "cd /d %CD%\frontend && npm run dev"

echo Waiting for frontend to start...
timeout /t 3 /nobreak >nul

echo.
echo [3/3] All services started!
echo.
echo ====================================
echo   TaskFlow is ready!
echo ====================================
echo.
echo   Frontend:  http://localhost:3000
echo   Backend:   http://localhost:8080
echo.
echo   Backend and Frontend are running in separate windows.
echo   Close those windows to stop the services.
echo.
