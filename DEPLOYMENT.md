# Deployment Instructions for Backend

## For Koyeb (or other cloud platforms):

### Option 1: Use Environment Variables (Recommended)
Set these environment variables in Koyeb dashboard:
```
PORT=8080
DB_HOST=your_db_host
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
```

The app will work WITHOUT a .env file!

### Option 2: Deploy without Database
Just set:
```
PORT=8080
```

The app will run without database connection.

## For cPanel:

### 1. Kill existing server instances:
```bash
# Upload kill-server.sh to ~/go/
chmod +x ~/go/kill-server.sh
./kill-server.sh
```

### 2. Set environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=your_cpanel_user
export DB_PASSWORD="your_password"
export DB_NAME=your_database
export PORT=8081
```

### 3. Start server:
```bash
cd ~/go
chmod +x server
nohup ./server > server.log 2>&1 &
```

### 4. Check if running:
```bash
ps aux | grep server
cat ~/go/server.log
```

## Troubleshooting:

### Port already in use:
```bash
# Find what's using the port
lsof -i :8081

# Kill it
pkill -9 server

# Or kill specific PID
kill -9 <PID>
```

### Check logs:
```bash
tail -f ~/go/server.log
```
