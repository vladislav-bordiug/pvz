CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('employee', 'moderator'))
);

CREATE TABLE IF NOT EXISTS pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMPTZ NOT NULL,
    city VARCHAR(50) NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
);

CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMPTZ NOT NULL,
    pvz_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('in_progress', 'close')),
    FOREIGN KEY (pvz_id) REFERENCES pvz(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMPTZ NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID NOT NULL,
    FOREIGN KEY (reception_id) REFERENCES receptions(id) ON DELETE CASCADE
);