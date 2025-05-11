# Student Service

This is the Student Service component of the Student Vaccination Portal microservices system. It manages student information, vaccination records, and supports bulk upload functionality using MinIO. The service also communicates with the Vaccine Service and utilizes RabbitMQ for messaging.

## Prerequisites

- Go (1.23+ recommended)
- MySQL 
- RabbitMQ
- MinIO (for file storage)
- Vaccine Service running and accessible

## Setup Instructions

### 1. Create `.env` File

Create a `.env` file in the project root directory and fill in the following configuration variables:

```env
DB_HOST=
DB_PORT=
DB_USER=
DB_PASS=
DB_NAME=

VACCINE_SERVICE=

RABBIT_USER=
RABBIT_PASS=
RABBIT_HOST=
RABBIT_PORT=

MINIO_SERVER=
MINIO_PORT=
MINIO_USERNAME=
MINIO_PASSWORD=
MINIO_BULK_UPLOAD_BUCKET=
MINIO_REGION=
```
### 2. Install Go Dependencies
Run the following command to download and tidy Go module dependencies:
```
go mod tidy
```
### 3. Database Setup
Before running the service, ensure the following tables are created in your database:
```
CREATE TABLE students (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  class VARCHAR(20) NOT NULL,
  gender VARCHAR(10),
  roll_number VARCHAR(50) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  phone_no VARCHAR(20)
);

CREATE TABLE vaccination_records (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  student_id INT NOT NULL,
  drive_id INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (student_id) REFERENCES students(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);

CREATE TABLE bulk_upload_jobs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  file_name VARCHAR(255) NOT NULL,
  file_path TEXT NOT NULL,
  status VARCHAR(50) DEFAULT 'PENDING',
  error_message TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  processed_records INT DEFAULT 0,
  total_records INT DEFAULT 0,
  request_id VARCHAR(200),
  request_type VARCHAR(20)
);
```
### 4. Run the Student Service
You can run the Student Service in two modes depending on the functionality you want to execute:

To run the Student Service Server:
```
go run cmd/main.go --service=server
```
To run the Bulk File Processor:
```
go run cmd/main.go --service=bulkprocessor
```
