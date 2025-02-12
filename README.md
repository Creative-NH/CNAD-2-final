# CNAD-2-final

## Microservices

### Self Assessment

### Risk Assessment

### Vision Check

### User

### Doctor

### Alert

### Email

## Instructions for Running Microservices 

## Overview of Nginx Server
The Nginx server handles load balancing, failover, security, and performance optimization in the **Fall-Risk Self-Assessment System**.

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