-- PostgreSQL initialization script for "Доступный путь"
-- Creates 6 isolated databases, one per service

-- Auth Service Database
CREATE DATABASE auth_db;
CREATE USER auth_user WITH ENCRYPTED PASSWORD 'auth_pass';
GRANT ALL PRIVILEGES ON DATABASE auth_db TO auth_user;
ALTER DATABASE auth_db OWNER TO auth_user;

-- Barrier Service Database
CREATE DATABASE barrier_db;
CREATE USER barrier_user WITH ENCRYPTED PASSWORD 'barrier_pass';
GRANT ALL PRIVILEGES ON DATABASE barrier_db TO barrier_user;
ALTER DATABASE barrier_db OWNER TO barrier_user;

-- Moderation Service Database
CREATE DATABASE moderation_db;
CREATE USER moderation_user WITH ENCRYPTED PASSWORD 'moderation_pass';
GRANT ALL PRIVILEGES ON DATABASE moderation_db TO moderation_user;
ALTER DATABASE moderation_db OWNER TO moderation_user;

-- Route Service Database
CREATE DATABASE route_db;
CREATE USER route_user WITH ENCRYPTED PASSWORD 'route_pass';
GRANT ALL PRIVILEGES ON DATABASE route_db TO route_user;
ALTER DATABASE route_db OWNER TO route_user;

-- POI Service Database
CREATE DATABASE poi_db;
CREATE USER poi_user WITH ENCRYPTED PASSWORD 'poi_pass';
GRANT ALL PRIVILEGES ON DATABASE poi_db TO poi_user;
ALTER DATABASE poi_db OWNER TO poi_user;

-- Notification Service Database
CREATE DATABASE notification_db;
CREATE USER notification_user WITH ENCRYPTED PASSWORD 'notification_pass';
GRANT ALL PRIVILEGES ON DATABASE notification_db TO notification_user;
ALTER DATABASE notification_db OWNER TO notification_user;

-- Enable extensions that might be needed
-- NOTE: postgis is NOT available in the postgres:16-alpine image and is not
-- used by any service (tables are created via SQLAlchemy create_all), so it
-- must not be requested here.
\c auth_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c barrier_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c moderation_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c route_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c poi_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

\c notification_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";