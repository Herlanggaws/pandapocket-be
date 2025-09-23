# Currency Setup Scripts

This directory contains scripts to add default currencies to the PandaPocket database.

## Available Scripts

### 1. `run_add_currencies.sh` (Recommended)
A comprehensive shell script that provides multiple options for adding currencies.

**Usage:**
```bash
# From the backend directory
./scripts/run_add_currencies.sh [go|sql|auto]
```

**Options:**
- `go` - Use Go script (recommended)
- `sql` - Use SQL script directly  
- `auto` - Try Go first, fallback to SQL (default)

**Environment Variables:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=herlangga.wicaksono
export DB_PASSWORD=your_password
export DB_NAME=panda_pocket
```

### 2. `add_currencies.go`
A Go script that connects to the database and adds currencies programmatically.

**Usage:**
```bash
go run scripts/add_currencies.go
```

### 3. `add_currencies.sql`
A SQL script that adds currencies directly to the database.

**Usage:**
```bash
psql -h localhost -U herlangga.wicaksono -d panda_pocket -f scripts/add_currencies.sql
```

## Currencies Added

The scripts add 20 default currencies including:

- **Major Currencies**: USD, EUR, GBP, JPY, AUD, CAD, CHF, CNY
- **European Currencies**: SEK, NOK, DKK, PLN, CZK, HUF
- **Asian Currencies**: KRW, SGD, **IDR** (Indonesian Rupiah)
- **Other Currencies**: RUB, BRL, INR

## Features

- ✅ **Duplicate Prevention**: Checks for existing currencies before adding
- ✅ **Error Handling**: Comprehensive error handling and logging
- ✅ **Verification**: Confirms successful insertion
- ✅ **Flexible**: Multiple execution methods
- ✅ **Safe**: Won't duplicate existing currencies

## Troubleshooting

### Database Connection Issues
```bash
# Test database connection
psql -h localhost -U herlangga.wicaksono -d panda_pocket -c "SELECT 1;"
```

### Permission Issues
```bash
# Make script executable
chmod +x scripts/run_add_currencies.sh
```

### Go Dependencies
```bash
# Install required Go packages
go mod tidy
```

## Server Deployment

For server deployment, you can:

1. **Copy the scripts to your server**
2. **Set environment variables**
3. **Run the appropriate script**

```bash
# Example for server deployment
export DB_HOST=your-server-host
export DB_USER=your-db-user
export DB_PASSWORD=your-db-password
export DB_NAME=panda_pocket

./scripts/run_add_currencies.sh go
```
