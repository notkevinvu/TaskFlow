@echo off
echo Starting TaskFlow Frontend...
echo.

cd frontend

REM Check if .env exists, if not create it
if not exist .env (
    echo Creating .env file...
    echo NEXT_PUBLIC_API_URL=http://localhost:8080 > .env
    echo NODE_ENV=production >> .env
    echo.
)

echo Starting development server...
npm run dev
