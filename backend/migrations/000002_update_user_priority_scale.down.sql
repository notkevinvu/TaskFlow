-- Revert user_priority back to 0-100 scale
-- Drop 1-10 constraint
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_user_priority_check;

-- Update existing tasks: scale 1-10 values back to 0-100
-- Formula: (user_priority - 1) * 10 + 5 (maps to 5, 15, 25, ..., 95)
UPDATE tasks SET user_priority = (user_priority - 1) * 10 + 5;

-- Add back 0-100 constraint
ALTER TABLE tasks ADD CONSTRAINT tasks_user_priority_check CHECK (user_priority >= 0 AND user_priority <= 100);

-- Revert default back to 50
ALTER TABLE tasks ALTER COLUMN user_priority SET DEFAULT 50;
