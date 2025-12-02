'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { taskAPI, CreateTaskDTO, getApiErrorMessage } from '@/lib/api';
import { toast } from 'sonner';

export interface TaskFilters {
  status?: string;
  category?: string;
  search?: string;
  min_priority?: number;
  max_priority?: number;
  due_date_start?: string; // YYYY-MM-DD
  due_date_end?: string;   // YYYY-MM-DD
  limit?: number;
  offset?: number;
}

export function useTasks(filters?: TaskFilters) {
  return useQuery({
    queryKey: ['tasks', filters],
    queryFn: async () => {
      const response = await taskAPI.list({
        limit: 100,
        offset: 0,
        ...filters,
      });
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
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to create task', 'Task Create'));
    },
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateTaskDTO> }) =>
      taskAPI.update(id, data),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success(`Task updated! New priority: ${Math.round(response.data.priority_score)}`);
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to update task', 'Task Update'));
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
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to bump task', 'Task Bump'));
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
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to complete task', 'Task Complete'));
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
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to delete task', 'Task Delete'));
    },
  });
}

export function useAtRiskTasks() {
  return useQuery({
    queryKey: ['tasks', 'at-risk'],
    queryFn: async () => {
      // Get all tasks and filter for at-risk ones
      const response = await taskAPI.list({ limit: 100, offset: 0 });
      const atRiskTasks = response.data.tasks.filter(t => t.bump_count >= 3);
      return { tasks: atRiskTasks, count: atRiskTasks.length };
    },
  });
}

export function useCalendarTasks(params: {
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  status?: string;
}) {
  return useQuery({
    queryKey: ['tasks', 'calendar', params.start_date, params.end_date, params.status],
    queryFn: async () => {
      const response = await taskAPI.getCalendar(params);
      return response.data;
    },
    enabled: !!params.start_date && !!params.end_date,
    staleTime: 5 * 60 * 1000, // 5 minutes - calendar data doesn't need to refetch as often
  });
}
