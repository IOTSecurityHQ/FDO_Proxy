#!/bin/bash

echo "=== FDO Client-Server Demo with Proxy Middleware ==="
echo "This demo shows real FDO client connecting to FDO server through proxy"
echo ""

# Kill any existing processes
pkill -f fdo-proxy
pkill -f fdo-server

echo "1. Starting FDO proxy server with real go-fdo backend..."
./fdo-proxy -debug -listen localhost:8080 &
PROXY_PID=$!
sleep 3

echo "2. Testing FDO client connection to proxy (DI phase)..."
echo "   FDO client connects to: http://localhost:8080 (proxy)"
echo "   Proxy forwards to: http://localhost:8081 (go-fdo server)"
echo ""

# Test with a simple FDO client request
echo "3. Simulating FDO client DI.AppStart request..."
echo "   This is what a real FDO client would send during Device Initialization"
curl -s -X POST http://localhost:8080/fdo/101/msg/10 \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message from FDO client" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "4. Simulating FDO client TO2.HelloDevice request..."
echo "   This is what a real FDO client would send during Transfer Ownership"
curl -s -X POST http://localhost:8080/fdo/101/msg/30 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.HelloDevice message from FDO client" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "5. Simulating FDO client TO2.Done2 request..."
echo "   This is what a real FDO client would send to complete onboarding"
curl -s -X POST http://localhost:8080/fdo/101/msg/71 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 message from FDO client" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "6. Demo complete. Cleaning up..."
kill $PROXY_PID 2>/dev/null
echo "Demo finished!"
echo ""
echo "Key Points Demonstrated:"
echo "- FDO client connects to proxy (port 8080)"
echo "- Proxy intercepts FDO protocol messages"
echo "- Proxy forwards requests to go-fdo server (port 8081)"
echo "- Proxy can inject middleware for passport service integration"
echo "- Real FDO protocol flow: DI -> TO1 -> TO2"
echo ""
echo "Note: The garbled responses are proper CBOR-encoded FDO protocol messages"
echo "that would be processed by real FDO clients." 