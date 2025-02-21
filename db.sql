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
	customer_code VARCHAR(100) NOT NULL
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
	'CHIPS'
);

CREATE TYPE OWNER AS ENUM ('Tag', 'Customer');

CREATE TABLE IF NOT EXISTS materials (
	material_id SERIAL PRIMARY KEY,
	stock_id VARCHAR(100) NOT NULL,
	location_id INT REFERENCES locations (location_id) UNIQUE,
	customer_id INT REFERENCES customers (customer_id) NOT NULL,
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

CREATE TABLE IF NOT EXISTS transactions_log (
	transaction_id SERIAL PRIMARY KEY,
	price_id INT REFERENCES prices (price_id) ON DELETE CASCADE NOT NULL,
	quantity_change INT NOT NULL,
	notes TEXT,
	job_ticket VARCHAR(100),
	updated_at DATE,
	serial_number_range VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS incoming_materials (
	shipping_id SERIAL PRIMARY KEY,
	customer_id INT REFERENCES customers (customer_id) NOT NULL,
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

CREATE TYPE ROLE AS ENUM ('admin', 'warehouse', 'csr', 'production');

CREATE TABLE IF NOT EXISTS users (
	user_id SERIAL PRIMARY KEY,
	username VARCHAR(100) NOT NULL,
	password VARCHAR(100) NOT NULL,
	role ROLE NOT NULL
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