CREATE DATABASE tag_db;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS requested_materials CASCADE;

DROP TABLE IF EXISTS incoming_materials CASCADE;

DROP TABLE IF EXISTS transactions_log CASCADE;

DROP TABLE IF EXISTS prices CASCADE;

DROP TABLE IF EXISTS material_usage_reasons CASCADE;

DROP TABLE IF EXISTS materials CASCADE;

DROP TABLE IF EXISTS customer_emails CASCADE;

DROP TABLE IF EXISTS customer_programs CASCADE;

DROP TABLE IF EXISTS locations CASCADE;

DROP TABLE IF EXISTS warehouses CASCADE;

DROP TABLE IF EXISTS customers CASCADE;

DROP TABLE IF EXISTS users CASCADE;

-- Drop types after all dependent tables are dropped
DROP TYPE IF EXISTS REQUEST_STATUS;

DROP TYPE IF EXISTS ROLE;

DROP TYPE IF EXISTS OWNER;

DROP TYPE IF EXISTS MATERIAL_TYPE;

DROP TYPE IF EXISTS REASON_TYPE;

CREATE TABLE IF NOT EXISTS customers (
	customer_id SERIAL PRIMARY KEY,
	customer_name VARCHAR(100) NOT NULL UNIQUE,
	user_id INT REFERENCES users (user_id) NOT NULL
);

CREATE TABLE IF NOT EXISTS customer_programs (
	program_id SERIAL PRIMARY KEY,
	program_name VARCHAR(100) NOT NULL UNIQUE,
	program_code VARCHAR(100) NOT NULL,
	is_active BOOLEAN,
	customer_id INT REFERENCES customers (customer_id) NOT NULL
);

CREATE TABLE IF NOT EXISTS customer_emails (
	id SERIAL PRIMARY KEY,
	customer_id INT NOT NULL REFERENCES customers (customer_id) ON DELETE CASCADE,
	email VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS warehouses (
	warehouse_id SERIAL PRIMARY KEY,
	name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
	location_id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	warehouse_id INT REFERENCES warehouses (warehouse_id) NOT NULL,
	CONSTRAINT unique_location_name_warehouse_id UNIQUE (name, warehouse_id)
);

CREATE TYPE MATERIAL_TYPE AS ENUM (
	'ACT LABEL',
	'BUBBLE',
	'BURGO',
	'CARRIER',
	'ENVELOPE',
	'FREE SHIPPING',
	'INSERT',
	'KEYCHAIN',
	'LABELS',
	'PAPER',
	'PRINT',
	'RIBBON',
	'SHIPPING',
	'STICKER',
	'WEARABLE',
	'CHIPS',
	'CARDS (METAL)',
	'CARDS (PVC)'
);

CREATE TYPE OWNER AS ENUM ('Tag', 'Customer');

CREATE TABLE IF NOT EXISTS materials (
	material_id SERIAL PRIMARY KEY,
	stock_id VARCHAR(100) NOT NULL,
	location_id INT REFERENCES locations (location_id) UNIQUE,
	program_id INT REFERENCES customer_programs (program_id) NOT NULL,
	material_type MATERIAL_TYPE NOT NULL,
	description TEXT,
	notes TEXT,
	quantity INT NOT NULL,
	min_required_quantity INT,
	max_required_quantity INT,
	updated_at DATE,
	is_active BOOLEAN NOT NULL,
	owner OWNER NOT NULL,
	is_primary BOOLEAN NOT NULL,
	serial_number_range VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS prices (
	price_id SERIAL PRIMARY KEY,
	material_id INT REFERENCES materials (material_id) ON DELETE CASCADE NOT NULL,
	quantity INT NOT NULL,
	cost DECIMAL NOT NULL,
	CONSTRAINT unique_material_id_cost UNIQUE (material_id, cost)
);

CREATE TYPE REASON_TYPE AS ENUM ('REMAKE');

CREATE TABLE IF NOT EXISTS material_usage_reasons (
	reason_id SERIAL PRIMARY KEY,
	reason_type REASON_TYPE NOT NULL,
	description VARCHAR(100) NOT NULL,
	code INT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions_log (
	transaction_id SERIAL PRIMARY KEY,
	price_id INT REFERENCES prices (price_id) ON DELETE CASCADE NOT NULL,
	quantity_change INT NOT NULL,
	notes TEXT,
	job_ticket VARCHAR(100),
	updated_at DATE,
	serial_number_range VARCHAR(100),
	reason_id INT REFERENCES material_usage_reasons (reason_id)
);

CREATE TABLE IF NOT EXISTS incoming_materials (
	shipping_id SERIAL PRIMARY KEY,
	program_id INT REFERENCES customer_programs (program_id) NOT NULL,
	stock_id VARCHAR(100) NOT NULL,
	cost DECIMAL NOT NULL,
	quantity INT NOT NULL,
	min_required_quantity INT,
	max_required_quantity INT,
	description TEXT NOT NULL,
	is_active BOOLEAN NOT NULL,
	type VARCHAR(100) NOT NULL,
	owner OWNER NOT NULL,
	user_id INT REFERENCES users (user_id) NOT NULL
);

CREATE TYPE ROLE AS ENUM (
	'admin',
	'warehouse',
	'csr',
	'production',
	'vault'
);

CREATE TABLE IF NOT EXISTS users (
	user_id SERIAL PRIMARY KEY,
	username VARCHAR(100) NOT NULL,
	password VARCHAR(100) NOT NULL,
	role ROLE NOT NULL,
	email VARCHAR(100) UNIQUE
);

CREATE TYPE REQUEST_STATUS AS ENUM ('pending', 'sent', 'declined');

CREATE TABLE IF NOT EXISTS requested_materials (
	request_id SERIAL PRIMARY KEY,
	user_id INT REFERENCES users (user_id),
	stock_id VARCHAR(100) NOT NULL,
	description TEXT NOT NULL,
	quantity_requested INT NOT NULL,
	quantity_used INT NOT NULL,
	status REQUEST_STATUS NOT NULL,
	notes TEXT NOT NULL,
	updated_at DATE,
	requested_at DATE
);