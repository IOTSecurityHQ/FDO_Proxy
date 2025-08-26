#!/bin/bash

echo "=== FDO Client-Server Demo with Passport Service Integration ==="
echo "This demo shows FDO client connecting through proxy with passport middleware"
echo ""

# Kill any existing processes
pkill -f fdo-proxy
pkill -f fdo-server
pkill -f mock-passport-service

echo "1. Starting mock passport service..."
./mock-passport-service &
PASSPORT_PID=$!
sleep 2

echo "2. Starting FDO proxy with passport service integration..."
./fdo-proxy -debug \
  -listen localhost:8080 \
  -product-base-url "http://localhost:8443" \
  -commissioning-url "http://localhost:8000/create-commissioning-passport" \
  -enable-product-passport &
PROXY_PID=$!
sleep 3

echo "3. Testing FDO client DI.AppStart with passport lookup..."
echo "   Proxy will intercept DI.AppStart and call passport service"
curl -s -X POST http://localhost:8080/fdo/101/msg/10 \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message from FDO client" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "4. Testing FDO client TO2.Done2 with commissioning passport creation..."
echo "   Proxy will intercept TO2.Done2 and create commissioning passport"
curl -s -X POST http://localhost:8080/fdo/101/msg/71 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 message from FDO client" \
  -w "\nStatus: %{http_code}\n"
echo ""

echo "5. Testing passport service endpoints directly..."
echo "   Product item passport endpoint:"
curl -s "http://localhost:8443/product_item/?uuid=fdo-device-123" | head -c 200
echo "..."
echo ""

echo "6. Demo complete. Cleaning up..."
kill $PROXY_PID 2>/dev/null
kill $PASSPORT_PID 2>/dev/null
echo "Demo finished!"
echo ""
echo "Key Points Demonstrated:"
echo "- FDO client connects to proxy (port 8080)"
echo "- Proxy intercepts FDO protocol messages"
echo "- Proxy calls passport service during DI.AppStart"
echo "- Proxy creates commissioning passport during TO2.Done2"
echo "- Middleware injection works without modifying go-fdo server"
echo "- Clean separation of concerns: proxy handles passport integration" 