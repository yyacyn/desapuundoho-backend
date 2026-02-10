#!/bin/bash

# Kill all running server instances
echo "Stopping all server instances..."
pkill -9 server

# Wait a moment
sleep 2

# Check if any are still running
if pgrep server > /dev/null; then
    echo "⚠️  Some server processes are still running:"
    ps aux | grep server
else
    echo "✅ All server processes stopped"
fi

# Show which ports are in use
echo ""
echo "Ports currently in use:"
netstat -tulpn 2>/dev/null | grep LISTEN || echo "netstat not available"
