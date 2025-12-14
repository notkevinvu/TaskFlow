# Module 04: Frontend Architecture

## Learning Objectives

By the end of this module, you will:
- Understand the separation of server state (React Query) and client state (Zustand)
- Master the optimistic updates pattern for instant UI feedback
- Learn the query key factory pattern for cache management
- Implement error handling utilities

---

## State Management Philosophy

TaskFlow separates state into two categories:

| Type | Tool | Examples |
|------|------|----------|
| **Server State** | React Query | Tasks, analytics, user data |
| **Client State** | Zustand | Auth token, theme, UI state |

```
┌─────────────────────────────────────────────────────────────┐
│                      Component Tree                          │
├─────────────────────────┬───────────────────────────────────┤
│     React Query         │         Zustand                   │
│   (Server State)        │     (Client State)                │
├─────────────────────────┼───────────────────────────────────┤
│ • Tasks from API        │ • Auth token in localStorage      │
│ • Analytics data        │ • Sidebar open/closed             │
│ • User profile          │ • Theme preference                │
│ • Gamification stats    │ • Selected task ID                │
│                         │ • Modal visibility                │
├─────────────────────────┼───────────────────────────────────┤
│ ✅ Automatic caching    │ ✅ Synchronous updates            │
│ ✅ Background refetch   │ ✅ Persists in localStorage       │
│ ✅ Stale-while-revalidate│ ✅ Zero network latency          │
│ ✅ Optimistic updates   │ ✅ No server round-trip           │
└─────────────────────────┴───────────────────────────────────┘
```

---

## Query Key Factory Pattern

Query keys are **cache identifiers**. A factory centralizes their definition:

```typescript
// frontend/lib/queryKeys.ts

export const taskKeys = {
  // Base key for all task queries
  all: ['tasks'] as const,

  // List queries (with optional filters)
  lists: () => [...taskKeys.all, 'list'] as const,
  list: (filters?: TaskFilters) => [...taskKeys.lists(), filters] as const,

  // Detail queries (single task)
  detail: (id: string) => [...taskKeys.all, 'detail', id] as const,

  // Special queries
  atRisk: () => [...taskKeys.all, 'at-risk'] as const,
  calendar: (params: CalendarParams) => [...taskKeys.all, 'calendar', params] as const,
  subtasks: (parentId: string) => [...taskKeys.all, 'subtasks', parentId] as const,
};

export const analyticsKeys = {
  all: ['analytics'] as const,
  summary: () => [...analyticsKeys.all, 'summary'] as const,
  heatmap: (daysBack: number) => [...analyticsKeys.all, 'heatmap', daysBack] as const,
  trends: (daysBack: number) => [...analyticsKeys.all, 'trends', daysBack] as const,
};

export const gamificationKeys = {
  all: ['gamification'] as const,
  stats: () => [...gamificationKeys.all, 'stats'] as const,
  achievements: () => [...gamificationKeys.all, 'achievements'] as const,
};
```

### Why This Pattern?

```typescript
// ❌ Bad: Hardcoded strings everywhere
useQuery({ queryKey: ['tasks', 'list', filters] })
queryClient.invalidateQueries({ queryKey: ['tasks'] })

// ✅ Good: Centralized, type-safe
useQuery({ queryKey: taskKeys.list(filters) })
queryClient.invalidateQueries({ queryKey: taskKeys.lists() })
```

Benefits:
1. **Type safety** - TypeScript catches typos
2. **Refactoring** - Change key structure in one place
3. **Targeted invalidation** - Invalidate specific query subsets

### Hierarchical Invalidation

```typescript
// Invalidate ALL task queries
queryClient.invalidateQueries({ queryKey: taskKeys.all })
// Matches: ['tasks', 'list'], ['tasks', 'detail', 'abc'], ['tasks', 'at-risk']

// Invalidate only list queries
queryClient.invalidateQueries({ queryKey: taskKeys.lists() })
// Matches: ['tasks', 'list'], ['tasks', 'list', {status: 'todo'}]
// Does NOT match: ['tasks', 'detail', 'abc']

// Invalidate specific task detail
queryClient.invalidateQueries({ queryKey: taskKeys.detail('abc') })
// Matches only: ['tasks', 'detail', 'abc']
```

---

## Basic Query Usage

### Fetching Tasks

```typescript
// frontend/hooks/useTasks.ts

import { useQuery } from '@tanstack/react-query';
import { taskAPI } from '@/lib/api';
import { taskKeys } from '@/lib/queryKeys';

export function useTasks(filters?: TaskFilters) {
  return useQuery({
    queryKey: taskKeys.list(filters),
    queryFn: async () => {
      const response = await taskAPI.list({
        limit: 100,
        offset: 0,
        ...filters,
      });
      return response.data;
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
}
```

### Using in Components

```tsx
// frontend/app/(dashboard)/dashboard/page.tsx

export default function DashboardPage() {
  const { data, isLoading, error } = useTasks({ status: 'todo' });

  if (isLoading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;

  return (
    <div>
      {data?.tasks.map((task) => (
        <TaskCard key={task.id} task={task} />
      ))}
    </div>
  );
}
```

---

## Optimistic Updates

The key to a snappy UI: **update immediately, rollback on error**.

### Pattern Overview

```
User clicks "Complete Task"
        │
        ▼
┌───────────────────────┐
│ 1. Save current state │ ← Snapshot for rollback
└───────────────────────┘
        │
        ▼
┌───────────────────────┐
│ 2. Update cache       │ ← User sees instant feedback
└───────────────────────┘
        │
        ▼
┌───────────────────────┐
│ 3. Send API request   │ ← Happens in background
└───────────────────────┘
        │
    ┌───┴───┐
    ▼       ▼
Success   Error
    │       │
    ▼       ▼
Invalidate Rollback
(get real  (restore
 data)     snapshot)
```

### Complete Implementation

```typescript
// frontend/hooks/useTasks.ts

export function useCompleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    // The actual API call
    mutationFn: (taskId: string) => taskAPI.complete(taskId),

    // BEFORE the API call: optimistic update
    onMutate: async (taskId: string) => {
      // 1. Cancel in-flight queries (prevent overwriting our optimistic update)
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      // 2. Snapshot current state for rollback
      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });

      // 3. Optimistically update the cache
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: taskKeys.lists() },
        (old) => {
          if (!old?.tasks) return old;
          return {
            ...old,
            tasks: old.tasks.map((task) =>
              task.id === taskId
                ? {
                    ...task,
                    status: 'done',
                    completed_at: new Date().toISOString(),
                  }
                : task
            ),
          };
        }
      );

      // 4. Return context for potential rollback
      return { previousLists };
    },

    // SUCCESS: Invalidate to get real data from server
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
      toast.success('Task completed!');
    },

    // ERROR: Rollback to snapshot
    onError: (err: unknown, _taskId, context) => {
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      toast.error(getApiErrorMessage(err, 'Failed to complete task'));
    },
  });
}
```

### Creating Tasks with Optimistic Update

```typescript
export function useCreateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTaskDTO) => taskAPI.create(data),

    onMutate: async (newTask) => {
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });

      // Create optimistic task with temporary ID
      const tempId = `temp-${crypto.randomUUID()}`;
      const optimisticTask: Task = {
        id: tempId,
        title: newTask.title,
        description: newTask.description,
        category: newTask.category,
        due_date: newTask.due_date,
        estimated_effort: newTask.estimated_effort,
        priority_score: 50, // Placeholder - backend calculates real priority
        status: 'todo',
        bump_count: 0,
        user_priority: newTask.user_priority || 5,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        user_id: '',
        task_type: 'regular',
      };

      // Add to beginning of list (new tasks appear first)
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: taskKeys.lists() },
        (old) => {
          if (!old?.tasks) return old;
          return {
            ...old,
            tasks: [optimisticTask, ...old.tasks],
            total_count: old.total_count + 1,
          };
        }
      );

      return { previousLists, optimisticTask };
    },

    onSuccess: () => {
      // Invalidate to replace temp ID with real ID and get real priority
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
      toast.success('Task created!');
    },

    onError: (err: unknown, _variables, context) => {
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      toast.error(getApiErrorMessage(err, 'Failed to create task'));
    },
  });
}
```

---

## Error Handling

### API Error Utility

```typescript
// frontend/lib/api.ts

import axios, { AxiosError, isAxiosError } from 'axios';

export function getApiErrorMessage(
  err: unknown,
  fallback: string,
  logContext?: string
): string {
  // Server returned an error message
  if (isAxiosError(err) && err.response?.data?.error) {
    return err.response.data.error;
  }

  // Network error (no response)
  if (isAxiosError(err) && !err.response) {
    if (err.code === 'ERR_NETWORK') {
      return 'Network error. Please check your connection.';
    }
    return 'Unable to connect to server.';
  }

  // HTTP status code errors
  if (isAxiosError(err) && err.response) {
    const status = err.response.status;
    if (status === 401) return 'Session expired. Please log in again.';
    if (status === 403) return 'You do not have permission.';
    if (status === 404) return 'The requested resource was not found.';
    if (status === 429) return 'Too many requests. Please wait.';
  }

  // Log unexpected errors
  if (logContext) {
    console.error(`[${logContext}]`, err);
  }

  return fallback;
}

export function isAuthError(err: unknown): boolean {
  return isAxiosError(err) && (err.response?.status === 401 || err.response?.status === 403);
}

export function isNetworkError(err: unknown): boolean {
  return isAxiosError(err) && !err.response;
}
```

### Usage in Mutations

```typescript
onError: (err: unknown) => {
  const message = getApiErrorMessage(err, 'Failed to save task', 'TaskUpdate');
  toast.error(message);

  // Special handling for auth errors
  if (isAuthError(err)) {
    router.push('/login');
  }
}
```

---

## API Client Setup

### Axios Instance with Interceptors

```typescript
// frontend/lib/api.ts

import axios from 'axios';

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  timeout: 10000,
});

// Request interceptor: Add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: Handle 401
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### API Methods

```typescript
// frontend/lib/api.ts

export const taskAPI = {
  list: (params: TaskListParams) =>
    api.get<TaskListResponse>('/api/v1/tasks', { params }),

  getById: (id: string) =>
    api.get<Task>(`/api/v1/tasks/${id}`),

  create: (data: CreateTaskDTO) =>
    api.post<Task>('/api/v1/tasks', data),

  update: (id: string, data: Partial<UpdateTaskDTO>) =>
    api.put<Task>(`/api/v1/tasks/${id}`, data),

  delete: (id: string) =>
    api.delete(`/api/v1/tasks/${id}`),

  complete: (id: string) =>
    api.post<Task>(`/api/v1/tasks/${id}/complete`),

  bump: (id: string) =>
    api.post<Task>(`/api/v1/tasks/${id}/bump`),
};
```

---

## Stale Time Configuration

Different queries have different freshness needs:

```typescript
// Tasks change frequently
export function useTasks() {
  return useQuery({
    queryKey: taskKeys.list(),
    queryFn: fetchTasks,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
}

// Calendar data changes less often
export function useCalendarTasks(params: CalendarParams) {
  return useQuery({
    queryKey: taskKeys.calendar(params),
    queryFn: () => fetchCalendarTasks(params),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Analytics can be stale longer
export function useAnalytics() {
  return useQuery({
    queryKey: analyticsKeys.summary(),
    queryFn: fetchAnalytics,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}
```

---

## React Query Provider Setup

```typescript
// frontend/app/providers.tsx

'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 2 * 60 * 1000, // 2 minutes default
            retry: 1,
            refetchOnWindowFocus: false,
          },
          mutations: {
            retry: 0,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

---

## Exercises

### 🔰 Beginner: Add a Query

1. Create a new hook `useArchivedTasks()` that fetches deleted tasks
2. Use the query key factory pattern
3. Set an appropriate stale time

### 🎯 Intermediate: Implement Optimistic Delete

1. Implement `useDeleteTask()` with optimistic update
2. Remove the task from cache immediately
3. Rollback if the API call fails

### 🚀 Advanced: Infinite Scroll

1. Convert `useTasks()` to use `useInfiniteQuery`
2. Implement "Load More" functionality
3. Handle cache invalidation for paginated data

---

## Reflection Questions

1. **Why not put everything in Redux?** What problems does separating server and client state solve?

2. **Why invalidate instead of updating with response data?** When would you use response data directly?

3. **What happens if two mutations run simultaneously?** How does React Query handle race conditions?

4. **Why snapshot ALL list queries, not just the active one?** What edge cases does this prevent?

---

## Key Takeaways

1. **Separate server and client state.** React Query for API data, Zustand for UI state.

2. **Query key factory for type safety.** Centralize keys, enable hierarchical invalidation.

3. **Optimistic updates for instant feedback.** Snapshot → Update → API → Success/Rollback.

4. **Stale times vary by data.** Frequent changes = short stale time.

5. **Error handling is critical.** Users need clear, actionable error messages.

---

## Next Module

Continue to **[Module 05: Database Design](./05-database-design.md)** to understand schema evolution and migration strategies.
