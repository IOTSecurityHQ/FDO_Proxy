# FDO Server Proxy Demo Guide

## Overview

This demo showcases the FDO Server Proxy, a thin middleware layer that intercepts FDO protocol messages and integrates with external passport services. The proxy acts as a reverse proxy for the go-fdo backend server, adding passport service functionality without modifying the core FDO code.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   FDO Client    │    │  FDO Proxy       │    │  Passport       │
│                 │    │  (Interceptor)   │    │  Service        │
│                 │◄──►│                  │◄──►│                 │
│                 │    │  • Request       │    │    Product      │
│                 │    │    Interception  │    │    Item         │
│                 │    │  • Response      │    │    Passports    │
│                 │    │    Interception  │    │  • Commissioning│
│                 │    │                  │    │    Passports    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Demo Components

1. **FDO Proxy** (`./fdo-proxy`) - The main proxy server
2. **Real go-fdo Server** (`../go-fdo/fdo-server`) - The actual FDO server backend
3. **Mock Passport Service** (`./mock-passport-service`) - Simulates external passport APIs

## Quick Demo

### Option 1: Real go-fdo Demo (Recommended)
```bash
./demo-real-fdo.sh
```

This demonstrates:
- Proxy startup and configuration with real go-fdo server
- Request forwarding to real FDO backend
- FDO protocol message handling (DI.AppStart, TO2.Done2)
- Proper CBOR-encoded FDO protocol responses

### Option 2: Mock Demo (For Testing)
```bash
./demo.sh
```

This demonstrates:
- Proxy startup and configuration with mock backend
- Request forwarding to mock backend
- Basic FDO protocol message handling

### Option 3: Full Demo (With Passport Services)
```bash
./demo-with-passport.sh
```

This demonstrates:
- Product item passport lookup during DI.AppStart
- Commissioning passport creation during TO2.Done2
- Integration with external passport services
- Error handling and graceful degradation

## Manual Demo Steps

### 1. Build Real go-fdo Server (One-time setup)
```bash
# From the fdo-server-wrapper directory
cd ../go-fdo/examples/cmd
go mod tidy
go build -o fdo-server .
cp fdo-server ../../../go-fdo/
cd ../../../fdo-server-wrapper
```

### 2. Start FDO Proxy with Real go-fdo Backend
```bash
./fdo-proxy -debug -listen localhost:8080
```

The proxy will automatically:
- Start the real go-fdo server on port 8081
- Forward requests from port 8080 to the go-fdo server
- Intercept FDO protocol messages for middleware processing

### 3. Test FDO Protocol Messages
```bash
# DI.AppStart message (Device Initialization)
curl -X POST http://localhost:8080/fdo/101/msg/10 \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message content"

# TO2.Done2 message (Transfer Ownership 2)
curl -X POST http://localhost:8080/fdo/101/msg/71 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 message content"
```

### 4. Optional: Add Passport Service Integration
```bash
# Start mock passport service
./mock-passport-service &

# Start proxy with passport integration
./fdo-proxy -debug \
  -listen localhost:8080 \
  -product-base-url "http://localhost:8443" \
  -commissioning-url "http://localhost:8000/create-commissioning-passport" \
  -enable-product-passport
```

### 4. Test FDO Protocol Messages

#### DI.AppStart Request (Product Passport Lookup)
```bash
curl -X POST http://localhost:8080/di \
  -H "Content-Type: application/octet-stream" \
  -d "DI.AppStart message content"
```

#### TO2.Done2 Response (Commissioning Passport Creation)
```bash
curl -X POST http://localhost:8080/to2 \
  -H "Content-Type: application/octet-stream" \
  -d "TO2.Done2 response content"
```

## Expected Demo Output

### Real go-fdo Demo Output
```
=== FDO Server Proxy with Real go-fdo Demo ===
1. Starting FDO proxy server with real go-fdo backend...
   [Proxy starts go-fdo server on port 8081]
2. Testing proxy with FDO protocol message (DI.AppStart - msg type 10)...
   [Real go-fdo processes DI.AppStart message]
   [CBOR-encoded response with proper FDO protocol]
3. Testing proxy with FDO protocol message (TO2.Done2 - msg type 71)...
   [Real go-fdo processes TO2.Done2 message]
   [CBOR-encoded response with proper FDO protocol]
4. Testing proxy with invalid FDO message...
   [Real go-fdo returns proper error response]
5. Testing proxy with non-FDO endpoint...
   404 page not found
```

### Mock Demo Output
```
=== FDO Server Proxy Demo ===
1. Starting mock FDO backend server...
2. Starting FDO proxy server...
3. Testing proxy with regular HTTP request...
   Mock FDO Backend Server
   Request: GET /
4. Testing proxy with FDO DI.AppStart request...
   Mock FDO Backend Server
   Request: POST /di
5. Testing proxy with FDO TO2.Done2 response...
   Mock FDO Backend Server
   Request: POST /to2
6. Testing proxy health endpoint...
   OK
```

### Full Demo Output (With Passport Services)
```
=== FDO Server Proxy with Passport Service Demo ===
1. Starting mock passport service...
2. Testing passport service endpoints...
   {"schema_version":"1.0","uuid":"test-device-123",...}
   {"status":"success","message":"Commissioning passport created",...}
3. Starting mock FDO backend server...
4. Starting FDO proxy server with passport service integration...
5. Testing proxy with FDO DI.AppStart request...
   [Proxy logs show passport lookup attempt]
6. Testing proxy with FDO TO2.Done2 response...
   [Proxy logs show commissioning passport creation]
```

## Key Features Demonstrated

1. **Request Interception**: Proxy intercepts FDO protocol messages
2. **Passport Integration**: Calls external passport services during FDO flow
3. **Graceful Degradation**: Continues working even if passport services fail
4. **Clean Architecture**: Thin layer that doesn't modify core FDO code
5. **Error Handling**: Proper logging and error management
6. **Health Monitoring**: Health endpoint for monitoring

## Production Setup

For production deployment:

1. **Real go-fdo Backend**: Replace mock backend with actual go-fdo server
2. **Real Passport Services**: Configure actual passport service URLs
3. **mTLS Certificates**: Add proper certificates for product passport service
4. **Docker Deployment**: Use provided Docker Compose setup
5. **Monitoring**: Add proper logging and monitoring

## Troubleshooting

- **Port conflicts**: Ensure ports 8080, 8081, 8000, 8443 are available
- **Certificate errors**: For production, provide valid mTLS certificates
- **Backend not found**: Ensure go-fdo is properly built or use Docker
- **Passport service errors**: Check network connectivity and service URLs

## Demo Scripts

- `demo.sh` - Basic functionality demo
- `demo-with-passport.sh` - Full integration demo
- `test-backend.go` - Mock FDO backend server
- `mock-passport.go` - Mock passport service

## Next Steps

1. **Real Integration**: Connect to actual passport services
2. **Production Deployment**: Use Docker Compose for containerized deployment
3. **Monitoring**: Add metrics and monitoring
4. **Testing**: Run comprehensive test suite
5. **Documentation**: Update documentation for production use 