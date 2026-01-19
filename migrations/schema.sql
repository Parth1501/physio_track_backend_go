-- Oracle-compatible schema

-- USERS
BEGIN
  EXECUTE IMMEDIATE '
    CREATE TABLE users (
      id VARCHAR2(36) PRIMARY KEY,
      username VARCHAR2(255) UNIQUE NOT NULL,
      password_hash VARCHAR2(255) NOT NULL,
      created_time TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL
    )';
EXCEPTION
  WHEN OTHERS THEN
    IF SQLCODE != -955 THEN RAISE; END IF;
END;
/

-- PATIENTS
BEGIN
  EXECUTE IMMEDIATE '
    CREATE TABLE patients (
      id VARCHAR2(36) PRIMARY KEY,
      full_name VARCHAR2(255) NOT NULL,
      phone_number VARCHAR2(64),
      age NUMBER,
      gender VARCHAR2(50),
      chief_complaint VARCHAR2(4000),
      present_history VARCHAR2(4000),
      medical_history VARCHAR2(4000),
      observation VARCHAR2(4000),
      palpation VARCHAR2(4000),
      examination VARCHAR2(4000),
      rehab VARCHAR2(4000),
      diagnosis VARCHAR2(4000),
      created_time TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
      updated_time TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
      last_paid_amount NUMBER,
      status VARCHAR2(100)
    )';
EXCEPTION
  WHEN OTHERS THEN
    IF SQLCODE != -955 THEN RAISE; END IF;
END;
/

-- PAYMENTS
BEGIN
  EXECUTE IMMEDIATE '
    CREATE TABLE payments (
      id VARCHAR2(36) PRIMARY KEY,
      patient_id VARCHAR2(36) NOT NULL,
      amount NUMBER NOT NULL,
      payment_mode VARCHAR2(100),
      paid_date DATE,
      CONSTRAINT fk_payment_patient FOREIGN KEY (patient_id) REFERENCES patients(id) ON DELETE CASCADE
    )';
EXCEPTION
  WHEN OTHERS THEN
    IF SQLCODE != -955 THEN RAISE; END IF;
END;
/

-- Indexes (ignore if already exist)
BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_patients_phone ON patients(phone_number)';
EXCEPTION
  WHEN OTHERS THEN
    IF SQLCODE != -955 AND SQLCODE != -1408 THEN RAISE; END IF;
END;
/
BEGIN
  EXECUTE IMMEDIATE 'CREATE INDEX idx_payments_patient ON payments(patient_id)';
EXCEPTION
  WHEN OTHERS THEN
    IF SQLCODE != -955 AND SQLCODE != -1408 THEN RAISE; END IF;
END;
/
