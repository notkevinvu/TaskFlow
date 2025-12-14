-- TaskFlow Database Schema for sqlc
-- This represents the current state after all migrations

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create task_status enum
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done');

-- Create task_effort enum
CREATE TYPE task_effort AS ENUM ('small', 'medium', 'large', 'xlarge');

-- Create task_history_event_type enum
CREATE TYPE task_history_event_type AS ENUM (
    'created',
    'updated',
    'bumped',
    'completed',
    'deleted',
    'status_changed'
);

-- Create recurrence_pattern enum
CREATE TYPE recurrence_pattern AS ENUM ('none', 'daily', 'weekly', 'monthly');

-- Create due_date_calculation enum
CREATE TYPE due_date_calculation AS ENUM ('from_original', 'from_completion');

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- Create task_series table to track recurring task series
CREATE TABLE task_series (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_task_id UUID NOT NULL,  -- First task in the series (FK added after tasks table)
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

-- Create tasks table
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
    related_people TEXT[], -- Array of strings
    priority_score INTEGER NOT NULL DEFAULT 50 CHECK (priority_score >= 0 AND priority_score <= 100),
    bump_count INTEGER NOT NULL DEFAULT 0,
    search_vector tsvector, -- For full-text search
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    -- Recurrence fields
    series_id UUID REFERENCES task_series(id) ON DELETE SET NULL,
    parent_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL
);

-- Add foreign key from task_series to tasks after tasks table exists
ALTER TABLE task_series ADD CONSTRAINT fk_task_series_original_task
    FOREIGN KEY (original_task_id) REFERENCES tasks(id) ON DELETE CASCADE;

-- Create indexes for tasks
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority_score ON tasks(priority_score DESC);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_category ON tasks(category);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);

-- Create GIN index for full-text search
CREATE INDEX idx_tasks_search_vector ON tasks USING GIN(search_vector);

-- Create indexes for recurrence fields
CREATE INDEX idx_tasks_series_id ON tasks(series_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);

-- Create task_history table for audit log
CREATE TABLE task_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event_type task_history_event_type NOT NULL,
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for task_history
CREATE INDEX idx_task_history_task_id ON task_history(task_id);
CREATE INDEX idx_task_history_user_id ON task_history(user_id);
CREATE INDEX idx_task_history_event_type ON task_history(event_type);
CREATE INDEX idx_task_history_created_at ON task_history(created_at DESC);

-- Create user_preferences table for storing user settings
CREATE TABLE user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_due_date_calculation due_date_calculation NOT NULL DEFAULT 'from_original',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

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

-- Create task_dependencies table (blocked-by relationships)
CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocked_by_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, blocked_by_id),
    CONSTRAINT no_self_dependency CHECK (task_id != blocked_by_id)
);

-- Index for efficient "what blocks this task?" queries
CREATE INDEX idx_task_dependencies_task ON task_dependencies(task_id);

-- Index for efficient "what does this task block?" queries
CREATE INDEX idx_task_dependencies_blocker ON task_dependencies(blocked_by_id);
