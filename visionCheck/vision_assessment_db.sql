-- Vision Assessment Database
DROP DATABASE IF EXISTS vision_assessment_db;
CREATE DATABASE vision_assessment_db;
USE vision_assessment_db;

-- Creating the VisionResults Table
CREATE TABLE VisionResults (
    ID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    LeftEyeScore INT NOT NULL,
    RightEyeScore INT NOT NULL,
    Comments TEXT NOT NULL,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Inserting Sample Vision Test Records (Following Your Screenshot)
INSERT INTO VisionResults (UserID, LeftEyeScore, RightEyeScore, Comments, CreatedAt)
VALUES 
    (5, 5, 5, 'Your vision in both eyes seems to be slightly reduced.', '2025-02-12 15:32:10'),
    (5, 1, 1, 'Your vision in both eyes is significantly reduced.', '2025-02-12 15:32:48');

-- Select all data to verify
SELECT * FROM VisionResults;
