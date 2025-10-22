#!/bin/bash

BASE_URL="http://localhost:8887/api/v1/users"

echo "=== GET all users ==="
curl -s -X GET "$BASE_URL" -H "Accept: application/json"
echo -e "\n"

echo "=== GET users with role=admin ==="
curl -s -X GET "$BASE_URL?role=admin" -H "Accept: application/json"
echo -e "\n"

echo "=== POST new user ==="
resp=$(curl -s -X POST "$BASE_URL" \
  -H "Content-Type: application/json" \
  -d '{"username":"TestUser","role":"user"}')
echo "$resp"
id=$(echo "$resp" | jq '.id')
echo "New user ID: $id"
echo -e "\n"

echo "=== GET user by ID ==="
curl -s -X GET "$BASE_URL/$id" -H "Accept: application/json"
echo -e "\n"

echo "=== PUT update user ==="
curl -s -X PUT "$BASE_URL/$id" \
  -H "Content-Type: application/json" \
  -d '{"username":"UpdatedUser","role":"user"}'
echo -e "\n"

echo "=== GET updated user ==="
curl -s -X GET "$BASE_URL/$id" -H "Accept: application/json"
echo -e "\n"

echo "=== DELETE user ==="
curl -s -X DELETE "$BASE_URL/$id"
echo -e "\n"

echo "=== GET deleted user (should 404) ==="
curl -s -X GET "$BASE_URL/$id" -H "Accept: application/json"
echo -e "\n"

echo "=== GET all users after deletion ==="
curl -s -X GET "$BASE_URL" -H "Accept: application/json"
echo -e "\n"

echo "=== Test complete ==="
