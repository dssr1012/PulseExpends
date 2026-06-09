#!/bin/bash
cd /root/PulseExpends/services/backend/mcp-server-integrated
export PATH=$PATH:/usr/local/go/bin
export PORT=3001
export DATABASE_URL="postgresql://pulseexpends:YOUR_PASSWORD@10.0.101.182:5432/pulseexpends_auth?sslmode=disable"
./mcp-server-integrated