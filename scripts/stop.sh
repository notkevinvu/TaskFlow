#!/bin/bash
# ============================================
# TaskFlow - Stop Development Servers
# ============================================

echo ""
echo "===================================="
echo "   Stopping TaskFlow Services"
echo "===================================="
echo ""

echo "Stopping Backend (Go on port 8080)..."
lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "Backend not running"

echo "Stopping Frontend (Next.js on port 3000)..."
lsof -ti:3000 | xargs kill -9 2>/dev/null || echo "Frontend not running"

echo ""
echo "All services stopped!"
echo ""
