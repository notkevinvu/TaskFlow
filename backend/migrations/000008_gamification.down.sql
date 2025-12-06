-- Rollback: Remove gamification system

-- Drop tables in reverse order (respect foreign key constraints)
DROP TABLE IF EXISTS category_mastery;
DROP TABLE IF EXISTS gamification_stats;
DROP TABLE IF EXISTS user_achievements;

-- Drop enum type
DROP TYPE IF EXISTS achievement_type;

-- Remove timezone column from user_preferences
ALTER TABLE user_preferences DROP COLUMN IF EXISTS timezone;
