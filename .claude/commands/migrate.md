---
description: Check for and apply database migrations to Supabase
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Database Migration Manager

This command helps manage database migrations for the Supabase PostgreSQL database.

### Step 1: Detect Migration Files

1. List all migration files in `backend/migrations/`:
   ```bash
   ls -la backend/migrations/*.up.sql
   ```

2. Parse migration files to extract:
   - Migration number (e.g., `000005` from `000005_subtasks_support.up.sql`)
   - Migration name (e.g., `subtasks_support`)
   - File path for reading content

### Step 2: Check Current Database State

Since we don't have direct database access, check for a migration tracking file:
- Location: `backend/migrations/.applied`
- Format: One migration name per line (e.g., `000004_recurring_tasks`)

If the file doesn't exist, ask the user which migrations are already applied:

```markdown
## Migration Status Unknown

No `.applied` tracking file found. Please check your Supabase database for these tables/columns:

| Migration | Check For |
|-----------|-----------|
| 000001_* | `tasks` table exists |
| 000002_* | Check specific schema change |
| 000003_* | Check specific schema change |
| 000004_* | `task_series` table, `series_id` column |
| 000005_* | `task_type` column on `tasks` table |

Which migrations are already applied? Enter the last applied number (e.g., "4" or "none"):
```

### Step 3: Identify Pending Migrations

Compare migration files against applied migrations to find pending ones.

Display status table:

```markdown
## Migration Status

| # | Migration | Status |
|---|-----------|--------|
| 000001 | initial_schema | ✅ Applied |
| 000002 | add_search | ✅ Applied |
| 000003 | add_analytics | ✅ Applied |
| 000004 | recurring_tasks | ✅ Applied |
| 000005 | subtasks_support | ⏳ Pending |
```

### Step 4: Apply Pending Migrations

For each pending migration:

1. **Read the migration file**:
   ```bash
   Read backend/migrations/NNNNNN_name.up.sql
   ```

2. **Display migration content for user**:
   ```markdown
   ## Apply Migration: NNNNNN_name

   **File:** `backend/migrations/NNNNNN_name.up.sql`

   ### SQL to Execute

   Copy the following SQL and run it in your **Supabase SQL Editor**:

   \`\`\`sql
   [MIGRATION SQL CONTENT HERE]
   \`\`\`

   ### Instructions

   1. Go to your **Supabase Dashboard** → **SQL Editor**
   2. Paste the SQL above
   3. Click **Run**
   4. Verify no errors occurred

   Did the migration apply successfully? (yes/no)
   ```

3. **Wait for user confirmation** before proceeding to the next migration

4. **Update tracking file** after successful confirmation:
   - Add the migration to `backend/migrations/.applied`
   - This file should be gitignored (local state only)

### Step 5: Verify and Report

After all migrations are applied:

```markdown
## Migration Complete ✅

All pending migrations have been applied:

| Migration | Status |
|-----------|--------|
| 000005_subtasks_support | ✅ Applied |

**Database Version:** 5 (subtasks_support)

**Next Steps:**
- Restart your backend server to pick up schema changes
- Test the new functionality
```

### Arguments

- `$ARGUMENTS` can be:
  - Empty: Check status and apply pending migrations
  - `status`: Only show migration status, don't apply
  - `reset`: Clear the `.applied` tracking file
  - `apply N`: Apply migration number N specifically

### Error Handling

- If migration fails, provide rollback instructions:
  - Show the corresponding `.down.sql` file content
  - Warn about potential data loss
  - Ask user to confirm before showing rollback SQL

### Notes

- Migrations are applied in numerical order
- Each migration must succeed before the next is attempted
- The `.applied` file is local-only and should be in `.gitignore`
- Always restart the backend after applying migrations
