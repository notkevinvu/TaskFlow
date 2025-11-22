#!/bin/bash
# ============================================
# TaskFlow - Start Local Development
# Backend: Go + Supabase
# Frontend: Next.js
# ============================================

# Change to project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo ""
echo "===================================="
echo "   TaskFlow - Starting Local Dev"
echo "===================================="
echo ""

# Check if backend .env exists
if [ ! -f "backend/.env" ]; then
    echo "ERROR: backend/.env not found!"
    echo "Please copy backend/.env.example to backend/.env and configure your Supabase connection."
    exit 1
fi

# Cleanup function
cleanup() {
    echo ""
    echo "Stopping services..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "[1/3] Starting Backend (Go + Supabase)..."
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

echo "Waiting for backend to start..."
sleep 3

echo ""
echo "[2/3] Starting Frontend (Next.js)..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo "Waiting for frontend to start..."
sleep 3

echo ""
echo "[3/3] All services started!"
echo ""
echo "===================================="
echo "   TaskFlow is ready!"
echo "===================================="
echo ""
echo "   Frontend:  http://localhost:3000"
echo "   Backend:   http://localhost:8080"
echo ""
echo "   Press Ctrl+C to stop all services"
echo ""

# Wait for background processes
wait
