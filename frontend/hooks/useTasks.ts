'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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
      const response = await taskAPI.getAtRisk();
      return response.data;
    },
  });
}
