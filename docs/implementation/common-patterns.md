# Common Patterns & Code Snippets
## Intelligent Task Prioritization System

Frequently used patterns for the TaskFlow application.

---

## Backend Patterns (Go)

### 1. Task CRUD Handler with Priority Calculation

```go
type TaskHandler struct {
    service ports.TaskService
    logger  *slog.Logger
}

// getUserID extracts authenticated user from context
func getUserID(c *gin.Context) (uuid.UUID, error) {
    userInterface, exists := c.Get("user")
    if !exists {
        return uuid.Nil, errors.New("user not found in context")
    }
    user := userInterface.(*domain.User)
    return user.ID, nil
}

// List returns priority-sorted tasks
func (h *TaskHandler) List(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    tasks, err := h.service.List(c.Request.Context(), userID, int32(limit), int32(offset))
    if err != nil {
        h.logger.Error("failed to list tasks", "error", err)
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }
    c.JSON(200, tasks)
}

// Create task with automatic priority calculation
func (h *TaskHandler) Create(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    var req domain.CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    task, err := h.service.Create(c.Request.Context(), userID, &req)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, task)
}

// Bump task (delay) and recalculate priority
func (h *TaskHandler) Bump(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid task ID"})
        return
    }

    var req domain.BumpTaskRequest
    _ = c.ShouldBindJSON(&req) // Reason is optional

    task, err := h.service.Bump(c.Request.Context(), taskID, userID, &req)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "message": "task bumped",
        "task":    task,
    })
}

// Complete task
func (h *TaskHandler) Complete(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(400, gin.H{"error": "invalid task ID"})
        return
    }

    task, err := h.service.Complete(c.Request.Context(), taskID, userID)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, task)
}

// GetAtRisk returns tasks bumped 3+ times
func (h *TaskHandler) GetAtRisk(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    tasks, err := h.service.GetAtRisk(c.Request.Context(), userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(200, gin.H{
        "tasks": tasks,
        "count": len(tasks),
    })
}
```

**Register routes:**

```go
func setupTaskRoutes(router *gin.RouterGroup, handler *TaskHandler) {
    tasks := router.Group("/tasks")
    {
        tasks.GET("", handler.List)
        tasks.POST("", handler.Create)
        tasks.GET("/at-risk", handler.GetAtRisk)
        tasks.GET("/:id", handler.Get)
        tasks.PUT("/:id", handler.Update)
        tasks.POST("/:id/bump", handler.Bump)
        tasks.POST("/:id/complete", handler.Complete)
        tasks.DELETE("/:id", handler.Delete)
    }
}
```

---

### 2. Priority Calculation Service Pattern

```go
type priorityService struct{}

func NewPriorityService() ports.PriorityService {
    return &priorityService{}
}

// CalculatePriority computes multi-factor priority score
// Formula: (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
func (s *priorityService) CalculatePriority(task *domain.Task) float64 {
    userPriority := float64(task.UserPriority)
    timeDecay := s.calculateTimeDecay(task)
    deadlineUrgency := s.calculateDeadlineUrgency(task)
    bumpPenalty := float64(task.BumpCount) * 10.0
    if bumpPenalty > 50 {
        bumpPenalty = 50
    }

    // Effort boost multiplier
    effortBoost := 1.0
    if task.EstimatedEffort != nil {
        switch *task.EstimatedEffort {
        case domain.TaskEffortSmall:
            effortBoost = 1.3 // 30% boost for small tasks
        case domain.TaskEffortMedium:
            effortBoost = 1.15 // 15% boost
        case domain.TaskEffortLarge:
            effortBoost = 1.0
        case domain.TaskEffortXLarge:
            effortBoost = 0.95 // Slight penalty for very large tasks
        }
    }

    score := (
        userPriority*0.4 +
        timeDecay*0.3 +
        deadlineUrgency*0.2 +
        bumpPenalty*0.1,
    ) * effortBoost

    if score > 100 {
        return 100
    }
    return math.Round(score*100) / 100
}

// Time decay: increases linearly over 30 days
func (s *priorityService) calculateTimeDecay(task *domain.Task) float64 {
    age := time.Since(task.CreatedAt)
    daysSinceCreation := age.Hours() / 24.0
    decay := (daysSinceCreation / 30.0) * 100.0

    if decay > 100 {
        return 100
    }
    return decay
}

// Deadline urgency: exponential increase in final 3 days
func (s *priorityService) calculateDeadlineUrgency(task *domain.Task) float64 {
    if task.DueDate == nil {
        return 0
    }

    daysUntilDue := time.Until(*task.DueDate).Hours() / 24.0

    if daysUntilDue < 0 {
        return 100 // Overdue
    }
    if daysUntilDue <= 3 {
        return 100 - (daysUntilDue/3.0)*50 // Exponential urgency
    }
    if daysUntilDue <= 7 {
        return ((7 - daysUntilDue) / 4.0) * 50
    }
    return 0
}
```

---

### 3. Pagination Pattern for Tasks

```go
type PaginationParams struct {
    Limit  int `form:"limit" binding:"min=1,max=100"`
    Offset int `form:"offset" binding:"min=0"`
}

type TaskListResponse struct {
    Tasks      []*domain.Task `json:"tasks"`
    TotalCount int            `json:"total_count"`
}

func (h *TaskHandler) List(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    var params PaginationParams
    params.Limit = 50
    params.Offset = 0

    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    response, err := h.service.List(
        c.Request.Context(),
        userID,
        int32(params.Limit),
        int32(params.Offset),
    )
    if err != nil {
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(200, response)
}
```

**sqlc query (priority-sorted):**

```sql
-- name: ListTasks :many
SELECT * FROM tasks
WHERE user_id = $1 AND deleted_at IS NULL AND status != 'done'
ORDER BY priority_score DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetAtRiskTasks :many
SELECT * FROM tasks
WHERE user_id = $1 AND deleted_at IS NULL AND status != 'done'
  AND bump_count >= 3
ORDER BY priority_score DESC;
```

---

### 3. Error Handling Pattern

```go
// Custom errors
var (
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrBadRequest    = errors.New("bad request")
    ErrInternal      = errors.New("internal server error")
)

// Error response
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
    Code    string `json:"code,omitempty"`
}

// Middleware
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) == 0 {
            return
        }

        err := c.Errors.Last().Err
        var statusCode int
        var message string

        switch {
        case errors.Is(err, ErrNotFound):
            statusCode = http.StatusNotFound
            message = "Resource not found"
        case errors.Is(err, ErrUnauthorized):
            statusCode = http.StatusUnauthorized
            message = "Unauthorized access"
        case errors.Is(err, ErrBadRequest):
            statusCode = http.StatusBadRequest
            message = err.Error()
        default:
            statusCode = http.StatusInternalServerError
            message = "Internal server error"
            logger.Error("internal error", "error", err)
        }

        c.JSON(statusCode, ErrorResponse{
            Error:   http.StatusText(statusCode),
            Message: message,
        })
    }
}
```

---

### 4. Database Transaction Pattern

```go
func (s *userService) CreateWithProfile(ctx context.Context, req *CreateUserRequest) error {
    // Begin transaction
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    queries := s.queries.WithTx(tx)

    // Create user
    user, err := queries.CreateUser(ctx, CreateUserParams{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return err
    }

    // Create profile
    _, err = queries.CreateProfile(ctx, CreateProfileParams{
        UserID: user.ID,
        Bio:    req.Bio,
    })
    if err != nil {
        return err
    }

    // Commit transaction
    return tx.Commit(ctx)
}
```

---

## Frontend Patterns (Next.js + TypeScript)

### 1. React Query Pattern for Tasks

```typescript
// hooks/useTasks.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { taskAPI, CreateTaskDTO } from '@/lib/api';
import { toast } from 'sonner';

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      const response = await taskAPI.list({ limit: 100, offset: 0 });
      return response.data;
    },
  });
}

export function useTask(id: string) {
  return useQuery({
    queryKey: ['tasks', id],
    queryFn: async () => {
      const response = await taskAPI.getById(id);
      return response.data;
    },
    enabled: !!id,
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTaskDTO) => taskAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Task created with priority calculated!');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to create task');
    },
  });
}

export function useBumpTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      taskAPI.bump(id, reason),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.info(`Task delayed. New priority: ${response.data.task.priority_score}`);
    },
  });
}

export function useCompleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.complete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Task completed!');
    },
  });
}

export function useAtRiskTasks() {
  return useQuery({
    queryKey: ['tasks', 'at-risk'],
    queryFn: async () => {
      const response = await taskAPI.getAtRisk();
      return response.data;
    },
  });
}
```

**Usage in component:**

```typescript
'use client';

import { useTasks, useBumpTask, useCompleteTask } from '@/hooks/useTasks';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';

export function TaskList() {
  const { data: tasksData, isLoading } = useTasks();
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();

  if (isLoading) return <div>Loading tasks...</div>;

  const tasks = tasksData?.tasks || [];

  return (
    <div className="space-y-4">
      {tasks.map(task => (
        <div key={task.id} className="flex items-center justify-between p-4 border rounded">
          <div>
            <h4 className="font-semibold">{task.title}</h4>
            <div className="flex gap-2 mt-2">
              <Badge variant={task.priority_score >= 90 ? "destructive" : "default"}>
                Priority: {Math.round(task.priority_score)}
              </Badge>
              {task.bump_count > 0 && (
                <Badge variant="outline">Bumped {task.bump_count}x</Badge>
              )}
            </div>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => bumpTask.mutate({ id: task.id })}
            >
              Bump
            </Button>
            <Button
              size="sm"
              onClick={() => completeTask.mutate(task.id)}
            >
              Complete
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
```

---

### 2. Task Creation Form Pattern

```typescript
'use client';

import { useState } from 'react';
import { z } from 'zod';
import { useCreateTask } from '@/hooks/useTasks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

const taskSchema = z.object({
  title: z.string().min(1, 'Title is required').max(200, 'Title too long'),
  description: z.string().optional(),
  user_priority: z.number().min(0).max(100).optional(),
  due_date: z.string().optional(),
  estimated_effort: z.enum(['small', 'medium', 'large', 'xlarge']).optional(),
  category: z.string().optional(),
  context: z.string().optional(),
});

type TaskFormData = z.infer<typeof taskSchema>;

export function TaskForm({ onSuccess }: { onSuccess?: () => void }) {
  const [formData, setFormData] = useState<TaskFormData>({
    title: '',
    user_priority: 50,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const createTask = useCreateTask();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate
    const result = taskSchema.safeParse(formData);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.errors.forEach((err) => {
        if (err.path[0]) {
          fieldErrors[err.path[0].toString()] = err.message;
        }
      });
      setErrors(fieldErrors);
      return;
    }

    // Submit (priority will be auto-calculated by backend)
    try {
      await createTask.mutateAsync(result.data);
      onSuccess?.();
    } catch (error) {
      // Error handled by mutation
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <Label htmlFor="title">Title *</Label>
        <Input
          id="title"
          value={formData.title}
          onChange={(e) => setFormData({ ...formData, title: e.target.value })}
          placeholder="Review design doc for auth flow"
        />
        {errors.title && <p className="text-red-500 text-sm">{errors.title}</p>}
      </div>

      <div>
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          value={formData.description}
          onChange={(e) => setFormData({ ...formData, description: e.target.value })}
          placeholder="Additional details..."
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="priority">User Priority (0-100)</Label>
          <Input
            id="priority"
            type="number"
            min="0"
            max="100"
            value={formData.user_priority}
            onChange={(e) => setFormData({ ...formData, user_priority: Number(e.target.value) })}
          />
        </div>

        <div>
          <Label htmlFor="effort">Estimated Effort</Label>
          <Select
            value={formData.estimated_effort}
            onValueChange={(value: any) => setFormData({ ...formData, estimated_effort: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder="Select size" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="small">Small (&lt; 1h)</SelectItem>
              <SelectItem value="medium">Medium (1-4h)</SelectItem>
              <SelectItem value="large">Large (4-8h)</SelectItem>
              <SelectItem value="xlarge">XLarge (&gt; 8h)</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="category">Category</Label>
          <Input
            id="category"
            value={formData.category}
            onChange={(e) => setFormData({ ...formData, category: e.target.value })}
            placeholder="Code Review"
          />
        </div>

        <div>
          <Label htmlFor="due_date">Due Date</Label>
          <Input
            id="due_date"
            type="date"
            value={formData.due_date}
            onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
          />
        </div>
      </div>

      <div>
        <Label htmlFor="context">Context</Label>
        <Textarea
          id="context"
          value={formData.context}
          onChange={(e) => setFormData({ ...formData, context: e.target.value })}
          placeholder="From meeting with Alice - needs feedback by Friday"
        />
      </div>

      <Button type="submit" disabled={createTask.isPending}>
        {createTask.isPending ? 'Creating...' : 'Create Task'}
      </Button>
    </form>
  );
}
```

---

### 3. Priority-Sorted Task Table

```typescript
'use client';

import { useTasks, useBumpTask, useCompleteTask } from '@/hooks/useTasks';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export function TasksTable() {
  const { data: tasksData, isLoading } = useTasks();
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();

  if (isLoading) return <div>Loading tasks...</div>;

  const tasks = tasksData?.tasks || [];

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Priority</TableHead>
          <TableHead>Task</TableHead>
          <TableHead>Category</TableHead>
          <TableHead>Due Date</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {tasks.map((task) => (
          <TableRow key={task.id}>
            <TableCell>
              <Badge
                variant={
                  task.priority_score >= 90
                    ? "destructive"
                    : task.priority_score >= 75
                    ? "default"
                    : "secondary"
                }
              >
                {Math.round(task.priority_score)}
              </Badge>
            </TableCell>
            <TableCell>
              <div>
                <div className="font-medium">{task.title}</div>
                {task.context && (
                  <div className="text-sm text-muted-foreground">{task.context}</div>
                )}
              </div>
            </TableCell>
            <TableCell>{task.category || '-'}</TableCell>
            <TableCell>
              {task.due_date
                ? new Date(task.due_date).toLocaleDateString()
                : '-'}
            </TableCell>
            <TableCell>
              <div className="flex gap-1">
                <Badge variant="outline">{task.status}</Badge>
                {task.bump_count > 0 && (
                  <Badge variant="outline" className="text-yellow-600">
                    ⚠️ {task.bump_count}
                  </Badge>
                )}
              </div>
            </TableCell>
            <TableCell>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => bumpTask.mutate({ id: task.id })}
                  disabled={bumpTask.isPending}
                >
                  Bump
                </Button>
                <Button
                  size="sm"
                  onClick={() => completeTask.mutate(task.id)}
                  disabled={completeTask.isPending}
                >
                  Complete
                </Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
```

---

### 4. Protected Route Pattern

```typescript
// middleware.ts (Next.js 15)
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  const token = request.cookies.get('token')?.value;

  if (!token && request.nextUrl.pathname.startsWith('/dashboard')) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: '/dashboard/:path*',
};
```

**Or with React hook:**

```typescript
// hooks/useRequireAuth.ts
'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from './useAuth';

export function useRequireAuth() {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !user) {
      router.push('/login');
    }
  }, [user, isLoading, router]);

  return { user, isLoading };
}
```

**Usage:**

```typescript
'use client';

export default function DashboardPage() {
  const { user, isLoading } = useRequireAuth();

  if (isLoading) return <div>Loading...</div>;
  if (!user) return null; // Redirecting...

  return <div>Welcome, {user.name}!</div>;
}
```

---

### 5. Loading States Pattern

```typescript
'use client';

import { useUsers } from '@/hooks/useUsers';
import { Skeleton } from '@/components/ui/skeleton';

export function UserList() {
  const { data: users, isLoading, error } = useUsers();

  if (isLoading) {
    return (
      <div className="space-y-2">
        {[...Array(5)].map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-red-500">
        Error loading users. Please try again.
      </div>
    );
  }

  if (!users || users.length === 0) {
    return <div className="text-muted-foreground">No users found.</div>;
  }

  return (
    <div>
      {users.map(user => (
        <div key={user.id}>{user.name}</div>
      ))}
    </div>
  );
}
```

---

## Database Patterns (SQL)

### 1. Full-Text Search on Tasks

```sql
-- Add tsvector column to tasks table
ALTER TABLE tasks ADD COLUMN search_vector tsvector;

-- Create GIN index for fast full-text search
CREATE INDEX idx_tasks_search ON tasks USING GIN(search_vector);

-- Auto-update search vector on insert/update
CREATE FUNCTION tasks_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.context, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tasks_search_update
BEFORE INSERT OR UPDATE ON tasks
FOR EACH ROW EXECUTE FUNCTION tasks_search_trigger();

-- Search tasks by keyword
SELECT
  id,
  title,
  priority_score,
  ts_rank(search_vector, query) AS rank
FROM tasks,
     to_tsquery('english', 'design & review') query
WHERE search_vector @@ query
  AND user_id = $1
  AND deleted_at IS NULL
ORDER BY rank DESC, priority_score DESC;
```

**sqlc query for search:**

```sql
-- name: SearchTasks :many
SELECT
  t.*,
  ts_rank(t.search_vector, to_tsquery('english', $2)) AS rank
FROM tasks t
WHERE t.user_id = $1
  AND t.search_vector @@ to_tsquery('english', $2)
  AND t.deleted_at IS NULL
ORDER BY rank DESC, t.priority_score DESC
LIMIT $3;
```

---

### 2. Soft Delete Pattern (Tasks)

```sql
-- Soft delete task
UPDATE tasks SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- Query only active (non-deleted) tasks
SELECT * FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status != 'done'
ORDER BY priority_score DESC;

-- Restore soft-deleted task
UPDATE tasks SET deleted_at = NULL, updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- Hard delete (cleanup old tasks)
DELETE FROM tasks
WHERE deleted_at < NOW() - INTERVAL '90 days';
```

---

### 3. Priority Recalculation Query

```sql
-- Update all task priorities (run periodically as background job)
UPDATE tasks
SET
  priority_score = calculate_priority(
    user_priority,
    created_at,
    due_date,
    bump_count,
    estimated_effort
  ),
  updated_at = NOW()
WHERE deleted_at IS NULL
  AND status != 'done'
  AND priority_last_calculated_at < NOW() - INTERVAL '6 hours';

-- Get tasks needing priority recalculation
SELECT * FROM tasks
WHERE deleted_at IS NULL
  AND status != 'done'
  AND priority_last_calculated_at < NOW() - INTERVAL '6 hours'
ORDER BY priority_last_calculated_at ASC
LIMIT 100;
```

---

These patterns cover all core TaskFlow use cases!
