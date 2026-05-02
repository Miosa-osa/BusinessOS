-- Migration: 104_seed_superadmin.sql
-- Description: Promotes Roberto's account to superadmin. Idempotent — safe to
--              re-run; only updates rows that are not already superadmin.
-- Created: 2026-04-11
--
-- NOTE: Run AFTER 103_platform_admin.sql (requires platform_role column).
-- Add additional email aliases below as needed.

UPDATE "user"
SET    platform_role = 'superadmin'
WHERE  email IN (
    'roberto@businessos.dev',
    'roberto@optimalos.dev',
    'roberto@miosa.ai',
    'roberto@lunivate.com',
    'rhl@miosa.ai'
)
AND platform_role != 'superadmin';
