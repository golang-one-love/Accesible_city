-- Migration: Drop nickname column from users
-- Service: auth-service

ALTER TABLE users DROP COLUMN nickname;
