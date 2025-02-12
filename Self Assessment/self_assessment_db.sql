-- Self-Assessment Service
DROP database IF EXISTS self_assessment_db;
CREATE database self_assessment_db;
USE self_assessment_db;

CREATE TABLE Assessments (
    AssessmentID INT AUTO_INCREMENT PRIMARY KEY,
    UserID INT NOT NULL,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    QuestionResponses TEXT NOT NULL,
    TotalScore INT,
    RiskLevel ENUM('Low', 'Moderate', 'High'),
    Recommendation TEXT
);

-- English Questions
CREATE TABLE QuestionsEn (
    QuestionID INT AUTO_INCREMENT PRIMARY KEY,
    QuestionContent VARCHAR(500) NOT NULL, 
    QuestionOptions VARCHAR(500) NOT NULL 
);

-- Chinese (Simplified) Questions
CREATE TABLE QuestionsCn (
    QuestionID INT AUTO_INCREMENT PRIMARY KEY,
    QuestionContent VARCHAR(500) NOT NULL, 
    QuestionOptions VARCHAR(500) NOT NULL 
);

-- Malay Questions
CREATE TABLE QuestionsMy (
    QuestionID INT AUTO_INCREMENT PRIMARY KEY,
    QuestionContent VARCHAR(500) NOT NULL, 
    QuestionOptions VARCHAR(500) NOT NULL 
);

-- Tamil Questions
CREATE TABLE QuestionsTa (
    QuestionID INT AUTO_INCREMENT PRIMARY KEY,
    QuestionContent VARCHAR(500) NOT NULL, 
    QuestionOptions VARCHAR(500) NOT NULL 
);


INSERT INTO Assessments (UserID, QuestionResponses, TotalScore, RiskLevel, Recommendation) VALUES
(1, '{ 1: 2, 2: 3, 3: 2, 4: 1, 5: 3, 6: 2, 7: 2, 8: 3, 9: 2, 10: 1 }', 15, 'Moderate', 'Consider physical therapy, improve home safety, and monitor medications.'),
(1, '{ 1: 2, 2: 2, 3: 1, 4: 2, 5: 1, 6: 2, 7: 2, 8: 2, 9: 2, 10: 1 }', 8, 'Low', 'Maintain a healthy lifestyle and exercise regularly.'),
(2, '{ 1: 1, 2: 1, 3: 2, 4: 4, 5: 3, 6: 1, 7: 2, 8: 2, 9: 1, 10: 3 }', 18, 'High', 'Consult a healthcare provider for a fall risk assessment and use mobility aids.');

INSERT INTO QuestionsEn (QuestionContent, QuestionOptions) 
VALUES 
("Do you experience dizziness?", '["Yes", "No"]'),
("How is your balance?", '["Good", "Moderate", "Poor"]'),
("How many times have you fallen in the past year?", '["0", "1-2", "3 or more"]'),
("Do you use any mobility aids?", '["None", "Cane", "Walker", "Wheelchair"]'),
("Do you feel unsteady when walking?", '["Never", "Sometimes", "Often", "Always"]'),
("Have you had a fall in the past 6 months?", '["Yes", "No"]'),
("Are you able to stand up from a chair without using your hands?", '["Yes", "No"]'),
("Do you take medications that cause dizziness?", '["Yes", "No", "Not sure"]'),
("Do you exercise regularly?", '["Yes", "No"]'),
("Do you experience numbness in your feet?", '["Yes", "No", "Sometimes"]');
INSERT INTO QuestionsCn (QuestionContent, QuestionOptions) 
VALUES 
("你是否感到头晕？", '["是", "否"]'),
("你的平衡能力如何？", '["良好", "一般", "差"]'),
("过去一年内你跌倒过几次？", '["0", "1-2", "3次或更多"]'),
("你使用助行器具吗？", '["无", "手杖", "助行器", "轮椅"]'),
("走路时你会感到不稳吗？", '["从不", "有时", "经常", "总是"]'),
("过去6个月内你是否跌倒过？", '["是", "否"]'),
("你能不用手站起来吗？", '["是", "否"]'),
("你是否服用会导致头晕的药物？", '["是", "否", "不确定"]'),
("你是否定期运动？", '["是", "否"]'),
("你的脚是否会感到麻木？", '["是", "否", "有时"]');
INSERT INTO QuestionsMy (QuestionContent, QuestionOptions) 
VALUES 
("Adakah anda mengalami pening?", '["Ya", "Tidak"]'),
("Bagaimanakah keseimbangan anda?", '["Baik", "Sederhana", "Buruk"]'),
("Berapa kali anda terjatuh dalam setahun yang lalu?", '["0", "1-2", "3 atau lebih"]'),
("Adakah anda menggunakan alat bantuan pergerakan?", '["Tiada", "Tongkat", "Walker", "Kerusi roda"]'),
("Adakah anda berasa tidak stabil semasa berjalan?", '["Tidak pernah", "Kadang-kadang", "Selalu", "Setiap masa"]'),
("Adakah anda pernah jatuh dalam 6 bulan terakhir?", '["Ya", "Tidak"]'),
("Bolehkah anda bangun dari kerusi tanpa menggunakan tangan?", '["Ya", "Tidak"]'),
("Adakah anda mengambil ubat yang menyebabkan pening?", '["Ya", "Tidak", "Tidak pasti"]'),
("Adakah anda bersenam secara berkala?", '["Ya", "Tidak"]'),
("Adakah anda mengalami kebas di kaki anda?", '["Ya", "Tidak", "Kadang-kadang"]');
INSERT INTO QuestionsTa (QuestionContent, QuestionOptions) 
VALUES 
("நீங்கள் மயக்கம் உணர்கிறீர்களா?", '["ஆம்", "இல்லை"]'),
("உங்கள் சமநிலை எப்படி உள்ளது?", '["நல்லது", "மிதமானது", "மோசமானது"]'),
("கடந்த ஆண்டு நீங்கள் எத்தனை முறை கீழே விழுந்தீர்கள்?", '["0", "1-2", "3 அல்லது அதற்கு மேல்"]'),
("நீங்கள் நகர்வதற்கு உதவிகள் பயன்படுத்துகிறீர்களா?", '["இல்லை", "சங்கில்", "நடக்க உதவும் கருவி", "சக்கர நாற்காலி"]'),
("நடக்கும் போது நீங்கள் நிலை தடுமாறுகிறீர்களா?", '["ஒருபோதும் இல்லை", "சில சமயங்களில்", "அடிக்கடி", "எப்போதும்"]'),
("கடந்த 6 மாதங்களில் நீங்கள் கீழே விழுந்தீர்களா?", '["ஆம்", "இல்லை"]'),
("நீங்கள் கைகளைப் பயன்படுத்தாமல் நாற்காலியில் இருந்து எழுந்திருக்க முடியுமா?", '["ஆம்", "இல்லை"]'),
("நீங்கள் மயக்கத்தை ஏற்படுத்தும் மருந்துகளை எடுத்துக்கொள்கிறீர்களா?", '["ஆம்", "இல்லை", "தெரியாது"]'),
("நீங்கள் முறையாக உடற்பயிற்சி செய்கிறீர்களா?", '["ஆம்", "இல்லை"]'),
("உங்கள் கால்களில் உணர்விழப்பு இருக்கிறதா?", '["ஆம்", "இல்லை", "சில சமயங்களில்"]');