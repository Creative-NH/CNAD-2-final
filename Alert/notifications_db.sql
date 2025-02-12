-- Notifications and Alerts Service
DROP database IF EXISTS notifications_db;
CREATE database notifications_db;
USE notifications_db;

CREATE TABLE Notifications (
    NotificationID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    Message TEXT NOT NULL,
    SentAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Alerts (
    AlertID INT AUTO_INCREMENT PRIMARY KEY,
    AssessmentID INT NOT NULL,
    SentAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);