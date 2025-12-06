'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { subtaskAPI, CreateSubtaskDTO, getApiErrorMessage } from '@/lib/api';
import { toast } from 'sonner';

/**
 * Hook for fetching subtasks of a parent task
 */
export function useSubtasks(parentTaskId: string) {
  return useQuery({
    queryKey: ['subtasks', parentTaskId],
    queryFn: async () => {
      const response = await subtaskAPI.list(parentTaskId);
      return response.data;
    },
    enabled: !!parentTaskId,
  });
}

/**
 * Hook for fetching subtask info (aggregated statistics) for a parent task
 */
export function useSubtaskInfo(parentTaskId: string) {
  return useQuery({
    queryKey: ['subtask-info', parentTaskId],
    queryFn: async () => {
      const response = await subtaskAPI.getInfo(parentTaskId);
      return response.data;
    },
    enabled: !!parentTaskId,
  });
}

/**
 * Hook for fetching a task with its subtask info and optionally expanded subtasks
 */
export function useTaskWithSubtasks(taskId: string, includeSubtasks = false) {
  return useQuery({
    queryKey: ['task-expanded', taskId, includeSubtasks],
    queryFn: async () => {
      const response = await subtaskAPI.getExpanded(taskId, includeSubtasks);
      return response.data;
    },
    enabled: !!taskId,
  });
}

/**
 * Hook for checking if a parent task can be completed (all subtasks done)
 */
export function useCanCompleteParent(parentTaskId: string) {
  return useQuery({
    queryKey: ['can-complete', parentTaskId],
    queryFn: async () => {
      const response = await subtaskAPI.canComplete(parentTaskId);
      return response.data.can_complete;
    },
    enabled: !!parentTaskId,
  });
}

/**
 * Hook for creating a subtask under a parent task
 */
export function useCreateSubtask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ parentTaskId, data }: { parentTaskId: string; data: CreateSubtaskDTO }) =>
      subtaskAPI.create(parentTaskId, data),
    onSuccess: (_response, variables) => {
      // Invalidate both subtasks list and parent task info
      queryClient.invalidateQueries({ queryKey: ['subtasks', variables.parentTaskId] });
      queryClient.invalidateQueries({ queryKey: ['subtask-info', variables.parentTaskId] });
      queryClient.invalidateQueries({ queryKey: ['task-expanded', variables.parentTaskId] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Subtask created!');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to create subtask', 'Subtask Create'));
    },
  });
}

/**
 * Hook for completing a subtask with parent completion prompt
 * Returns special response to indicate if all subtasks are now complete
 */
export function useCompleteSubtask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (subtaskId: string) => subtaskAPI.complete(subtaskId),
    onSuccess: (response) => {
      const { completed_task, all_subtasks_complete, message } = response.data;
      const parentTaskId = completed_task.parent_task_id;

      // Invalidate relevant queries
      if (parentTaskId) {
        queryClient.invalidateQueries({ queryKey: ['subtasks', parentTaskId] });
        queryClient.invalidateQueries({ queryKey: ['subtask-info', parentTaskId] });
        queryClient.invalidateQueries({ queryKey: ['task-expanded', parentTaskId] });
        queryClient.invalidateQueries({ queryKey: ['can-complete', parentTaskId] });
      }
      queryClient.invalidateQueries({ queryKey: ['tasks'] });

      // Show appropriate message
      if (all_subtasks_complete && message) {
        toast.success(message, {
          duration: 5000,
          action: {
            label: 'Complete Parent',
            onClick: () => {
              // This will be handled by the component that uses this hook
              // The parent task ID is available in the response
            },
          },
        });
      } else {
        toast.success('Subtask completed!');
      }
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to complete subtask', 'Subtask Complete'));
    },
  });
}
