-- Notifications and Alerts Service
DROP DATABASE IF EXISTS notifications_db;
CREATE DATABASE notifications_db;
USE notifications_db;

-- Notifications Table
CREATE TABLE Notifications (
    NotificationID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    Message TEXT NOT NULL,
    SentAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Alerts Table (Updated with `type`)
CREATE TABLE Alerts (
    AlertID INT AUTO_INCREMENT PRIMARY KEY,
    AssessmentID INT NOT NULL,
    Type ENUM('HealthAssessment', 'VisionAssessment') NOT NULL, -- Added Type Column
    SentAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
