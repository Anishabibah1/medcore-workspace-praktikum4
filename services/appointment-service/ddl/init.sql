CREATE TABLE IF NOT EXISTS appointments (
  id SERIAL PRIMARY KEY, patient_id INT NOT NULL,
  doctor VARCHAR(100) NOT NULL, scheduled_at TIMESTAMP NOT NULL);
INSERT INTO appointments (patient_id, doctor, scheduled_at) VALUES (1, 'dr. Wayan', NOW());
