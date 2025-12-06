-- Migration: Add task templates
-- User-specific templates for quickly creating tasks with pre-filled values

CREATE TABLE task_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Template metadata
    name VARCHAR(100) NOT NULL,

    -- Task field values (match CreateTaskDTO structure)
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    estimated_effort task_effort,
    user_priority INTEGER NOT NULL DEFAULT 5 CHECK (user_priority >= 1 AND user_priority <= 10),
    context VARCHAR(500),
    related_people TEXT[],

    -- Relative due date (days from creation)
    -- NULL = no due date, positive = days from now
    due_date_offset INTEGER CHECK (due_date_offset IS NULL OR (due_date_offset >= 0 AND due_date_offset <= 365)),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure unique template names per user
    CONSTRAINT unique_user_template_name UNIQUE(user_id, name)
);

-- Index for efficient user template listing
CREATE INDEX idx_task_templates_user_id ON task_templates(user_id);

-- Index for category-based filtering (optional future feature)
CREATE INDEX idx_task_templates_user_category ON task_templates(user_id, category)
    WHERE category IS NOT NULL;

-- Auto-update trigger for updated_at
CREATE TRIGGER update_task_templates_updated_at
    BEFORE UPDATE ON task_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Documentation
COMMENT ON TABLE task_templates IS 'User-defined templates for quickly creating tasks with pre-filled values';
COMMENT ON COLUMN task_templates.name IS 'User-facing template name (e.g., "Weekly Report", "Bug Fix")';
COMMENT ON COLUMN task_templates.due_date_offset IS 'Days from creation (NULL = no due date, 7 = due in 1 week)';
