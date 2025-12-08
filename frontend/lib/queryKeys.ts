/**
 * Query Key Factory for React Query
 *
 * This module provides type-safe, hierarchical query keys for all data fetching.
 * Using a factory pattern enables:
 * - Targeted cache invalidation (invalidate lists without invalidating details)
 * - Better TypeScript inference
 * - Centralized query key management
 *
 * @see https://tkdodo.eu/blog/effective-react-query-keys
 */

// =============================================================================
// Task Query Keys
// =============================================================================

export interface TaskFilters {
  status?: string;
  category?: string;
  search?: string;
  min_priority?: number;
  max_priority?: number;
  due_date_start?: string;
  due_date_end?: string;
  limit?: number;
  offset?: number;
}

export interface CalendarParams {
  start_date: string;
  end_date: string;
  status?: string;
}

/**
 * Query key factory for task-related queries.
 *
 * Hierarchy:
 * - tasks (all)
 *   - tasks.lists (all list queries)
 *     - tasks.list(filters) (specific filtered list)
 *   - tasks.details (all detail queries)
 *     - tasks.detail(id) (specific task)
 *   - tasks.calendar(params) (calendar view)
 *   - tasks.completed(filters) (completed tasks)
 *   - tasks.atRisk (at-risk tasks)
 */
export const taskKeys = {
  all: ['tasks'] as const,

  // List queries
  lists: () => [...taskKeys.all, 'list'] as const,
  list: (filters?: TaskFilters) => [...taskKeys.lists(), filters ?? {}] as const,

  // Detail queries
  details: () => [...taskKeys.all, 'detail'] as const,
  detail: (id: string) => [...taskKeys.details(), id] as const,

  // Calendar queries
  calendar: (params: CalendarParams) => [...taskKeys.all, 'calendar', params] as const,

  // Completed tasks (separate from active lists)
  completed: (filters?: Omit<TaskFilters, 'status'>) => [...taskKeys.all, 'completed', filters ?? {}] as const,

  // At-risk tasks
  atRisk: () => [...taskKeys.all, 'at-risk'] as const,
};

// =============================================================================
// Analytics Query Keys
// =============================================================================

export const analyticsKeys = {
  all: ['analytics'] as const,
  summary: (days: number) => [...analyticsKeys.all, 'summary', days] as const,
  trends: (days: number) => [...analyticsKeys.all, 'trends', days] as const,
  heatmap: (days: number) => [...analyticsKeys.all, 'heatmap', days] as const,
  categoryTrends: (days: number) => [...analyticsKeys.all, 'categoryTrends', days] as const,
  insights: () => [...analyticsKeys.all, 'insights'] as const,
};

// =============================================================================
// Subtask Query Keys
// =============================================================================

export const subtaskKeys = {
  all: ['subtasks'] as const,
  byParent: (parentId: string) => [...subtaskKeys.all, 'parent', parentId] as const,
};

// =============================================================================
// Template Query Keys
// =============================================================================

export const templateKeys = {
  all: ['templates'] as const,
  lists: () => [...templateKeys.all, 'list'] as const,
  detail: (id: string) => [...templateKeys.all, 'detail', id] as const,
};
