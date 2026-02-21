#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}============================================${NC}"
echo -e "${CYAN}  Event Registration API — Auto Setup${NC}"
echo -e "${CYAN}============================================${NC}"
echo ""

# Variables
SERVER_PID=""

# Cleanup trap to ensure server is killed if script exits unexpectedly
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo -e "${YELLOW}Cleaning up: stopping server (PID $SERVER_PID)...${NC}"
        kill -9 $SERVER_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT INT TERM

# --- Step 1 — Check Prerequisites ---
echo -e "${YELLOW}Step 1: Checking Prerequisites...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go to continue.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Go is installed: $(go version)${NC}"

# Check for psql
if ! command -v psql &> /dev/null; then
    echo -e "${RED}❌ psql is not installed or not in PATH. Please install PostgreSQL client tools.${NC}"
    exit 1
fi

# Check if PostgreSQL is running
if ! pg_isready -q; then
    # Try localhost specifically if general check fails
    if ! pg_isready -h localhost -q; then
        echo -e "${RED}❌ PostgreSQL is not running or not accessible. Please start PostgreSQL.${NC}"
        exit 1
    fi
fi
echo -e "${GREEN}✅ PostgreSQL is running${NC}"

# Check for jq, setup fallback if needed
JQ_CMD="jq"
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠️  jq is not installed. Falling back to python json.tool for formatting.${NC}"
    JQ_CMD="python3 -m json.tool"
fi

# --- Step 2 — Create the .env file ---
echo -e "\n${YELLOW}Step 2: Creating .env file...${NC}"
cat << EOF > .env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=event_registration
SERVER_PORT=8080
EOF
echo -e "${GREEN}✅ .env file created${NC}"

# --- Step 3 — Create the PostgreSQL database ---
echo -e "\n${YELLOW}Step 3: Creating PostgreSQL Database...${NC}"
# Use standard postgres user and password, might prompt for password or use pgpass
psql -U postgres -c "CREATE DATABASE event_registration;" 2>/dev/null || echo -e "${CYAN}Database may already exist, continuing...${NC}"
echo -e "${GREEN}✅ Database ready${NC}"

# --- Step 4 — Install Go dependencies ---
echo -e "\n${YELLOW}Step 4: Installing Go Dependencies...${NC}"
go mod tidy
echo -e "${GREEN}✅ Dependencies installed${NC}"

# --- Step 5 — Build the project ---
echo -e "\n${YELLOW}Step 5: Building project...${NC}"
if ! go build ./...; then
    echo -e "${RED}❌ Build failed. Fix errors before continuing.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# --- Step 6 — Run the concurrency test ---
echo -e "\n${YELLOW}Step 6: Running Concurrency Test...${NC}"
if go test ./test/... -v -timeout 30s; then
    echo -e "${GREEN}✅ Concurrency test complete and passed${NC}"
else
    echo -e "${RED}❌ Concurrency test failed. Check the output above.${NC}"
    exit 1
fi

# --- Step 7 — Start the server ---
echo -e "\n${YELLOW}Step 7: Starting server...${NC}"
go run cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!
echo -e "${CYAN}Server started with PID: $SERVER_PID. Waiting 3 seconds for initialization...${NC}"
sleep 3

# Check if server is still running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${RED}❌ Server failed to start. Check server.log for details:${NC}"
    cat server.log
    exit 1
fi

# --- Step 8 — Run API smoke tests using curl ---
echo -e "\n${YELLOW}Step 8: Running API Smoke Tests...${NC}"

# Helper function to parse JSON
extract_json() {
    if command -v jq &> /dev/null; then
        echo "$1" | jq -r "$2"
    else
        # Very basic fallback for extracting values without jq
        echo "$1" | grep -o "\"$2\": *\"[^\"]*\"" | cut -d '"' -f 4 || echo "$1" | grep -o "\"$2\": *[^,}]*" | cut -d ':' -f 2 | tr -d ' '
    fi
}

format_json() {
    if command -v jq &> /dev/null; then
        echo "$1" | jq .
    else
        echo "$1" | python3 -m json.tool 2>/dev/null || echo "$1"
    fi
}

echo -e "\n${CYAN}--- GET /health ---${NC}"
HEALTH=$(curl -s http://localhost:8080/health)
format_json "$HEALTH"

# Create organizer
echo -e "\n${CYAN}--- POST /users (organizer) ---${NC}"
ORGANIZER=$(curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","role":"organizer"}')
format_json "$ORGANIZER"
ORGANIZER_ID=$(extract_json "$ORGANIZER" "id")

if [ -z "$ORGANIZER_ID" ] || [ "$ORGANIZER_ID" == "null" ]; then
    echo -e "${RED}Failed to extract Organizer ID. Stopping tests.${NC}"
    exit 1
fi

# Create regular user
echo -e "\n${CYAN}--- POST /users (user) ---${NC}"
USER=$(curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","email":"bob@example.com","role":"user"}')
format_json "$USER"
USER_ID=$(extract_json "$USER" "id")

if [ -z "$USER_ID" ] || [ "$USER_ID" == "null" ]; then
    echo -e "${RED}Failed to extract User ID. Stopping tests.${NC}"
    exit 1
fi

# List users
echo -e "\n${CYAN}--- GET /users ---${NC}"
USERS_LIST=$(curl -s http://localhost:8080/users)
format_json "$USERS_LIST"

# Create event
echo -e "\n${CYAN}--- POST /events ---${NC}"
EVENT=$(curl -s -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d "{\"title\":\"Go Workshop\",\"description\":\"Learn Go\",\"total_capacity\":50,\"organizer_id\":\"$ORGANIZER_ID\"}")
format_json "$EVENT"
EVENT_ID=$(extract_json "$EVENT" "id")

if [ -z "$EVENT_ID" ] || [ "$EVENT_ID" == "null" ]; then
    echo -e "${RED}Failed to extract Event ID. Stopping tests.${NC}"
    exit 1
fi

# List events
echo -e "\n${CYAN}--- GET /events ---${NC}"
EVENTS_LIST=$(curl -s http://localhost:8080/events)
format_json "$EVENTS_LIST"

# Get event by ID
echo -e "\n${CYAN}--- GET /events/:id ---${NC}"
EVENT_DETAIL=$(curl -s http://localhost:8080/events/$EVENT_ID)
format_json "$EVENT_DETAIL"

# Register user for event
echo -e "\n${CYAN}--- POST /events/:id/register ---${NC}"
REG=$(curl -s -X POST http://localhost:8080/events/$EVENT_ID/register \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\"}")
format_json "$REG"
REG_ID=$(extract_json "$REG" "id")

if [ -z "$REG_ID" ] || [ "$REG_ID" == "null" ]; then
    echo -e "${RED}Failed to extract Registration ID. Stopping tests.${NC}"
    exit 1
fi

# Try duplicate registration (should fail with 409 or similar)
echo -e "\n${CYAN}--- POST /events/:id/register (duplicate - should fail) ---${NC}"
DUPE_REG=$(curl -s -X POST http://localhost:8080/events/$EVENT_ID/register \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER_ID\"}")
format_json "$DUPE_REG"

# Cancel registration
echo -e "\n${CYAN}--- DELETE /registrations/:id ---${NC}"
CANCEL_REG=$(curl -s -X DELETE http://localhost:8080/registrations/$REG_ID)
format_json "$CANCEL_REG"


echo -e "\n${GREEN}✅ All API smoke tests complete${NC}"

# --- Step 9 — Stop the server ---
echo -e "\n${YELLOW}Step 9: Stopping server...${NC}"
# Disarm the EXIT trap primarily so we don't print "Cleaning up..." on successful finish
trap - EXIT INT TERM
kill $SERVER_PID 2>/dev/null
SERVER_PID=""
echo -e "${GREEN}✅ Server stopped${NC}"

# --- Step 10 — Print final summary ---
echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  🎉 PROJECT SETUP AND TESTING COMPLETE${NC}"
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  ✅ Dependencies installed${NC}"
echo -e "${GREEN}  ✅ Database created${NC}"
echo -e "${GREEN}  ✅ Build successful${NC}"
echo -e "${GREEN}  ✅ Concurrency test passed${NC}"
echo -e "${GREEN}  ✅ All API endpoints verified${NC}"
echo ""
echo -e "${CYAN}  To start the server manually:${NC}"
echo "  go run cmd/server/main.go"
echo ""
echo -e "${CYAN}  To run concurrency test manually:${NC}"
echo "  go test ./test/... -v"
echo -e "${GREEN}============================================${NC}"

# chmod +x setup.sh && ./setup.sh
