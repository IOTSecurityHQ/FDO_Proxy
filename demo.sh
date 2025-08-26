#!/bin/bash

echo "=== FDO Server Proxy Demo ==="
echo "This demo shows the proxy intercepting FDO protocol messages"
echo ""

# Kill any existing processes
pkill -f fdo-proxy
pkill -f mock-fdo-backend

echo "1. Starting mock FDO backend server..."
./mock-fdo-backend &
BACKEND_PID=$!
sleep 2

echo "2. Starting FDO proxy server..."
./fdo-proxy -debug -listen localhost:8080 &
PROXY_PID=$!
sleep 3

echo "3. Testing proxy with regular HTTP request..."
echo "   Request: GET http://localhost:8080/"
curl -s http://localhost:8080/
echo ""
echo ""

echo "4. Testing proxy with FDO DI.AppStart request (should trigger product passport lookup)..."
echo "   Request: POST http://localhost:8080/di with DI.AppStart message"
curl -s -X POST http://localhost:8080/di \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "5. Testing proxy with FDO TO2.Done2 response (should trigger commissioning passport creation)..."
echo "   Request: POST http://localhost:8080/to2 with TO2.Done2 response"
curl -s -X POST http://localhost:8080/to2 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 response content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "6. Testing proxy health endpoint..."
curl -s http://localhost:8080/health
echo ""
echo ""

echo "7. Demo complete. Cleaning up..."
kill $PROXY_PID
kill $BACKEND_PID
echo "Demo finished!" 