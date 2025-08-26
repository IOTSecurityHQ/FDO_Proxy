#!/bin/bash

echo "=== FDO Server Proxy with Passport Service Demo ==="
echo "This demo shows the proxy intercepting FDO protocol messages and calling passport services"
echo ""

# Kill any existing processes
pkill -f fdo-proxy
pkill -f mock-fdo-backend
pkill -f mock-passport-service

echo "1. Starting mock passport service..."
./mock-passport-service &
PASSPORT_PID=$!
sleep 3

echo "2. Testing passport service endpoints..."
echo "   Testing product item passport endpoint:"
curl -s "http://localhost:8443/product_item/?uuid=test-device-123" | jq . 2>/dev/null || curl -s "http://localhost:8443/product_item/?uuid=test-device-123"
echo ""
echo "   Testing commissioning passport endpoint:"
curl -s -X POST http://localhost:8000/create-commissioning-passport \
  -H "Content-Type: application/json" \
  -d '{"controller_uuid":"test-device-123","cert":"","deployed_location":"","timestamp":"1234567890"}' | jq . 2>/dev/null || curl -s -X POST http://localhost:8000/create-commissioning-passport \
  -H "Content-Type: application/json" \
  -d '{"controller_uuid":"test-device-123","cert":"","deployed_location":"","timestamp":"1234567890"}'
echo ""
echo ""

echo "3. Starting mock FDO backend server..."
./mock-fdo-backend &
BACKEND_PID=$!
sleep 2

echo "4. Starting FDO proxy server with passport service integration..."
./fdo-proxy -debug \
  -listen localhost:8080 \
  -product-base-url "http://localhost:8443" \
  -commissioning-url "http://localhost:8000/create-commissioning-passport" \
  -enable-product-passport &
PROXY_PID=$!
sleep 3

echo "5. Testing proxy with FDO DI.AppStart request (should trigger product passport lookup)..."
echo "   Request: POST http://localhost:8080/di with DI.AppStart message"
curl -s -X POST http://localhost:8080/di \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "6. Testing proxy with FDO TO2.Done2 response (should trigger commissioning passport creation)..."
echo "   Request: POST http://localhost:8080/to2 with TO2.Done2 response"
curl -s -X POST http://localhost:8080/to2 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 response content" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "7. Testing proxy health endpoint..."
curl -s http://localhost:8080/health
echo ""
echo ""

echo "8. Demo complete. Cleaning up..."
kill $PROXY_PID 2>/dev/null
kill $BACKEND_PID 2>/dev/null
kill $PASSPORT_PID 2>/dev/null
echo "Demo finished!"
echo ""
echo "Key Features Demonstrated:"
echo "- Proxy intercepts FDO protocol messages"
echo "- Product item passport lookup during DI.AppStart"
echo "- Commissioning passport creation during TO2.Done2"
echo "- Graceful error handling when passport services are unavailable"
echo "- Clean startup and shutdown" 