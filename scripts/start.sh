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

# Check and kill processes on port 8080 (Backend)
echo "[0/3] Checking ports..."
if command -v lsof &> /dev/null; then
    BACKEND_PID_ON_PORT=$(lsof -ti:8080 2>/dev/null)
    if [ ! -z "$BACKEND_PID_ON_PORT" ]; then
        echo "Killing process on port 8080 (PID: $BACKEND_PID_ON_PORT)..."
        kill -9 $BACKEND_PID_ON_PORT 2>/dev/null
    fi

    # Check and kill processes on port 3000 (Frontend)
    FRONTEND_PID_ON_PORT=$(lsof -ti:3000 2>/dev/null)
    if [ ! -z "$FRONTEND_PID_ON_PORT" ]; then
        echo "Killing process on port 3000 (PID: $FRONTEND_PID_ON_PORT)..."
        kill -9 $FRONTEND_PID_ON_PORT 2>/dev/null
    fi
fi

echo "Ports 8080 and 3000 are now available."
echo ""

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

echo "Waiting for frontend to be ready..."
while ! curl -s -o /dev/null -w "" http://localhost:3000 2>/dev/null; do
    echo "  Still waiting for frontend..."
    sleep 2
done

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

# Open frontend in default browser
echo "Opening http://localhost:3000 in your browser..."
if command -v open &> /dev/null; then
    # macOS
    open http://localhost:3000
elif command -v xdg-open &> /dev/null; then
    # Linux
    xdg-open http://localhost:3000
fi

# Wait for background processes
wait
