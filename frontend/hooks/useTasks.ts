'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { taskAPI, CreateTaskDTO } from '@/lib/api';
import { toast } from 'sonner';

// Mock task data for development
const mockTasks = [
  {
    id: '1',
    user_id: 'dev-user-123',
    title: 'Review design doc for auth flow',
    description: 'Need to review the authentication flow design document and provide feedback',
    status: 'todo' as const,
    user_priority: 75,
    due_date: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    estimated_effort: 'small' as const,
    category: 'Code Review',
    context: 'From Alice - needs feedback by tomorrow',
    related_people: ['Alice'],
    priority_score: 95,
    bump_count: 0,
    created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '2',
    user_id: 'dev-user-123',
    title: 'Update README with new API endpoints',
    description: 'Document the new task management endpoints',
    status: 'todo' as const,
    user_priority: 50,
    estimated_effort: 'medium' as const,
    category: 'Documentation',
    priority_score: 78,
    bump_count: 2,
    created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '3',
    user_id: 'dev-user-123',
    title: 'Fix bug in priority calculation',
    description: 'Time decay is not calculating correctly for tasks older than 30 days',
    status: 'todo' as const,
    user_priority: 90,
    due_date: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
    estimated_effort: 'medium' as const,
    category: 'Bug Fix',
    context: 'Reported by QA team - critical for release',
    priority_score: 92,
    bump_count: 0,
    created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '4',
    user_id: 'dev-user-123',
    title: 'Refactor database schema',
    description: 'Normalize the task_history table to reduce redundancy',
    status: 'todo' as const,
    user_priority: 40,
    estimated_effort: 'large' as const,
    category: 'Refactoring',
    priority_score: 55,
    bump_count: 5,
    created_at: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '5',
    user_id: 'dev-user-123',
    title: 'Set up CI/CD pipeline',
    description: 'Configure GitHub Actions for automated testing and deployment',
    status: 'todo' as const,
    user_priority: 60,
    estimated_effort: 'xlarge' as const,
    category: 'DevOps',
    context: 'Team requested for faster deployments',
    related_people: ['DevOps Team'],
    priority_score: 68,
    bump_count: 1,
    created_at: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: '6',
    user_id: 'dev-user-123',
    title: 'Write unit tests for priority service',
    description: 'Ensure priority calculation has good test coverage',
    status: 'todo' as const,
    user_priority: 70,
    estimated_effort: 'small' as const,
    category: 'Testing',
    priority_score: 82,
    bump_count: 0,
    created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      // In development, return mock data
      if (process.env.NODE_ENV === 'development') {
        return { tasks: mockTasks, total_count: mockTasks.length };
      }
      const response = await taskAPI.list({ limit: 100, offset: 0 });
      return response.data;
    },
  });
}

export function useTask(id: string) {
  return useQuery({
    queryKey: ['tasks', id],
    queryFn: async () => {
      // In development, return mock data
      if (process.env.NODE_ENV === 'development') {
        const task = mockTasks.find(t => t.id === id);
        if (!task) {
          throw new Error('Task not found');
        }
        return task;
      }
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
      toast.info(`Task delayed. New priority: ${Math.round(response.data.task.priority_score)}`);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to bump task');
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
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to complete task');
    },
  });
}

export function useDeleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Task deleted');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete task');
    },
  });
}

export function useAtRiskTasks() {
  return useQuery({
    queryKey: ['tasks', 'at-risk'],
    queryFn: async () => {
      // In development, return mock at-risk tasks
      if (process.env.NODE_ENV === 'development') {
        const atRiskTasks = mockTasks.filter(t => t.bump_count >= 3);
        return { tasks: atRiskTasks, count: atRiskTasks.length };
      }
      const response = await taskAPI.getAtRisk();
      return response.data;
    },
  });
}
