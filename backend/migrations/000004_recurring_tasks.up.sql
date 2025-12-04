-- Create recurrence_pattern enum
CREATE TYPE recurrence_pattern AS ENUM ('none', 'daily', 'weekly', 'monthly');

-- Create due_date_calculation enum
CREATE TYPE due_date_calculation AS ENUM ('from_original', 'from_completion');

-- Create task_series table to track recurring task series
CREATE TABLE task_series (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_task_id UUID NOT NULL,  -- First task in the series
    pattern recurrence_pattern NOT NULL DEFAULT 'none',
    interval_value INTEGER NOT NULL DEFAULT 1 CHECK (interval_value > 0 AND interval_value <= 365),
    end_date TIMESTAMP WITH TIME ZONE,  -- Optional: when the series should stop
    due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for task_series
CREATE INDEX idx_task_series_user_id ON task_series(user_id);
CREATE INDEX idx_task_series_original_task_id ON task_series(original_task_id);
CREATE INDEX idx_task_series_is_active ON task_series(is_active) WHERE is_active = true;

-- Add series_id and parent_task_id to tasks table
ALTER TABLE tasks ADD COLUMN series_id UUID REFERENCES task_series(id) ON DELETE SET NULL;
ALTER TABLE tasks ADD COLUMN parent_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL;

-- Create index for series lookups
CREATE INDEX idx_tasks_series_id ON tasks(series_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);

-- Add trigger for updating updated_at on task_series
CREATE TRIGGER update_task_series_updated_at
    BEFORE UPDATE ON task_series
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create user_preferences table for storing user settings
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add trigger for updating updated_at on user_preferences
CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create category_preferences table for per-category settings
CREATE TABLE category_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, category)
);

-- Create index for category preferences
CREATE INDEX idx_category_preferences_user_id ON category_preferences(user_id);

-- Add trigger for updating updated_at on category_preferences
CREATE TRIGGER update_category_preferences_updated_at
    BEFORE UPDATE ON category_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Now add the foreign key constraint for original_task_id after tasks table has series_id
ALTER TABLE task_series ADD CONSTRAINT fk_task_series_original_task
    FOREIGN KEY (original_task_id) REFERENCES tasks(id) ON DELETE CASCADE;
