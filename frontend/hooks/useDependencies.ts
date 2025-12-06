'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { dependencyAPI, getApiErrorMessage } from '@/lib/api';
import { toast } from 'sonner';

/**
 * Hook for fetching dependency info for a task
 */
export function useDependencyInfo(taskId: string) {
  return useQuery({
    queryKey: ['dependencies', taskId],
    queryFn: async () => {
      const response = await dependencyAPI.getInfo(taskId);
      return response.data;
    },
    enabled: !!taskId,
  });
}

/**
 * Hook for checking if a task can be completed (no incomplete blockers)
 */
export function useCanCompleteDependencies(taskId: string) {
  return useQuery({
    queryKey: ['dependencies-can-complete', taskId],
    queryFn: async () => {
      const response = await dependencyAPI.canComplete(taskId);
      return response.data;
    },
    enabled: !!taskId,
  });
}

/**
 * Hook for adding a blocker to a task
 */
export function useAddBlocker() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, blockedById }: { taskId: string; blockedById: string }) =>
      dependencyAPI.addBlocker(taskId, blockedById),
    onSuccess: (_response, variables) => {
      // Invalidate dependencies for both tasks
      queryClient.invalidateQueries({ queryKey: ['dependencies', variables.taskId] });
      queryClient.invalidateQueries({ queryKey: ['dependencies', variables.blockedById] });
      queryClient.invalidateQueries({ queryKey: ['dependencies-can-complete', variables.taskId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Blocker added!');
    },
    onError: (err: unknown) => {
      const message = getApiErrorMessage(err, 'Failed to add blocker', 'Add Blocker');
      // Provide more user-friendly messages for common errors
      if (message.includes('cycle')) {
        toast.error('Cannot add: This would create a circular dependency');
      } else if (message.includes('already exists')) {
        toast.error('This dependency already exists');
      } else if (message.includes('regular tasks')) {
        toast.error('Only regular tasks can have dependencies');
      } else {
        toast.error(message);
      }
    },
  });
}

/**
 * Hook for removing a blocker from a task
 */
export function useRemoveBlocker() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ taskId, blockedById }: { taskId: string; blockedById: string }) =>
      dependencyAPI.removeBlocker(taskId, blockedById),
    onSuccess: (_response, variables) => {
      // Invalidate dependencies for both tasks
      queryClient.invalidateQueries({ queryKey: ['dependencies', variables.taskId] });
      queryClient.invalidateQueries({ queryKey: ['dependencies', variables.blockedById] });
      queryClient.invalidateQueries({ queryKey: ['dependencies-can-complete', variables.taskId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Blocker removed!');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to remove blocker', 'Remove Blocker'));
    },
  });
}
