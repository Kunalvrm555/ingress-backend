# Ingress Backend

This project provides backend services for a system managing ingress of students. It handles login, student logs, student statistics and allows student check-in and check-out.

This repository is a submodule of [**Project Ingress**](https://github.com/Kunalvrm555/project-ingress)

## Setup

### Prerequisites

1. GoLang
2. PostgreSQL

### Steps

1. Clone the repository.
```
git clone https://github.com/Kunalvrm555/ingress-backend/`
```

2. Change directory into the project's root directory.
```
cd ingress_backend
```

3. Set up environment variables in `.env` file.
```env
POSTGRES_USERNAME=<postgres_username>
POSTGRES_PASSWORD=<postgres_password>
ADMIN_USERNAME=<admin_username>
ADMIN_PASSWORD=<admin_password>
USERNAME=<username>
PASSWORD=<password>
JWT_SECRET_KEY=<jwt_secret_key>
```

4. Install the necessary Go dependencies.
```
go mod download
```
5. Set up the database.Create a new database named `ingress` and use your terminal or GUI to create the necessary tables using the SQL commands below:
```sql
CREATE TABLE students (
    rollno VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    dept VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL
);

CREATE TABLE logs (
    rollno VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    checkintime TIMESTAMP NOT NULL,
    checkouttime TIMESTAMP,
);

CREATE TABLE users (
    username VARCHAR(50) PRIMARY KEY,
    password VARCHAR(255) NOT NULL,
    usertype VARCHAR(50) NOT NULL
);
```
6. Run the application.

```
go run cmd/main.go
```

## Services

### Login Service
The login service accepts username and password, validates them and if successful, returns a JWT token.

### Logs Service
The logs service retrieves and returns a list of student logs - it lists students who have checked in but haven't checked out yet.

### Statistics Service
The statistics service retrieves and returns the total count of students who have checked in for the day and the count of students who are currently checked in.

### Student Service
The student service handles student check-in and check-out operations.

## Seed Users
This application uses a utility script `seedUser.go` located in the util folder to seed an initial set of users in the database. The script picks up the credentials from the environment variables.

Note: The students data have to be populated manually or you could write a script to seed the data. The SeedUsers function only seeds the users data.










