@echo off
echo Testing TaskFlow Backend...
echo.

echo 1. Health Check...
curl http://localhost:8080/health
echo.
echo.

echo 2. Registering test user...
curl -X POST http://localhost:8080/api/v1/auth/register -H "Content-Type: application/json" -d "{\"email\":\"test@taskflow.dev\",\"name\":\"Test User\",\"password\":\"Test1234\"}"
echo.
echo.

echo 3. Logging in...
curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d "{\"email\":\"test@taskflow.dev\",\"password\":\"Test1234\"}"
echo.
echo.

echo.
echo Copy the access_token from above to use in further tests!
echo.
pause
