-- Update user_priority to use 1-10 scale instead of 0-100
-- Drop old constraint
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_user_priority_check;

-- Add new constraint for 1-10 range
ALTER TABLE tasks ADD CONSTRAINT tasks_user_priority_check CHECK (user_priority >= 1 AND user_priority <= 10);

-- Update default value from 50 to 5 (middle of new range)
ALTER TABLE tasks ALTER COLUMN user_priority SET DEFAULT 5;

-- Update existing tasks: scale 0-100 values to 1-10
-- Formula: GREATEST(1, LEAST(10, ROUND((user_priority / 10.0) + 0.5)))
UPDATE tasks SET user_priority = GREATEST(1, LEAST(10, ROUND((user_priority / 10.0) + 0.5)));
