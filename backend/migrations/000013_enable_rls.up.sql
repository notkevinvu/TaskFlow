-- Enable Row Level Security (RLS) on all tables
-- This addresses Supabase Security Advisor warnings about exposed tables
--
-- Architecture: Frontend -> Go Backend -> PostgreSQL
-- The Go backend uses the postgres superuser which bypasses RLS,
-- so enabling RLS without policies blocks PostgREST access while
-- maintaining backend functionality.

-- ============================================================
-- ENABLE RLS ON ALL TABLES
-- ============================================================

-- Core application tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_history ENABLE ROW LEVEL SECURITY;

-- Recurring tasks & preferences
ALTER TABLE task_series ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE category_preferences ENABLE ROW LEVEL SECURITY;

-- Dependencies & templates
ALTER TABLE task_dependencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_templates ENABLE ROW LEVEL SECURITY;

-- Gamification tables
ALTER TABLE user_achievements ENABLE ROW LEVEL SECURITY;
ALTER TABLE gamification_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE category_mastery ENABLE ROW LEVEL SECURITY;

-- System/utility tables
ALTER TABLE anonymous_user_cleanups ENABLE ROW LEVEL SECURITY;
ALTER TABLE schema_migrations ENABLE ROW LEVEL SECURITY;

-- ============================================================
-- REVOKE POSTGREST ACCESS (Belt and Suspenders)
-- ============================================================
-- Even with RLS enabled and no policies, explicitly revoking
-- access ensures PostgREST cannot query these tables.

-- Revoke from 'anon' role (unauthenticated Supabase requests)
REVOKE ALL ON users FROM anon;
REVOKE ALL ON tasks FROM anon;
REVOKE ALL ON task_history FROM anon;
REVOKE ALL ON task_series FROM anon;
REVOKE ALL ON user_preferences FROM anon;
REVOKE ALL ON category_preferences FROM anon;
REVOKE ALL ON task_dependencies FROM anon;
REVOKE ALL ON task_templates FROM anon;
REVOKE ALL ON user_achievements FROM anon;
REVOKE ALL ON gamification_stats FROM anon;
REVOKE ALL ON category_mastery FROM anon;
REVOKE ALL ON anonymous_user_cleanups FROM anon;
REVOKE ALL ON schema_migrations FROM anon;

-- Revoke from 'authenticated' role (authenticated Supabase requests)
REVOKE ALL ON users FROM authenticated;
REVOKE ALL ON tasks FROM authenticated;
REVOKE ALL ON task_history FROM authenticated;
REVOKE ALL ON task_series FROM authenticated;
REVOKE ALL ON user_preferences FROM authenticated;
REVOKE ALL ON category_preferences FROM authenticated;
REVOKE ALL ON task_dependencies FROM authenticated;
REVOKE ALL ON task_templates FROM authenticated;
REVOKE ALL ON user_achievements FROM authenticated;
REVOKE ALL ON gamification_stats FROM authenticated;
REVOKE ALL ON category_mastery FROM authenticated;
REVOKE ALL ON anonymous_user_cleanups FROM authenticated;
REVOKE ALL ON schema_migrations FROM authenticated;

-- ============================================================
-- NOTES
-- ============================================================
-- 1. The Go backend connects as 'postgres' superuser which bypasses RLS
-- 2. No RLS policies are created because we don't want any PostgREST access
-- 3. If you later need PostgREST access, create specific policies for those tables
-- 4. The service_role key in Supabase also bypasses RLS
