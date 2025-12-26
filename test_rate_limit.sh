#!/bin/bash

echo "Testing Rate Limiter Script to Test (10 requests limit, 1 token/sec refill)"
echo "============================================================"
echo ""

# Start server in background
echo "Starting server..."
go run cmd/main.go > /dev/null 2>&1 &
SERVER_PID=$!
sleep 2

echo "Making 15 rapid requests (should allow first 10, block next 5):"
echo ""

for i in {1..15}; do
    response=$(curl -s http://localhost:8080/v1/resource)
    if echo "$response" | grep -q "Too many requests"; then
        echo "Request $i: RATE LIMITED"
    else
        echo "Request $i: ALLOWED"
    fi
done

echo ""
echo "Waiting 2 seconds for token refill..."
sleep 2

echo ""
echo "Making 3 more requests after refill (should allow 2 more):"
for i in {16..18}; do
    response=$(curl -s http://localhost:8080/v1/resource)
    if echo "$response" | grep -q "Too many requests"; then
        echo "Request $i: RATE LIMITED"
    else
        echo "Request $i: ALLOWED"
    fi
done

# Cleanup
kill $SERVER_PID 2>/dev/null
echo ""
echo "Test complete!"

