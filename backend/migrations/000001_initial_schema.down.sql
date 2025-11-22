-- Drop task_history table and related objects
DROP TABLE IF EXISTS task_history CASCADE;
DROP TYPE IF EXISTS task_history_event_type CASCADE;

-- Drop tasks table and related objects
DROP TRIGGER IF EXISTS tasks_search_vector_trigger ON tasks;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
DROP FUNCTION IF EXISTS tasks_search_vector_update();
DROP TABLE IF EXISTS tasks CASCADE;
DROP TYPE IF EXISTS task_effort CASCADE;
DROP TYPE IF EXISTS task_status CASCADE;

-- Drop users table and related objects
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
