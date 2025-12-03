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

REM Check and kill processes on port 8080 (Backend)
echo [0/3] Checking ports...

REM Kill backend port 8080
netstat -ano | findstr ":8080" | findstr "LISTENING" > nul 2>&1
if %errorlevel% equ 0 (
    echo Found process on port 8080, killing...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING"') do (
        taskkill /F /PID %%a >nul 2>&1
    )
)

REM Kill frontend port 3000
netstat -ano | findstr ":3000" | findstr "LISTENING" > nul 2>&1
if %errorlevel% equ 0 (
    echo Found process on port 3000, killing...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000" ^| findstr "LISTENING"') do (
        taskkill /F /PID %%a >nul 2>&1
    )
)

echo Ports 8080 and 3000 are now available.
echo.

echo [1/3] Starting Backend (Go + Supabase)...
start "TaskFlow Backend" cmd /k "cd /d %CD%\backend && go run cmd/server/main.go"

echo Waiting for backend to start...
timeout /t 3 /nobreak >nul

echo.
echo [2/3] Starting Frontend (Next.js)...
start "TaskFlow Frontend" cmd /k "cd /d %CD%\frontend && npm run dev"

echo Waiting for frontend to be ready...
:wait_for_frontend
timeout /t 2 /nobreak >nul
curl -s -o nul -w "" http://localhost:3000 >nul 2>&1
if %errorlevel% neq 0 (
    echo   Still waiting for frontend...
    goto wait_for_frontend
)

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

REM Open frontend in default browser
echo Opening http://localhost:3000 in your browser...
start http://localhost:3000
