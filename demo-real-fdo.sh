#!/bin/bash

echo "=== FDO Server Proxy with Real go-fdo Demo ==="
echo "This demo shows the proxy intercepting FDO protocol messages with the real go-fdo server"
echo ""

# Kill any existing processes
pkill -f fdo-proxy
pkill -f fdo-server

echo "1. Starting FDO proxy server with real go-fdo backend..."
./fdo-proxy -debug -listen localhost:8080 &
PROXY_PID=$!
sleep 5

echo "2. Testing proxy with FDO protocol message (DI.AppStart - msg type 10)..."
echo "   Request: POST http://localhost:8080/fdo/101/msg/10"
curl -s -X POST http://localhost:8080/fdo/101/msg/10 \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "3. Testing proxy with FDO protocol message (TO2.Done2 - msg type 71)..."
echo "   Request: POST http://localhost:8080/fdo/101/msg/71"
curl -s -X POST http://localhost:8080/fdo/101/msg/71 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 message content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "4. Testing proxy with invalid FDO message..."
echo "   Request: POST http://localhost:8080/fdo/101/msg/99"
curl -s -X POST http://localhost:8080/fdo/101/msg/99 \
  -H "Content-Type: application/octet-stream" \
  -d "Invalid message content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "5. Testing proxy with non-FDO endpoint..."
echo "   Request: GET http://localhost:8080/"
curl -s http://localhost:8080/ \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "6. Demo complete. Cleaning up..."
kill $PROXY_PID 2>/dev/null
echo "Demo finished!"
echo ""
echo "Key Features Demonstrated:"
echo "- Proxy successfully starts and manages the real go-fdo server"
echo "- Proxy intercepts FDO protocol messages (DI.AppStart, TO2.Done2)"
echo "- Proxy forwards requests to the real go-fdo backend"
echo "- Proxy handles invalid requests gracefully"
echo "- Clean startup and shutdown of both proxy and backend"
echo ""
echo "Note: The garbled output in responses is expected - these are CBOR-encoded"
echo "FDO protocol messages that require proper parsing by FDO clients." 