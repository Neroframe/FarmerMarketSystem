# FarmerMarketSystem (FMS)

## Setup    
I am running my DB inside Windows, while my go server is in Windows Subsystem for Linux (WSL). This is why your setup might slightly differ from mine.
### 1. Configure Database Connection
Set the following environment variables to configure the database connection:
```bash
export DB_HOST=<database_host_ip>
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=<database_password>
export DB_NAME=fms
```

To find your `<database_host_ip>` when using WSL, run:
```bash
WIN_IP=$(cat /etc/resolv.conf | grep nameserver | awk '{print $2}')
echo $WIN_IP
```

Then make similar adjustments for DB initializer in backend/cmd/main.go

### 2. Test Database Connection

To ensure the database connection is correctly set up in WSL, run:
```bash
psql -h <DB_HOST> -U <DB_USER> -d <DB_NAME>
```

Replace `<DB_HOST>`, `<DB_USER>`, and `<DB_NAME>` with the actual values from your setup.

### 3. Build and Run the Project
Once the database connection is confirmed, build and run the project with:
```bash
go build -o fms-backend ./backend/cmd
./fms-backend
```

Your project should now be accessible at http://localhost:8080/register.