-- =============================================================================
-- Accessible Path - databases for MANAGED PostgreSQL (run in Beget panel SQL
-- console as the superuser of the server, or via psql)
-- Creates 6 databases and 6 service users.
-- Replace the passwords with the ones from .env.prod before running!
-- =============================================================================

CREATE USER auth_user        WITH ENCRYPTED PASSWORD 'change-me-strong-1';
CREATE DATABASE auth_db      OWNER auth_user;

CREATE USER barrier_user     WITH ENCRYPTED PASSWORD 'change-me-strong-2';
CREATE DATABASE barrier_db   OWNER barrier_user;

CREATE USER moderation_user  WITH ENCRYPTED PASSWORD 'change-me-strong-3';
CREATE DATABASE moderation_db OWNER moderation_user;

CREATE USER route_user       WITH ENCRYPTED PASSWORD 'change-me-strong-4';
CREATE DATABASE route_db     OWNER route_user;

CREATE USER poi_user         WITH ENCRYPTED PASSWORD 'change-me-strong-5';
CREATE DATABASE poi_db       OWNER poi_user;

CREATE USER notification_user WITH ENCRYPTED PASSWORD 'change-me-strong-6';
CREATE DATABASE notification_db OWNER notification_user;

-- Tables are created automatically by each service on first start
-- (Go services: EnsureSchema; Python services: metadata.create_all).
-- The PostGIS extension is NOT required at runtime.
