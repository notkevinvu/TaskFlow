-- Migration: Add gamification system (streaks, achievements, productivity scores)
-- Phase 5B.3: Gamification feature

-- Add timezone to user_preferences for streak calculation in user's local time
ALTER TABLE user_preferences
ADD COLUMN IF NOT EXISTS timezone VARCHAR(100) NOT NULL DEFAULT 'UTC';

-- Create achievement_type enum for type safety
CREATE TYPE achievement_type AS ENUM (
    -- Milestone achievements (task completion counts)
    'first_task',        -- Complete 1 task
    'milestone_10',      -- Complete 10 tasks
    'milestone_50',      -- Complete 50 tasks
    'milestone_100',     -- Complete 100 tasks

    -- Streak achievements (consecutive days)
    'streak_3',          -- 3-day streak
    'streak_7',          -- 7-day streak (Week Warrior)
    'streak_14',         -- 14-day streak (Fortnight Champion)
    'streak_30',         -- 30-day streak (Monthly Master)

    -- Category mastery (10 tasks in same category)
    'category_master',   -- Earned per category (metadata contains category name)

    -- Speed achievements
    'speed_demon',       -- Complete task within 24h of creation

    -- Consistency achievements
    'consistency_king'   -- Complete tasks 5+ days in a week
);

-- Create user_achievements table (earned badges)
CREATE TABLE user_achievements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_type achievement_type NOT NULL,
    earned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB,  -- e.g., {"category": "Bug Fix", "count": 10} for category_master

    -- Note: Unique constraint for achievements is handled by a functional index below
    -- This allows multiple category_master achievements (different categories)
    -- but prevents duplicate milestone/streak achievements
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for user_achievements
CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);
CREATE INDEX idx_user_achievements_earned_at ON user_achievements(earned_at DESC);
CREATE INDEX idx_user_achievements_type ON user_achievements(achievement_type);

-- Unique functional index: prevents duplicate achievements while allowing multiple category_master per category
-- Uses COALESCE to handle NULL metadata for non-category achievements
CREATE UNIQUE INDEX unique_achievement_per_user
    ON user_achievements(user_id, achievement_type, COALESCE((metadata->>'category')::text, ''));

-- Create gamification_stats table (cached computed stats for performance)
CREATE TABLE gamification_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    -- Streak data
    current_streak INTEGER NOT NULL DEFAULT 0,
    longest_streak INTEGER NOT NULL DEFAULT 0,
    last_completion_date DATE,  -- In user's timezone

    -- Computed totals
    total_completed INTEGER NOT NULL DEFAULT 0,

    -- Productivity score (0.00 to 100.00)
    productivity_score NUMERIC(5,2) NOT NULL DEFAULT 0.00,

    -- Component scores for productivity breakdown
    completion_rate NUMERIC(5,2) DEFAULT 0.00,      -- % of tasks completed (0-100)
    streak_score NUMERIC(5,2) DEFAULT 0.00,         -- Bonus from streak (0-100)
    on_time_percentage NUMERIC(5,2) DEFAULT 0.00,   -- % completed before due date (0-100)
    effort_mix_score NUMERIC(5,2) DEFAULT 0.00,     -- Balance of effort levels (0-100)

    -- Cache management
    last_computed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for leaderboard-style queries (future feature)
CREATE INDEX idx_gamification_stats_productivity ON gamification_stats(productivity_score DESC);
CREATE INDEX idx_gamification_stats_streak ON gamification_stats(current_streak DESC);

-- Auto-update trigger for gamification_stats
CREATE TRIGGER update_gamification_stats_updated_at
    BEFORE UPDATE ON gamification_stats
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create category_mastery table (track progress towards category achievements)
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

-- Create indexes for category mastery
CREATE INDEX idx_category_mastery_user_id ON category_mastery(user_id);
CREATE INDEX idx_category_mastery_count ON category_mastery(completed_count DESC);

-- Auto-update trigger for category_mastery
CREATE TRIGGER update_category_mastery_updated_at
    BEFORE UPDATE ON category_mastery
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE user_achievements IS 'Stores unlocked achievement badges for gamification';
COMMENT ON TABLE gamification_stats IS 'Cached productivity stats (streaks, scores) for fast dashboard loading';
COMMENT ON TABLE category_mastery IS 'Tracks task completion counts per category for mastery achievements';
COMMENT ON COLUMN gamification_stats.productivity_score IS 'Weighted score: 30% completion + 25% streak + 25% on-time + 20% effort mix';
COMMENT ON COLUMN user_preferences.timezone IS 'IANA timezone (e.g., America/New_York) for streak calculation';
