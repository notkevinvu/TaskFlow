import axios from 'axios';

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

// Task API
export const taskAPI = {
  // Create a new task
  create: (data: CreateTaskDTO) =>
    api.post<Task>('/api/v1/tasks', data),

  // Get all tasks (priority-sorted)
  list: (params?: { limit?: number; offset?: number }) =>
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
};
