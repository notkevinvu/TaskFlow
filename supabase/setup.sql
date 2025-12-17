-- ============================================================
-- TaskFlow - Supabase Database Setup
-- ============================================================
-- This consolidated script sets up a fresh Supabase database
-- with the complete TaskFlow schema (equivalent to all migrations).
--
-- Usage:
-- 1. Create a new Supabase project at https://supabase.com
-- 2. Go to SQL Editor in your Supabase dashboard
-- 3. Paste this entire script and click "Run"
-- 4. Copy your database connection string from Settings > Database
--
-- Version: 13 (includes RLS security)
-- Last Updated: 2025-12-16
-- ============================================================

-- ============================================================
-- EXTENSIONS
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- ENUM TYPES
-- ============================================================

-- Task status workflow
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done', 'on_hold', 'blocked');
COMMENT ON TYPE task_status IS 'Task status: todo, in_progress, done, on_hold (paused), blocked (manually marked as blocked)';

-- Task effort estimation
CREATE TYPE task_effort AS ENUM ('small', 'medium', 'large', 'xlarge');

-- Task history event types
CREATE TYPE task_history_event_type AS ENUM (
    'created',
    'updated',
    'bumped',
    'completed',
    'deleted',
    'status_changed'
);

-- Recurring task patterns
CREATE TYPE recurrence_pattern AS ENUM ('none', 'daily', 'weekly', 'monthly');

-- Due date calculation strategy
CREATE TYPE due_date_calculation AS ENUM ('from_original', 'from_completion');

-- Task relationship types
CREATE TYPE task_type AS ENUM ('regular', 'recurring', 'subtask');

-- User account types
CREATE TYPE user_type AS ENUM ('registered', 'anonymous');

-- Achievement types for gamification
CREATE TYPE achievement_type AS ENUM (
    'first_task',
    'milestone_10',
    'milestone_50',
    'milestone_100',
    'streak_3',
    'streak_7',
    'streak_14',
    'streak_30',
    'category_master',
    'speed_demon',
    'consistency_king'
);

-- ============================================================
-- CORE TABLES
-- ============================================================

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    password_hash VARCHAR(255),
    user_type user_type NOT NULL DEFAULT 'registered',
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User validation constraints
ALTER TABLE users ADD CONSTRAINT users_type_credentials_check
    CHECK (
        (user_type = 'registered' AND email IS NOT NULL AND password_hash IS NOT NULL AND name IS NOT NULL)
        OR (user_type = 'anonymous')
    );

ALTER TABLE users ADD CONSTRAINT users_type_expiry_check
    CHECK (
        (user_type = 'registered' AND expires_at IS NULL)
        OR (user_type = 'anonymous' AND expires_at IS NOT NULL)
    );

-- User indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_user_type ON users(user_type);
CREATE INDEX idx_users_anonymous_expires_at ON users(expires_at)
    WHERE user_type = 'anonymous' AND expires_at IS NOT NULL;

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status task_status NOT NULL DEFAULT 'todo',
    user_priority INTEGER NOT NULL DEFAULT 5 CHECK (user_priority >= 1 AND user_priority <= 10),
    due_date TIMESTAMP WITH TIME ZONE,
    estimated_effort task_effort,
    category VARCHAR(50),
    context TEXT,
    related_people TEXT[],
    priority_score INTEGER NOT NULL DEFAULT 50 CHECK (priority_score >= 0 AND priority_score <= 100),
    bump_count INTEGER NOT NULL DEFAULT 0,
    search_vector tsvector,
    task_type task_type NOT NULL DEFAULT 'regular',
    series_id UUID,
    parent_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

COMMENT ON COLUMN tasks.parent_task_id IS
    'Parent task reference. Interpretation depends on task_type:
     - task_type=recurring: links to previous occurrence in series (temporal chain)
     - task_type=subtask: links to parent task (hierarchical relationship)
     - task_type=regular: should be NULL';

COMMENT ON COLUMN tasks.task_type IS
    'Discriminator for task relationships:
     - regular: standalone task (default)
     - recurring: part of a recurring series (has series_id)
     - subtask: child of another task (has parent_task_id, no series_id)';

COMMENT ON COLUMN tasks.deleted_at IS 'Soft delete timestamp - NULL means not deleted, set value means deleted at that time';

-- Task indexes
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority_score ON tasks(priority_score DESC);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_category ON tasks(category);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_search_vector ON tasks USING GIN(search_vector);
CREATE INDEX idx_tasks_series_id ON tasks(series_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);
CREATE INDEX idx_tasks_subtask_lookup ON tasks(parent_task_id, task_type) WHERE task_type = 'subtask';
CREATE INDEX idx_tasks_not_deleted ON tasks(user_id, status) WHERE deleted_at IS NULL;

-- Composite indexes for query performance
CREATE INDEX idx_tasks_user_status_priority ON tasks(user_id, status, priority_score DESC);
CREATE INDEX idx_tasks_user_category_priority ON tasks(user_id, category, priority_score DESC) WHERE category IS NOT NULL;
CREATE INDEX idx_tasks_user_status_due_date ON tasks(user_id, status, due_date);

-- Optimized partial index for main task list
CREATE INDEX idx_tasks_user_active_main ON tasks(user_id, priority_score DESC, created_at DESC)
    WHERE status != 'done' AND deleted_at IS NULL AND (task_type IS NULL OR task_type != 'subtask');

-- Index for analytics queries
CREATE INDEX idx_tasks_user_analytics ON tasks(user_id, bump_count) WHERE deleted_at IS NULL;

-- Task history table (audit log)
CREATE TABLE task_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event_type task_history_event_type NOT NULL,
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_history_task_id ON task_history(task_id);
CREATE INDEX idx_task_history_user_id ON task_history(user_id);
CREATE INDEX idx_task_history_event_type ON task_history(event_type);
CREATE INDEX idx_task_history_created_at ON task_history(created_at DESC);

-- ============================================================
-- RECURRING TASKS & PREFERENCES
-- ============================================================

-- Task series (recurring task groups)
CREATE TABLE task_series (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_task_id UUID NOT NULL,
    pattern recurrence_pattern NOT NULL DEFAULT 'none',
    interval_value INTEGER NOT NULL DEFAULT 1 CHECK (interval_value > 0 AND interval_value <= 365),
    end_date TIMESTAMP WITH TIME ZONE,
    due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_series_user_id ON task_series(user_id);
CREATE INDEX idx_task_series_original_task_id ON task_series(original_task_id);
CREATE INDEX idx_task_series_is_active ON task_series(is_active) WHERE is_active = true;

-- Add foreign key from tasks.series_id to task_series
ALTER TABLE tasks ADD CONSTRAINT fk_tasks_series FOREIGN KEY (series_id) REFERENCES task_series(id) ON DELETE SET NULL;

-- Add foreign key from task_series.original_task_id to tasks
ALTER TABLE task_series ADD CONSTRAINT fk_task_series_original_task
    FOREIGN KEY (original_task_id) REFERENCES tasks(id) ON DELETE CASCADE;

-- User preferences
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN user_preferences.timezone IS 'IANA timezone (e.g., America/New_York) for streak calculation';

-- Category-specific preferences
CREATE TABLE category_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, category)
);

CREATE INDEX idx_category_preferences_user_id ON category_preferences(user_id);

-- ============================================================
-- TASK DEPENDENCIES
-- ============================================================

CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocked_by_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, blocked_by_id),
    CONSTRAINT no_self_dependency CHECK (task_id != blocked_by_id)
);

COMMENT ON TABLE task_dependencies IS
    'Junction table for blocked-by task relationships.
     task_id is blocked by blocked_by_id.
     Only regular tasks can have/be dependencies (enforced at application layer).
     Cycles are prevented at application layer via graph traversal.';

COMMENT ON COLUMN task_dependencies.task_id IS 'The task that is blocked (cannot be completed until blocker is done)';
COMMENT ON COLUMN task_dependencies.blocked_by_id IS 'The task that is blocking (must be completed first)';

CREATE INDEX idx_task_dependencies_task ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_blocker ON task_dependencies(blocked_by_id);

-- ============================================================
-- TASK TEMPLATES
-- ============================================================

CREATE TABLE task_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    estimated_effort task_effort,
    user_priority INTEGER NOT NULL DEFAULT 5 CHECK (user_priority >= 1 AND user_priority <= 10),
    context VARCHAR(500),
    related_people TEXT[],
    due_date_offset INTEGER CHECK (due_date_offset IS NULL OR (due_date_offset >= 0 AND due_date_offset <= 365)),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_template_name UNIQUE(user_id, name)
);

COMMENT ON TABLE task_templates IS 'User-defined templates for quickly creating tasks with pre-filled values';
COMMENT ON COLUMN task_templates.name IS 'User-facing template name (e.g., "Weekly Report", "Bug Fix")';
COMMENT ON COLUMN task_templates.due_date_offset IS 'Days from creation (NULL = no due date, 7 = due in 1 week)';

CREATE INDEX idx_task_templates_user_id ON task_templates(user_id);
CREATE INDEX idx_task_templates_user_category ON task_templates(user_id, category) WHERE category IS NOT NULL;

-- ============================================================
-- GAMIFICATION TABLES
-- ============================================================

-- User achievements (earned badges)
CREATE TABLE user_achievements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_type achievement_type NOT NULL,
    earned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE user_achievements IS 'Stores unlocked achievement badges for gamification';

CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX idx_user_achievements_earned_at ON user_achievements(earned_at DESC);
CREATE INDEX idx_user_achievements_type ON user_achievements(achievement_type);

-- Unique constraint allowing multiple category_master per category
CREATE UNIQUE INDEX unique_achievement_per_user
    ON user_achievements(user_id, achievement_type, COALESCE((metadata->>'category')::text, ''));

-- Gamification stats (cached computed stats)
CREATE TABLE gamification_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_streak INTEGER NOT NULL DEFAULT 0,
    longest_streak INTEGER NOT NULL DEFAULT 0,
    last_completion_date DATE,
    total_completed INTEGER NOT NULL DEFAULT 0,
    productivity_score NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    completion_rate NUMERIC(5,2) DEFAULT 0.00,
    streak_score NUMERIC(5,2) DEFAULT 0.00,
    on_time_percentage NUMERIC(5,2) DEFAULT 0.00,
    effort_mix_score NUMERIC(5,2) DEFAULT 0.00,
    last_computed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE gamification_stats IS 'Cached productivity stats (streaks, scores) for fast dashboard loading';
COMMENT ON COLUMN gamification_stats.productivity_score IS 'Weighted score: 30% completion + 25% streak + 25% on-time + 20% effort mix';

CREATE INDEX idx_gamification_stats_productivity ON gamification_stats(productivity_score DESC);
CREATE INDEX idx_gamification_stats_streak ON gamification_stats(current_streak DESC);

-- Category mastery tracking
CREATE TABLE category_mastery (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    completed_count INTEGER NOT NULL DEFAULT 0,
    last_completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, category)
);

COMMENT ON TABLE category_mastery IS 'Tracks task completion counts per category for mastery achievements';

CREATE INDEX idx_category_mastery_user_id ON category_mastery(user_id);
CREATE INDEX idx_category_mastery_count ON category_mastery(completed_count DESC);

-- ============================================================
-- SYSTEM TABLES
-- ============================================================

-- Anonymous user cleanup audit log
CREATE TABLE anonymous_user_cleanups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    task_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_anonymous_cleanups_deleted_at ON anonymous_user_cleanups(deleted_at DESC);

-- Schema migrations tracking
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- ============================================================
-- FUNCTIONS
-- ============================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Full-text search vector update
CREATE OR REPLACE FUNCTION tasks_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.context, '')), 'C');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Count incomplete subtasks
CREATE OR REPLACE FUNCTION count_incomplete_subtasks(parent_id UUID)
RETURNS INTEGER AS $$
    SELECT COUNT(*)::INTEGER
    FROM tasks
    WHERE parent_task_id = $1
        AND task_type = 'subtask'
        AND status != 'done';
$$ LANGUAGE SQL STABLE;

-- Count incomplete blockers
CREATE OR REPLACE FUNCTION count_incomplete_blockers(p_task_id UUID)
RETURNS INTEGER AS $$
    SELECT COUNT(*)::INTEGER
    FROM task_dependencies td
    INNER JOIN tasks t ON t.id = td.blocked_by_id
    WHERE td.task_id = p_task_id
      AND t.status != 'done';
$$ LANGUAGE SQL STABLE;

-- Get blocker IDs for cycle detection
CREATE OR REPLACE FUNCTION get_blocker_ids(p_task_id UUID)
RETURNS UUID[] AS $$
    SELECT COALESCE(array_agg(blocked_by_id), ARRAY[]::UUID[])
    FROM task_dependencies
    WHERE task_id = p_task_id;
$$ LANGUAGE SQL STABLE;

-- ============================================================
-- TRIGGERS
-- ============================================================

-- Users
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Tasks
CREATE TRIGGER update_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER tasks_search_vector_trigger
    BEFORE INSERT OR UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION tasks_search_vector_update();

-- Task series
CREATE TRIGGER update_task_series_updated_at
    BEFORE UPDATE ON task_series
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- User preferences
CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Category preferences
CREATE TRIGGER update_category_preferences_updated_at
    BEFORE UPDATE ON category_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Task templates
CREATE TRIGGER update_task_templates_updated_at
    BEFORE UPDATE ON task_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Gamification stats
CREATE TRIGGER update_gamification_stats_updated_at
    BEFORE UPDATE ON gamification_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Category mastery
CREATE TRIGGER update_category_mastery_updated_at
    BEFORE UPDATE ON category_mastery
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================
-- Enable RLS to satisfy Supabase Security Advisor.
-- Since TaskFlow uses a Go backend (not direct PostgREST access),
-- no permissive policies are needed - the backend uses postgres superuser.

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_series ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE category_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_dependencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_achievements ENABLE ROW LEVEL SECURITY;
ALTER TABLE gamification_stats ENABLE ROW LEVEL SECURITY;
ALTER TABLE category_mastery ENABLE ROW LEVEL SECURITY;
ALTER TABLE anonymous_user_cleanups ENABLE ROW LEVEL SECURITY;
ALTER TABLE schema_migrations ENABLE ROW LEVEL SECURITY;

-- Revoke direct PostgREST access (belt and suspenders)
REVOKE ALL ON users FROM anon, authenticated;
REVOKE ALL ON tasks FROM anon, authenticated;
REVOKE ALL ON task_history FROM anon, authenticated;
REVOKE ALL ON task_series FROM anon, authenticated;
REVOKE ALL ON user_preferences FROM anon, authenticated;
REVOKE ALL ON category_preferences FROM anon, authenticated;
REVOKE ALL ON task_dependencies FROM anon, authenticated;
REVOKE ALL ON task_templates FROM anon, authenticated;
REVOKE ALL ON user_achievements FROM anon, authenticated;
REVOKE ALL ON gamification_stats FROM anon, authenticated;
REVOKE ALL ON category_mastery FROM anon, authenticated;
REVOKE ALL ON anonymous_user_cleanups FROM anon, authenticated;
REVOKE ALL ON schema_migrations FROM anon, authenticated;

-- ============================================================
-- INITIAL DATA
-- ============================================================

-- Record that this setup script was applied
INSERT INTO schema_migrations (version, applied_at) VALUES
    ('000001_initial_schema', NOW()),
    ('000002_update_user_priority_scale', NOW()),
    ('000003_add_composite_indexes', NOW()),
    ('000004_recurring_tasks', NOW()),
    ('000005_subtasks_support', NOW()),
    ('000006_task_dependencies', NOW()),
    ('000007_task_templates', NOW()),
    ('000008_gamification', NOW()),
    ('000009_anonymous_users', NOW()),
    ('000010_soft_delete', NOW()),
    ('000011_add_task_statuses', NOW()),
    ('000012_optimize_active_task_index', NOW()),
    ('000013_enable_rls', NOW());

-- ============================================================
-- SETUP COMPLETE!
-- ============================================================
-- Your TaskFlow database is now ready.
--
-- Next steps:
-- 1. Copy your database connection string from Settings > Database
-- 2. Set it as DATABASE_URL in backend/.env
-- 3. Start the backend: cd backend && go run cmd/server/main.go
-- 4. Start the frontend: cd frontend && npm run dev
-- ============================================================
