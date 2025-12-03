import axios, { AxiosError } from 'axios';

// =============================================================================
// Error Handling Utilities
// =============================================================================

/**
 * Type guard to check if an error is an Axios error.
 * Note: Returns true for both errors with responses (API errors)
 * and errors without responses (network errors).
 */
export function isAxiosError(error: unknown): error is AxiosError<{ error?: string }> {
  return axios.isAxiosError(error);
}

/**
 * Extract a user-friendly error message from an API error.
 * Handles Axios errors (with status-specific messages for 401, 403, 404, 5xx),
 * network errors, timeouts, and generic Error objects.
 *
 * @param err - The caught error (unknown type)
 * @param fallback - Default message if error cannot be parsed
 * @param logContext - Optional context string. When provided, logs error via console.error
 * @returns A user-friendly error message
 *
 * @example
 * try {
 *   await api.post('/tasks', data);
 * } catch (err) {
 *   toast.error(getApiErrorMessage(err, 'Failed to create task', 'TaskCreate'));
 * }
 */
export function getApiErrorMessage(
  err: unknown,
  fallback: string,
  logContext?: string
): string {
  // Always log the error for debugging
  if (logContext) {
    console.error(`[${logContext}]`, err);
  }

  // Axios error with response (API returned an error)
  if (isAxiosError(err) && err.response?.data?.error) {
    return err.response.data.error;
  }

  // Axios network error (no response received)
  if (isAxiosError(err) && !err.response) {
    if (err.code === 'ECONNABORTED') {
      return 'Request timed out. Please try again.';
    }
    if (err.code === 'ERR_NETWORK') {
      return 'Network error. Please check your connection.';
    }
    return 'Unable to connect to server. Please try again.';
  }

  // Axios error with response but no specific error message
  if (isAxiosError(err) && err.response) {
    const status = err.response.status;
    if (status === 401) return 'Session expired. Please log in again.';
    if (status === 403) return 'You do not have permission to perform this action.';
    if (status === 404) return 'The requested resource was not found.';
    if (status >= 500) return 'Server error. Please try again later.';
  }

  // Standard Error object
  if (err instanceof Error) {
    return err.message || fallback;
  }

  return fallback;
}

/**
 * Check if an error is an authentication or authorization error (401/403).
 * Used to determine if user credentials/session are invalid.
 */
export function isAuthError(err: unknown): boolean {
  if (isAxiosError(err) && err.response) {
    return err.response.status === 401 || err.response.status === 403;
  }
  return false;
}

/**
 * Check if an error is a network/connection error.
 * Returns true when the request was made but no response was received
 * (e.g., server unreachable, DNS failure, CORS blocked, timeout).
 */
export function isNetworkError(err: unknown): boolean {
  if (isAxiosError(err) && !err.response) {
    return true;
  }
  return false;
}

// =============================================================================
// API Client Setup
// =============================================================================

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Auth API
export const authAPI = {
  register: (data: { email: string; name: string; password: string }) =>
    api.post('/api/v1/auth/register', data),

  login: (data: { email: string; password: string }) =>
    api.post('/api/v1/auth/login', data),

  me: () => api.get('/api/v1/auth/me'),
};

// Task Types
export interface CreateTaskDTO {
  title: string;
  description?: string;
  user_priority?: number;
  due_date?: string; // RFC3339/ISO 8601 format (e.g., "2025-11-25T00:00:00Z")
  estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
  category?: string;
  context?: string;
  related_people?: string[];
}

export interface Task {
  id: string;
  user_id: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'done';
  user_priority: number;
  due_date?: string;
  estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
  category?: string;
  context?: string;
  related_people?: string[];
  priority_score: number;
  bump_count: number;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

// Calendar Types
export interface CalendarDayData {
  count: number;
  badge_color: 'red' | 'yellow' | 'blue';
  tasks: Task[];
}

export interface CalendarResponse {
  dates: Record<string, CalendarDayData>; // key format: "2025-11-23"
}

// Task API
export const taskAPI = {
  // Create a new task
  create: (data: CreateTaskDTO) =>
    api.post<Task>('/api/v1/tasks', data),

  // Get all tasks (priority-sorted) with optional filters
  list: (params?: {
    limit?: number;
    offset?: number;
    status?: string;
    category?: string;
    search?: string;
    min_priority?: number;
    max_priority?: number;
    due_date_start?: string;
    due_date_end?: string;
  }) =>
    api.get<{ tasks: Task[]; total_count: number }>('/api/v1/tasks', { params }),

  // Get single task by ID
  getById: (id: string) =>
    api.get<Task>(`/api/v1/tasks/${id}`),

  // Update task
  update: (id: string, data: Partial<CreateTaskDTO>) =>
    api.put<Task>(`/api/v1/tasks/${id}`, data),

  // Bump task (delay it)
  bump: (id: string, reason?: string) =>
    api.post<{ message: string; task: Task }>(`/api/v1/tasks/${id}/bump`, { reason }),

  // Complete task
  complete: (id: string) =>
    api.post<Task>(`/api/v1/tasks/${id}/complete`),

  // Delete task
  delete: (id: string) =>
    api.delete(`/api/v1/tasks/${id}`),

  // Get at-risk tasks (bumped 3+ times)
  getAtRisk: () =>
    api.get<{ tasks: Task[]; count: number }>('/api/v1/tasks/at-risk'),

  // Get calendar data for date range
  getCalendar: (params: {
    start_date: string; // YYYY-MM-DD format
    end_date: string;   // YYYY-MM-DD format
    status?: string;    // Optional: comma-separated statuses (e.g., "todo,in_progress")
  }) =>
    api.get<CalendarResponse>('/api/v1/tasks/calendar', { params }),

  // Bulk delete tasks
  bulkDelete: (taskIds: string[]) =>
    api.post<BulkOperationResponse>('/api/v1/tasks/bulk-delete', { task_ids: taskIds }),

  // Bulk restore tasks (completed -> todo)
  bulkRestore: (taskIds: string[]) =>
    api.post<BulkOperationResponse>('/api/v1/tasks/bulk-restore', { task_ids: taskIds }),
};

// Bulk operation response type
export interface BulkOperationResponse {
  success_count: number;
  failed_ids?: string[];
  message: string;
}

// Category API
export const categoryAPI = {
  // Rename a category (updates all tasks with old category to new category)
  rename: (oldName: string, newName: string) =>
    api.put('/api/v1/categories/rename', { old_name: oldName, new_name: newName }),

  // Delete a category (removes category from all tasks)
  delete: (name: string) =>
    api.delete(`/api/v1/categories/${encodeURIComponent(name)}`),
};

// Analytics Types
export interface CompletionStats {
  total_tasks: number;
  completed_tasks: number;
  pending_tasks: number;
  completion_rate: number;
}

export interface BumpAnalytics {
  average_bump_count: number;
  at_risk_count: number;
  bump_distribution: Record<string, number>;
}

export interface CategoryStat {
  category: string;
  task_count: number;
  completion_rate: number;
}

export interface PriorityDistribution {
  priority_range: string;
  task_count: number;
}

export interface VelocityMetric {
  date: string;
  completed_count: number;
}

export interface AnalyticsSummary {
  period_days: number;
  completion_stats: CompletionStats;
  bump_analytics: BumpAnalytics;
  category_breakdown: CategoryStat[];
  priority_distribution: PriorityDistribution[];
}

export interface AnalyticsTrends {
  period_days: number;
  velocity_metrics: VelocityMetric[];
}

// Heatmap types
export interface HeatmapCell {
  day_of_week: number; // 0 = Sunday, 6 = Saturday
  hour: number; // 0-23
  count: number;
}

export interface ProductivityHeatmap {
  cells: HeatmapCell[];
  max_count: number;
}

export interface HeatmapResponse {
  period_days: number;
  heatmap: ProductivityHeatmap;
}

// Category trends types
export interface CategoryTrendPoint {
  week_start: string; // YYYY-MM-DD
  categories: Record<string, number>;
}

export interface CategoryTrends {
  weeks: CategoryTrendPoint[];
  categories: string[];
}

export interface CategoryTrendsResponse {
  period_days: number;
  trends: CategoryTrends;
}

// Analytics API
export const analyticsAPI = {
  // Get comprehensive analytics summary
  getSummary: (params?: { days?: number }) =>
    api.get<AnalyticsSummary>('/api/v1/analytics/summary', { params }),

  // Get completion trends over time
  getTrends: (params?: { days?: number }) =>
    api.get<AnalyticsTrends>('/api/v1/analytics/trends', { params }),

  // Get productivity heatmap data
  getHeatmap: (params?: { days?: number }) =>
    api.get<HeatmapResponse>('/api/v1/analytics/heatmap', { params }),

  // Get category trends over time
  getCategoryTrends: (params?: { days?: number }) =>
    api.get<CategoryTrendsResponse>('/api/v1/analytics/category-trends', { params }),
};

// =============================================================================
// Insights Types & API
// =============================================================================

export type InsightType =
  | 'avoidance_pattern'
  | 'peak_performance'
  | 'quick_wins'
  | 'deadline_clustering'
  | 'at_risk_alert'
  | 'category_overload';

export type InsightPriority = 1 | 2 | 3 | 4 | 5;

export interface Insight {
  type: InsightType;
  title: string;
  message: string;
  priority: InsightPriority;
  action_url?: string;
  data?: Record<string, unknown>;
  generated_at: string;
}

export interface InsightResponse {
  insights: Insight[];
  cached_at: string;
}

export interface TimeEstimate {
  estimated_days: number;
  confidence_level: 'low' | 'medium' | 'high';
  based_on: number;
  factors: {
    base_estimate: number;
    category_factor: number;
    effort_factor: number;
    bump_factor: number;
  };
}

export interface CategorySuggestion {
  category: string;
  confidence: number;
  matched_keywords?: string[];
}

export interface CategorySuggestionResponse {
  suggestions: CategorySuggestion[];
}

// Insights API
export const insightsAPI = {
  // Get smart suggestions and insights
  getInsights: () =>
    api.get<InsightResponse>('/api/v1/insights'),

  // Get time estimate for a specific task
  getTimeEstimate: (taskId: string) =>
    api.get<TimeEstimate>(`/api/v1/tasks/${taskId}/estimate`),

  // Get category suggestions based on task content
  suggestCategory: (data: { title: string; description?: string }) =>
    api.post<CategorySuggestionResponse>('/api/v1/tasks/suggest-category', data),
};
