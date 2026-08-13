-- Migration: Add nickname column to users
-- Service: auth-service

ALTER TABLE users ADD COLUMN nickname VARCHAR(100) NOT NULL DEFAULT '';
