-- Rollback: Disable Row Level Security on all tables
-- WARNING: This re-exposes tables to PostgREST API access

-- ============================================================
-- RESTORE POSTGREST ACCESS
-- ============================================================
-- Re-grant access to anon and authenticated roles
-- (Grants SELECT, INSERT, UPDATE, DELETE which is Supabase default)

-- Grant to 'anon' role
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON tasks TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_history TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_series TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_preferences TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON category_preferences TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_dependencies TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_templates TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_achievements TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON gamification_stats TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON category_mastery TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON anonymous_user_cleanups TO anon;
GRANT SELECT, INSERT, UPDATE, DELETE ON schema_migrations TO anon;

-- Grant to 'authenticated' role
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON tasks TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_history TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_series TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_preferences TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON category_preferences TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_dependencies TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON task_templates TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON user_achievements TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON gamification_stats TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON category_mastery TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON anonymous_user_cleanups TO authenticated;
GRANT SELECT, INSERT, UPDATE, DELETE ON schema_migrations TO authenticated;

-- ============================================================
-- DISABLE RLS ON ALL TABLES
-- ============================================================

-- Core application tables
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE tasks DISABLE ROW LEVEL SECURITY;
ALTER TABLE task_history DISABLE ROW LEVEL SECURITY;

-- Recurring tasks & preferences
ALTER TABLE task_series DISABLE ROW LEVEL SECURITY;
ALTER TABLE user_preferences DISABLE ROW LEVEL SECURITY;
ALTER TABLE category_preferences DISABLE ROW LEVEL SECURITY;

-- Dependencies & templates
ALTER TABLE task_dependencies DISABLE ROW LEVEL SECURITY;
ALTER TABLE task_templates DISABLE ROW LEVEL SECURITY;

-- Gamification tables
ALTER TABLE user_achievements DISABLE ROW LEVEL SECURITY;
ALTER TABLE gamification_stats DISABLE ROW LEVEL SECURITY;
ALTER TABLE category_mastery DISABLE ROW LEVEL SECURITY;

-- System/utility tables
ALTER TABLE anonymous_user_cleanups DISABLE ROW LEVEL SECURITY;
ALTER TABLE schema_migrations DISABLE ROW LEVEL SECURITY;
