-- Doctor Service
DROP database IF EXISTS doctor_db;
CREATE database doctor_db;
USE doctor_db;

-- Table for storing doctor information
CREATE TABLE IF NOT EXISTS Doctors (
    DoctorID INT AUTO_INCREMENT PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Email VARCHAR(100) UNIQUE NOT NULL,
    PasswordHash VARCHAR(255) NOT NULL -- Changed to store hashed passwords
);

-- Insert a test doctor with a hashed password
INSERT INTO Doctors (Name, Email, PasswordHash) VALUES
('Dr. John Doe', 'johndoe7@gmail.com', '$2a$10$.Nl5tbg7enAt9PUEG.DBe.DHj0vRrldExrcaLN9e9I6dZ.bMPy/ha');
