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

// =============================================================================
// User Types
// =============================================================================

export type UserType = 'registered' | 'anonymous';

export interface User {
  id: string;
  user_type: UserType;
  email?: string;
  name?: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
}

export interface ConvertGuestDTO {
  email: string;
  name: string;
  password: string;
}

// Auth API
export const authAPI = {
  register: (data: { email: string; name: string; password: string }) =>
    api.post<AuthResponse>('/api/v1/auth/register', data),

  login: (data: { email: string; password: string }) =>
    api.post<AuthResponse>('/api/v1/auth/login', data),

  guest: () =>
    api.post<AuthResponse>('/api/v1/auth/guest'),

  convert: (data: ConvertGuestDTO) =>
    api.post<AuthResponse>('/api/v1/auth/convert', data),

  me: () => api.get<User>('/api/v1/auth/me'),
};

// =============================================================================
// Recurrence Types
// =============================================================================

export type RecurrencePattern = 'none' | 'daily' | 'weekly' | 'monthly';
export type DueDateCalculation = 'from_original' | 'from_completion';

export interface RecurrenceRule {
  pattern: RecurrencePattern;
  interval_value?: number; // e.g., 2 for "every 2 weeks"
  end_date?: string;
  due_date_calculation?: DueDateCalculation;
}

export interface TaskSeries {
  id: string;
  user_id: string;
  original_task_id: string;
  pattern: RecurrencePattern;
  interval_value: number;
  end_date?: string;
  due_date_calculation: DueDateCalculation;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpdateTaskSeriesDTO {
  pattern?: RecurrencePattern;
  interval_value?: number;
  end_date?: string | null;
  due_date_calculation?: DueDateCalculation;
  is_active?: boolean;
}

export interface SeriesHistoryEntry {
  task_id: string;
  title: string;
  status: 'todo' | 'in_progress' | 'done' | 'on_hold' | 'blocked';
  due_date?: string;
  completed_at?: string;
  created_at: string;
}

export interface SeriesHistory {
  series: TaskSeries;
  tasks: SeriesHistoryEntry[];
  total: number;
}

// User Preferences Types
export interface UserPreferences {
  user_id: string;
  default_due_date_calculation: DueDateCalculation;
  created_at: string;
  updated_at: string;
}

export interface CategoryPreference {
  user_id: string;
  category: string;
  due_date_calculation: DueDateCalculation;
  created_at: string;
  updated_at: string;
}

export interface AllPreferences {
  user_preferences?: UserPreferences;
  category_preferences: CategoryPreference[];
}

// =============================================================================
// Task Types
// =============================================================================

export interface CreateTaskDTO {
  title: string;
  description?: string;
  user_priority?: number;
  due_date?: string; // RFC3339/ISO 8601 format (e.g., "2025-11-25T00:00:00Z")
  estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
  category?: string;
  context?: string;
  related_people?: string[];
  recurrence?: RecurrenceRule;
}

// PriorityBreakdown shows the individual components of the priority calculation
export interface PriorityBreakdown {
  // Raw component values (0-100 scale, except effort_boost which is 1.0-1.3)
  user_priority: number;     // User's set priority scaled to 0-100
  time_decay: number;        // Age-based urgency (0-100)
  deadline_urgency: number;  // Due date proximity (0-100)
  bump_penalty: number;      // Penalty for delays (0-50)
  effort_boost: number;      // Effort multiplier (1.0-1.3)

  // Weighted contributions (after applying weights and effort boost)
  user_priority_weighted: number;    // × 0.4 × effort_boost
  time_decay_weighted: number;       // × 0.3 × effort_boost
  deadline_urgency_weighted: number; // × 0.2 × effort_boost
  bump_penalty_weighted: number;     // × 0.1 × effort_boost
}

export type TaskType = 'regular' | 'recurring' | 'subtask';

export interface Task {
  id: string;
  user_id: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'done' | 'on_hold' | 'blocked';
  task_type: TaskType;
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
  // Relationship fields
  series_id?: string;      // Links to task_series if recurring
  parent_task_id?: string; // For subtasks: parent task; for recurring: previous in series
  // Optional: populated when fetching single task details
  priority_breakdown?: PriorityBreakdown;
}

// =============================================================================
// Subtask Types
// =============================================================================

export interface SubtaskInfo {
  total_count: number;
  completed_count: number;
  in_progress_count: number;
  todo_count: number;
  completion_rate: number; // 0.0 - 1.0
  all_complete: boolean;
}

export interface TaskWithSubtasks extends Task {
  subtask_info?: SubtaskInfo;
  subtasks?: Task[];
}

export interface CreateSubtaskDTO {
  title: string;
  description?: string;
  user_priority?: number;
  due_date?: string; // YYYY-MM-DD format
  estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
  context?: string;
}

export interface SubtaskCompletionResponse {
  completed_task: Task;
  all_subtasks_complete: boolean;
  parent_task?: Task;
  message?: string;
}

// =============================================================================
// Dependency Types
// =============================================================================

export interface DependencyWithTask {
  task_id: string;
  title: string;
  status: 'todo' | 'in_progress' | 'done' | 'on_hold' | 'blocked';
  created_at: string;
}

export interface DependencyInfo {
  task_id: string;
  blockers: DependencyWithTask[];    // Tasks blocking this task
  blocking: DependencyWithTask[];    // Tasks this task is blocking
  is_blocked: boolean;               // Has incomplete blockers
  can_complete: boolean;             // All blockers are done
}

export interface BlockerCompletionInfo {
  completed_task_id: string;
  unblocked_task_ids: string[];
  unblocked_count: number;
}

export interface AddDependencyDTO {
  blocked_by_id: string;
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

  // Complete task (returns full response with gamification data)
  complete: (id: string) =>
    api.post<TaskCompletionResponse>(`/api/v1/tasks/${id}/complete`),

  // Uncomplete task (revert completion back to todo)
  uncomplete: (id: string) =>
    api.post<Task>(`/api/v1/tasks/${id}/uncomplete`),

  // Delete task (soft delete)
  delete: (id: string) =>
    api.delete(`/api/v1/tasks/${id}`),

  // Restore a deleted task
  restore: (id: string) =>
    api.post<Task>(`/api/v1/tasks/${id}/restore`),

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

// =============================================================================
// Series API - Task Series Management
// =============================================================================

export const seriesAPI = {
  // List all task series for the user
  list: (params?: { active_only?: boolean }) =>
    api.get<{ series: TaskSeries[]; total_count: number }>('/api/v1/series', { params }),

  // Get history of a specific series (all tasks in the series)
  getHistory: (seriesId: string) =>
    api.get<SeriesHistory>(`/api/v1/series/${seriesId}/history`),

  // Update series settings
  update: (seriesId: string, data: UpdateTaskSeriesDTO) =>
    api.put<TaskSeries>(`/api/v1/series/${seriesId}`, data),

  // Deactivate (stop) a recurring series
  deactivate: (seriesId: string) =>
    api.post<{ message: string }>(`/api/v1/series/${seriesId}/deactivate`),
};

// =============================================================================
// Recurrence Preferences API
// =============================================================================

export const recurrencePreferencesAPI = {
  // Get all user preferences for recurring tasks
  getAll: () =>
    api.get<AllPreferences>('/api/v1/preferences/recurrence'),

  // Get effective due date calculation for a category (considers hierarchy)
  getEffective: (category?: string) =>
    api.get<{ due_date_calculation: DueDateCalculation; category?: string }>(
      '/api/v1/preferences/recurrence/effective',
      { params: { category } }
    ),

  // Set default due date calculation preference
  setDefault: (dueDateCalculation: DueDateCalculation) =>
    api.put<{ message: string; due_date_calculation: DueDateCalculation }>(
      '/api/v1/preferences/recurrence/default',
      { due_date_calculation: dueDateCalculation }
    ),

  // Set category-specific due date calculation preference
  setCategoryPreference: (category: string, dueDateCalculation: DueDateCalculation) =>
    api.put<{ message: string; category: string; due_date_calculation: DueDateCalculation }>(
      `/api/v1/preferences/recurrence/category/${encodeURIComponent(category)}`,
      { due_date_calculation: dueDateCalculation }
    ),

  // Delete a category preference
  deleteCategoryPreference: (category: string) =>
    api.delete(`/api/v1/preferences/recurrence/category/${encodeURIComponent(category)}`),
};

// =============================================================================
// Subtask API
// =============================================================================

export const subtaskAPI = {
  // Create a subtask under a parent task
  create: (parentTaskId: string, data: CreateSubtaskDTO) =>
    api.post<Task>(`/api/v1/tasks/${parentTaskId}/subtasks`, data),

  // Get all subtasks for a parent task
  list: (parentTaskId: string) =>
    api.get<{ subtasks: Task[]; total_count: number }>(`/api/v1/tasks/${parentTaskId}/subtasks`),

  // Get aggregated subtask statistics for a parent task
  getInfo: (parentTaskId: string) =>
    api.get<SubtaskInfo>(`/api/v1/tasks/${parentTaskId}/subtask-info`),

  // Get task with subtask info and optionally expanded subtasks
  getExpanded: (taskId: string, includeSubtasks = false) =>
    api.get<TaskWithSubtasks>(`/api/v1/tasks/${taskId}/expanded`, {
      params: { include_subtasks: includeSubtasks },
    }),

  // Check if a parent task can be completed (all subtasks done)
  canComplete: (parentTaskId: string) =>
    api.get<{ can_complete: boolean }>(`/api/v1/tasks/${parentTaskId}/can-complete`),

  // Complete a subtask with parent completion prompt info
  complete: (subtaskId: string) =>
    api.post<SubtaskCompletionResponse>(`/api/v1/subtasks/${subtaskId}/complete`),
};

// =============================================================================
// Dependency API
// =============================================================================

export const dependencyAPI = {
  // Get dependency info for a task (blockers and blocking)
  getInfo: (taskId: string) =>
    api.get<DependencyInfo>(`/api/v1/tasks/${taskId}/dependencies`),

  // Add a blocker to a task
  addBlocker: (taskId: string, blockedById: string) =>
    api.post<DependencyInfo>(`/api/v1/tasks/${taskId}/dependencies`, {
      blocked_by_id: blockedById,
    }),

  // Remove a blocker from a task
  removeBlocker: (taskId: string, blockedById: string) =>
    api.delete<{ message: string }>(`/api/v1/tasks/${taskId}/dependencies/${blockedById}`),

  // Check if a task can be completed (no incomplete blockers)
  canComplete: (taskId: string) =>
    api.get<{
      can_complete: boolean;
      is_blocked: boolean;
      incomplete_blockers: number;
    }>(`/api/v1/tasks/${taskId}/can-complete-dependencies`),
};

// =============================================================================
// Task Template Types
// =============================================================================

export type TaskEffort = 'small' | 'medium' | 'large' | 'xlarge';

export interface TaskTemplate {
  id: string;
  user_id: string;
  name: string;
  title: string;
  description?: string;
  category?: string;
  estimated_effort?: TaskEffort;
  user_priority: number;
  context?: string;
  related_people?: string[];
  due_date_offset?: number; // Days from creation
  created_at: string;
  updated_at: string;
}

export interface CreateTaskTemplateDTO {
  name: string;
  title: string;
  description?: string;
  category?: string;
  estimated_effort?: TaskEffort;
  user_priority?: number;
  context?: string;
  related_people?: string[];
  due_date_offset?: number;
}

export interface UpdateTaskTemplateDTO {
  name?: string;
  title?: string;
  description?: string;
  category?: string;
  estimated_effort?: TaskEffort;
  user_priority?: number;
  context?: string;
  related_people?: string[];
  due_date_offset?: number;
}

export interface TaskTemplateListResponse {
  templates: TaskTemplate[];
  total_count: number;
}

// =============================================================================
// Template API
// =============================================================================

export const templateAPI = {
  // Create a new template
  create: (data: CreateTaskTemplateDTO) =>
    api.post<TaskTemplate>('/api/v1/templates', data),

  // List all templates for the user
  list: () =>
    api.get<TaskTemplateListResponse>('/api/v1/templates'),

  // Get a specific template
  getById: (id: string) =>
    api.get<TaskTemplate>(`/api/v1/templates/${id}`),

  // Update a template
  update: (id: string, data: UpdateTaskTemplateDTO) =>
    api.put<TaskTemplate>(`/api/v1/templates/${id}`, data),

  // Delete a template
  delete: (id: string) =>
    api.delete(`/api/v1/templates/${id}`),

  // Get pre-filled CreateTaskDTO from template (for use with CreateTaskDialog)
  use: (id: string, overrides?: Partial<CreateTaskDTO>) =>
    api.post<CreateTaskDTO>(`/api/v1/templates/${id}/use`, overrides || {}),
};

// =============================================================================
// Gamification Types
// =============================================================================

export type AchievementType =
  | 'first_task'
  | 'milestone_10'
  | 'milestone_50'
  | 'milestone_100'
  | 'streak_3'
  | 'streak_7'
  | 'streak_14'
  | 'streak_30'
  | 'category_master'
  | 'speed_demon'
  | 'consistency_king';

export interface AchievementDefinition {
  type: AchievementType;
  title: string;
  description: string;
  icon: string;
  category: 'milestone' | 'streak' | 'special';
}

export interface UserAchievement {
  id: string;
  user_id: string;
  achievement_type: AchievementType;
  earned_at: string;
  metadata?: Record<string, unknown>;
}

export interface AchievementEarnedEvent {
  achievement: UserAchievement;
  definition: AchievementDefinition;
}

export interface GamificationStats {
  user_id: string;
  current_streak: number;
  longest_streak: number;
  last_completion_date?: string;
  total_completed: number;
  productivity_score: number;
  completion_rate: number;
  streak_score: number;
  on_time_percentage: number;
  effort_mix_score: number;
  last_computed_at: string;
  created_at?: string;
  updated_at?: string;
}

export interface CategoryMastery {
  id: string;
  user_id: string;
  category: string;
  completed_count: number;
  last_completed_at: string;
  created_at: string;
  updated_at: string;
}

export interface GamificationDashboard {
  stats: GamificationStats;
  recent_achievements: UserAchievement[];
  all_achievements: UserAchievement[];
  available_achievements: AchievementDefinition[];
  category_progress: CategoryMastery[];
  unviewed_count: number;
}

export interface TaskCompletionGamificationResult {
  updated_stats: GamificationStats;
  new_achievements: AchievementEarnedEvent[];
  streak_extended: boolean;
  previous_streak: number;
}

// Full task completion response (includes recurring task and gamification data)
export interface TaskCompletionResponse {
  completed_task: Task;
  next_task?: Task; // For recurring tasks
  series?: TaskSeries; // For recurring tasks
  message: string;
  gamification?: TaskCompletionGamificationResult;
}

// =============================================================================
// Gamification API
// =============================================================================

export const gamificationAPI = {
  // Get full dashboard data (stats + achievements + category progress)
  getDashboard: () =>
    api.get<GamificationDashboard>('/api/v1/gamification/dashboard'),

  // Get current stats only (lighter endpoint for sidebar widget)
  getStats: () =>
    api.get<GamificationStats>('/api/v1/gamification/stats'),

  // Get user's timezone setting
  getTimezone: () =>
    api.get<{ timezone: string }>('/api/v1/gamification/timezone'),

  // Update user's timezone for streak calculation
  setTimezone: (timezone: string) =>
    api.put<{ message: string; timezone: string }>(
      '/api/v1/gamification/timezone',
      { timezone }
    ),
};
