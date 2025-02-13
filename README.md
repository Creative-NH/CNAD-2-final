# CNAD-2-final





## Befrienders Fall Risk Self-Assessment System

### Overview

The Fall Risk Self-Assessment System enables seniors to perform self-assessments at home, evaluate their fall risk, and notify doctors or caregivers when needed. This system consists of multiple microservices, each responsible for specific functionalities, such as user management, risk assessment, vision assessment, notifications, and reporting.

## Databases

### User Database

- **Users** (*UserID, Name, Email, PasswordHash, DateOfBirth, PhoneNumber, Address, CreatedAt, UpdatedAt*)

### Doctor Database

- **Doctors** (*DoctorID, Name, Email, PasswordHash*)

### Self-Assessment Database

- **Assessments** (*AssessmentID, UserID, DateCreated, QuestionResponses, TotalScore, RiskLevel, Recommendation*)
- **Questions** (*QuestionID, QuestionContent, QuestionOptions*)

### Vision Assessment Database

- **VisionResults** (*ID, UserID, LeftEyeScore, RightEyeScore, Comments, CreatedAt*)

### Alert Database

- **Notifications** (*NotificationID, UserID, Message, SentAt*)
- **Alerts** (*AlertID, AssessmentID, Type, SentAt*)

## Microservices

### User Service

Handles user registration, authentication, and profile management.

#### API Endpoints

- **POST /api/user/register** – Registers a new user.
  - **Input:** JSON object containing `name`, `email`, `password`, `dateOfBirth`, `phoneNumber`, and `address`.
  - **Output:** Success message or validation errors.
- **POST /api/user/authenticate** – Authenticates an existing user.
  - **Input:** JSON object with `email` and `password`.
  - **Output:** Authentication token and user details.
- **POST /api/user/getUserDetails** – Retrieves user details.
  - **Input:** JSON object containing `user_id`.
  - **Output:** JSON object with user profile details.
- **PUT /api/user/updateUserDetails** – Updates user details.
  - **Input:** JSON object with `user_id` and updated profile details.
  - **Output:** Success message or error details.

### Doctor Service

Allows doctors to authenticate and access user assessments.

#### API Endpoints

- **POST /api/doctor/authenticate** – Authenticates a doctor.
  - **Input:** JSON object with `email` and `password`.
  - **Output:** Authentication token and doctor details.
- **POST /api/doctor/getDoctorDetails** – Retrieves doctor details.
  - **Input:** JSON object with `doctor_id`.
  - **Output:** JSON object with doctor profile details.

### Self-Assessment Service

Provides a questionnaire and records assessment results.

#### API Endpoints

- **GET /api/self-assessment/questionnaire** – Fetches assessment questions.
  - **Output:** JSON object containing questions and options.
- **POST /api/self-assessment/addAssessmentResults** – Submits assessment results.
  - **Input:** JSON object with `user_id` and `answers`.
  - **Output:** JSON object with `assessment_id`, `total_score`, `risk_level`, and `recommendation`.
- **POST /api/self-assessment/getLastAssessment** – Retrieves the last assessment result.
  - **Input:** JSON object with `user_id`.
  - **Output:** JSON object with latest assessment details.
- **POST /api/self-assessment/getAssessment** – Fetches a specific assessment result.
  - **Input:** JSON object with `assessment_id`.
  - **Output:** JSON object with assessment details.
- **POST /api/self-assessment/assessmentHistory** – Retrieves assessment history.
  - **Input:** JSON object with `user_id`.
  - **Output:** List of past assessments.

### Risk Assessment Service

Analyzes the fall risk of users based on questionnaire responses.

#### API Endpoints

- **POST /api/risk-assessment/analyzeRisk** – Processes and calculates fall risk.
  - **Input:** JSON object with `user_id` and `answers`.
  - **Output:** JSON object with `total_score`, `risk_level`, and `recommendation`.

### Vision Assessment Service

Handles vision test results for further risk evaluation.

#### API Endpoints

- **POST /api/vision-assessment/postVisionResult** – Stores vision test results.
  - **Input:** JSON object with `user_id`, `leftEyeScore`, `rightEyeScore`, and `comments`.
  - **Output:** Success message or error details.
- **GET /api/vision-assessment/getLatestResult** – Retrieves the latest vision test result.
  - **Input:** Query parameter `userID`.
  - **Output:** JSON object with latest vision assessment.
- **GET /api/vision-assessment/getAllVisionResults** – Fetches all vision test results for a user.
  - **Input:** Query parameter `userID`.
  - **Output:** List of vision test results.
- **POST /api/vision-assessment/getVisionResult** – Retrieves a specific vision test result.
  - **Input:** JSON object with `visionAssessment_id`.
  - **Output:** JSON object with vision test details.

### Alert Service

Manages notifications and alerts for users and doctors.

#### API Endpoints

- **POST /api/notifications/getNotifications** – Retrieves notifications for users.
  - **Input:** JSON object with `user_id`.
  - **Output:** List of notifications.
- **POST /api/notifications/postNotifications** – Sends notifications to users.
  - **Input:** JSON object with `user_id` and `message`.
  - **Output:** Success message.
- **GET /api/notifications/getAlerts** – Retrieves alerts for doctors.
  - **Output:** List of recent alerts.
- **POST /api/notifications/postAlerts** – Sends alerts to doctors.
  - **Input:** JSON object with `assessment_id` and `type`.
  - **Output:** Success message.
- **DELETE /api/notifications/resolveAlerts/{assessment\_id}** – Resolves an alert related to an assessment.
  - **Output:** Success message.

### Email Service

Sends email notifications to doctors when urgent assessments are detected.

#### API Endpoints

- **POST /api/email/sendReportToDoctor** – Sends assessment reports to doctors.
  - **Input:** JSON object with assessment details.
  - **Output:** Success message.


## Instructions for Running Microservices 
Install NGINX, and replace nginx.conf with the file of the same name in this project.
Within nginx.conf, add the path to the project directory in lines 71 and 150.

Traverse to your nginx directory, then enter "nginx" to start the server.
```
cd ../nginx
nginx
```

This runs the web server on localhost:8555.

Start the microservices by traversing to each folder, and running the main.go file within.
```
cd ../Self Assessment
go run main.go
```


## Overview of Nginx Server
The Nginx server handles load balancing, failover, security, and performance optimization in the **Befrienders Fall-Risk Self-Assessment System**.

## Server Configuration

### Server Path
The server is configured to serve static files from:
```
../CNAD-2-final/Front-End/static
```

### Listening Port
Nginx listens on **port 8555**:
```
server {
    listen 8555;
    server_name localhost;
```

## Load Balancing Configuration
Load balancing is implemented using the **least connection** method to ensure efficient request distribution.

### Load Balancer with Health Checks
```
upstream user_service {
    least_conn;
    server localhost:8080 max_fails=3 fail_timeout=10s;
    server localhost:8081 backup;
}
```
- `least_conn`: Routes traffic to the server with the fewest connections.
- `max_fails=3 fail_timeout=10s`: Removes a failed server after 3 failures for 10 seconds.
- `backup`: Used only if primary servers fail.

This setup is applied to **all services**:
- `user_service (8080, 8081)`
- `self_assessment_service (8082, 8083)`
- `risk_assessment_service (8084, 8085)`
- `notifications_service (8086, 8087)`
- `doctor_service (8088, 8089)`
- `report_service (8090, 8091)`
- `vision_check_service (8092, 8093)`

## Request Rate Limiting
To prevent abuse, request rates are limited.
```
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=5r/s;
```
Each API request is subject to:
```
limit_req zone=api_limit burst=10 nodelay;
```
- `rate=5r/s`: Allows **5 requests per second** per IP.
- `burst=10`: Allows temporary bursts up to **10 requests**.

## API Proxying
All API requests are routed through Nginx to backend services:
```
location /api/user/ {
    proxy_pass http://user_service;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_connect_timeout 5s;
    proxy_read_timeout 10s;
}
```
This ensures:
- Correct forwarding of real client IPs.
- Fast failure detection with `proxy_connect_timeout 5s`.
- Requests are dropped if they take more than `10s` (`proxy_read_timeout 10s`).

## Failover & Circuit Breaker
If a server fails multiple times, it is temporarily removed.
```
server localhost:8080 max_fails=3 fail_timeout=10s;
```
- Ensures **automatic failover** to backup servers.
- Helps **prevent sending requests to failing services**.

## Security Measures
### Hide Nginx Version
To prevent attackers from detecting vulnerabilities:
```
server_tokens off;
```

### Restrict Access to Sensitive Routes
Doctor dashboard should only be accessed by trusted networks:
```
location /doctor/ {
    allow 192.168.1.0/24;
    deny all;
}
```

## Performance Optimizations
### Gzip Compression
Reduces bandwidth usage and speeds up responses:
```
gzip on;
gzip_types text/plain text/css application/json application/javascript;
gzip_vary on;
```

### Keepalive & Timeout Optimization
Idle connections are closed quickly to save resources:
```
keepalive_timeout 10;
proxy_connect_timeout 5s;
proxy_read_timeout 10s;
```