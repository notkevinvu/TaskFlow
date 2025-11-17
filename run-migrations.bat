@echo off
echo Running database migrations...
echo.

REM Wait for PostgreSQL to be ready
timeout /t 5 /nobreak

REM Run migrations using golang-migrate in Docker
docker exec taskflow-backend sh -c "apk add --no-cache curl && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xz && mv migrate /usr/local/bin/ && migrate -path /root/migrations -database 'postgres://taskflow_user:taskflow_dev_password@postgres:5432/taskflow?sslmode=disable' up"

echo.
echo Migrations complete!
pause
