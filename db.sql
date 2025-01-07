CREATE DATABASE tag_db;

DROP TABLE IF EXISTS transactions_log;

DROP TABLE IF EXISTS prices;

DROP TABLE IF EXISTS materials;

DROP TABLE IF EXISTS incoming_materials;

DROP TABLE IF EXISTS customers;

DROP TABLE IF EXISTS locations;

DROP TABLE IF EXISTS warehouses;

DROP TYPE IF EXISTS material_type;

DROP TYPE IF EXISTS owner;

CREATE TABLE IF NOT EXISTS customers (
	customer_id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL UNIQUE,
	customer_code VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS warehouses (
	warehouse_id SERIAL PRIMARY KEY,
	name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS locations (
	location_id SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	warehouse_id INT REFERENCES warehouses (warehouse_id),
	CONSTRAINT unique_location_name_warehouse_id UNIQUE (name, warehouse_id)
);

CREATE TYPE material_type AS ENUM (
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
	'WEARABLE'
);

CREATE TYPE owner AS ENUM ('Tag', 'Customer');

CREATE TABLE IF NOT EXISTS materials (
	material_id SERIAL PRIMARY KEY,
	stock_id VARCHAR(100) NOT NULL,
	location_id INT REFERENCES locations (location_id) UNIQUE,
	customer_id INT REFERENCES customers (customer_id),
	material_type MATERIAL_TYPE NOT NULL,
	description TEXT,
	notes TEXT,
	quantity INT NOT NULL,
	min_required_quantity INT,
	max_required_quantity INT,
	updated_at DATE,
	is_active BOOLEAN NOT NULL,
	owner OWNER NOT NULL,
	is_primary BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS prices (
	price_id SERIAL PRIMARY KEY,
	material_id INT REFERENCES materials (material_id) ON DELETE CASCADE,
	quantity INT NOT NULL,
	cost DECIMAL NOT NULL,
	CONSTRAINT unique_material_id_cost UNIQUE (material_id, cost)
);

CREATE TABLE IF NOT EXISTS transactions_log (
	transaction_id SERIAL PRIMARY KEY,
	price_id INT REFERENCES prices (price_id) ON DELETE CASCADE,
	quantity_change INT NOT NULL,
	notes TEXT,
	job_ticket VARCHAR(100),
	updated_at DATE
);

CREATE TABLE IF NOT EXISTS incoming_materials (
	shipping_id SERIAL PRIMARY KEY,
	customer_id INT REFERENCES customers (customer_id),
	stock_id VARCHAR(100) NOT NULL,
	cost DECIMAL NOT NULL,
	quantity INT NOT NULL,
	min_required_quantity INT,
	max_required_quantity INT,
	description TEXT,
	is_active BOOLEAN NOT NULL,
	type VARCHAR(100) NOT NULL,
	owner OWNER NOT NULL
);