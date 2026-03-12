-- Migration: Add owner_id to shops table for multi-tenant SaaS access control
-- Run this manually if the automatic migration (AddOwnerIDToShops) does not run on startup.
-- owner_id is INTEGER to match users.id (SERIAL).

ALTER TABLE shops
ADD COLUMN owner_id INTEGER;

ALTER TABLE shops
ADD CONSTRAINT fk_shop_owner
FOREIGN KEY (owner_id)
REFERENCES users(id)
ON DELETE CASCADE;
