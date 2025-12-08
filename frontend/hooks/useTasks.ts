'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { taskAPI, CreateTaskDTO, getApiErrorMessage, AchievementEarnedEvent, Task } from '@/lib/api';
import { toast } from 'sonner';
import { gamificationKeys, getAchievementIcon, getAchievementTitle } from './useGamification';
import { taskKeys, analyticsKeys } from '@/lib/queryKeys';

// Re-export TaskFilters from queryKeys for backward compatibility
export type { TaskFilters } from '@/lib/queryKeys';

// Type for task list response
interface TaskListResponse {
  tasks: Task[];
  total_count: number;
}

export function useTasks(filters?: Parameters<typeof taskKeys.list>[0]) {
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
    staleTime: 2 * 60 * 1000, // 2 minutes - tasks don't change that often
  });
}

export function useTask(id: string) {
  return useQuery({
    queryKey: taskKeys.detail(id),
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
    // Optimistic update: immediately add task to list with placeholder priority
    onMutate: async (newTask) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      // Snapshot all list queries for rollback
      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });

      // Optimistically add the new task to all list caches
      const optimisticTask: Task = {
        id: `temp-${Date.now()}`, // Temporary ID
        title: newTask.title,
        description: newTask.description,
        category: newTask.category,
        due_date: newTask.due_date,
        estimated_effort: newTask.estimated_effort,
        priority_score: 50, // Placeholder - will be calculated by backend
        status: 'todo',
        bump_count: 0,
        user_priority: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        user_id: '', // Will be set by backend
        task_type: 'regular',
      };

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
      // Invalidate lists only to get the real task with proper ID and priority
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      // Also invalidate analytics since task counts changed
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
      toast.success('Task created with priority calculated!');
    },
    onError: (err: unknown, _variables, context) => {
      // Rollback on error
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      toast.error(getApiErrorMessage(err, 'Failed to create task', 'Task Create'));
    },
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CreateTaskDTO> }) =>
      taskAPI.update(id, data),
    // Optimistic update: immediately update task in cache
    onMutate: async ({ id, data }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });
      await queryClient.cancelQueries({ queryKey: taskKeys.detail(id) });

      // Snapshot for rollback
      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });
      const previousDetail = queryClient.getQueryData<Task>(taskKeys.detail(id));

      // Optimistically update in all list caches
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: taskKeys.lists() },
        (old) => {
          if (!old?.tasks) return old;
          return {
            ...old,
            tasks: old.tasks.map((task) =>
              task.id === id ? { ...task, ...data, updated_at: new Date().toISOString() } : task
            ),
          };
        }
      );

      // Optimistically update detail cache
      if (previousDetail) {
        queryClient.setQueryData<Task>(taskKeys.detail(id), {
          ...previousDetail,
          ...data,
          updated_at: new Date().toISOString(),
        });
      }

      return { previousLists, previousDetail, taskId: id };
    },
    onSuccess: (response, { id }) => {
      // Update the detail cache with the real response
      queryClient.setQueryData(taskKeys.detail(id), response.data);
      // Invalidate lists to ensure consistency
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      toast.success(`Task updated! New priority: ${Math.round(response.data.priority_score)}`);
    },
    onError: (err: unknown, { id }, context) => {
      // Rollback on error
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      if (context?.previousDetail) {
        queryClient.setQueryData(taskKeys.detail(id), context.previousDetail);
      }
      toast.error(getApiErrorMessage(err, 'Failed to update task', 'Task Update'));
    },
  });
}

export function useBumpTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      taskAPI.bump(id, reason),
    // Optimistic update: immediately increment bump count
    onMutate: async ({ id }) => {
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });

      // Optimistically update bump count
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: taskKeys.lists() },
        (old) => {
          if (!old?.tasks) return old;
          return {
            ...old,
            tasks: old.tasks.map((task) =>
              task.id === id
                ? { ...task, bump_count: task.bump_count + 1, updated_at: new Date().toISOString() }
                : task
            ),
          };
        }
      );

      return { previousLists };
    },
    onSuccess: (response) => {
      // Invalidate to get the real priority score
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      // Invalidate at-risk since bump count changed
      queryClient.invalidateQueries({ queryKey: taskKeys.atRisk() });
      toast.info(`Task delayed. New priority: ${Math.round(response.data.task.priority_score)}`);
    },
    onError: (err: unknown, _variables, context) => {
      // Rollback on error
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      toast.error(getApiErrorMessage(err, 'Failed to bump task', 'Task Bump'));
    },
  });
}

export function useCompleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.complete(id),
    // Optimistic update: immediately mark task as completed in cache
    onMutate: async (taskId: string) => {
      try {
        // Cancel outgoing refetches for lists only (not details, not gamification yet)
        await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

        // Snapshot previous task list queries for rollback
        const previousTaskQueries = queryClient.getQueriesData<TaskListResponse>({
          queryKey: taskKeys.lists(),
        });

        // Optimistically update all task list caches
        queryClient.setQueriesData<TaskListResponse>(
          { queryKey: taskKeys.lists() },
          (old) => {
            if (!old?.tasks) return old;
            return {
              ...old,
              tasks: old.tasks.map((task) =>
                task.id === taskId
                  ? { ...task, status: 'done' as const, completed_at: new Date().toISOString() }
                  : task
              ),
            };
          }
        );

        // Return context for rollback
        return { previousTaskQueries };
      } catch (error) {
        // Log but don't throw - mutation should still proceed even if optimistic update fails
        console.error('[useCompleteTask.onMutate] Optimistic update failed:', error);
        return { previousTaskQueries: [] };
      }
    },
    onSuccess: (response) => {
      // Targeted invalidation: lists, completed, at-risk, gamification, analytics
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: taskKeys.completed() });
      queryClient.invalidateQueries({ queryKey: taskKeys.atRisk() });
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });

      const gamification = response.data.gamification;

      // Show achievement toasts for new achievements
      if (gamification?.new_achievements && gamification.new_achievements.length > 0) {
        toast.success('Task completed!');

        gamification.new_achievements.forEach((achievement: AchievementEarnedEvent, index: number) => {
          setTimeout(() => {
            const icon = getAchievementIcon(achievement.achievement.achievement_type);
            const title = getAchievementTitle(achievement.achievement);
            toast.success(
              `${icon} Achievement Unlocked: ${title}!`,
              {
                description: achievement.definition.description,
                duration: 5000,
              }
            );
          }, 500 * (index + 1));
        });

        if (gamification.streak_extended && gamification.updated_stats.current_streak > 1) {
          setTimeout(() => {
            toast.success(
              `🔥 ${gamification.updated_stats.current_streak}-day streak!`,
              {
                description: 'Keep up the momentum!',
                duration: 3000,
              }
            );
          }, 500 * (gamification.new_achievements.length + 1));
        }
      } else {
        toast.success('Task completed!');

        if (gamification?.streak_extended && gamification.updated_stats.current_streak > 1) {
          setTimeout(() => {
            toast.success(
              `🔥 ${gamification.updated_stats.current_streak}-day streak!`,
              { duration: 3000 }
            );
          }, 500);
        }
      }
    },
    onError: (err: unknown, _taskId: string, context) => {
      // Rollback to previous state on error
      if (context?.previousTaskQueries && context.previousTaskQueries.length > 0) {
        context.previousTaskQueries.forEach(([queryKey, data]) => {
          if (data !== undefined) {
            queryClient.setQueryData(queryKey, data);
          } else {
            queryClient.invalidateQueries({ queryKey });
          }
        });
      }
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
      toast.error(getApiErrorMessage(err, 'Failed to complete task', 'Task Complete'));
    },
  });
}

export function useDeleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.delete(id),
    // Optimistic update: immediately remove task from list
    onMutate: async (taskId: string) => {
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });

      // Optimistically remove the task from all list caches
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: taskKeys.lists() },
        (old) => {
          if (!old?.tasks) return old;
          return {
            ...old,
            tasks: old.tasks.filter((task) => task.id !== taskId),
            total_count: old.total_count - 1,
          };
        }
      );

      return { previousLists };
    },
    onSuccess: () => {
      // Targeted invalidation
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
      toast.success('Task deleted');
    },
    onError: (err: unknown, _taskId, context) => {
      // Rollback on error
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      toast.error(getApiErrorMessage(err, 'Failed to delete task', 'Task Delete'));
    },
  });
}

export function useAtRiskTasks() {
  return useQuery({
    queryKey: taskKeys.atRisk(),
    queryFn: async () => {
      const response = await taskAPI.list({ limit: 100, offset: 0 });
      const atRiskTasks = response.data.tasks.filter(t => t.bump_count >= 3);
      return { tasks: atRiskTasks, count: atRiskTasks.length };
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
  });
}

export function useCalendarTasks(params: {
  start_date: string;
  end_date: string;
  status?: string;
}) {
  return useQuery({
    queryKey: taskKeys.calendar(params),
    queryFn: async () => {
      const response = await taskAPI.getCalendar(params);
      return response.data;
    },
    enabled: !!params.start_date && !!params.end_date,
    staleTime: 5 * 60 * 1000, // 5 minutes - calendar data doesn't need to refetch as often
  });
}

export function useCompletedTasks(filters?: Omit<Parameters<typeof taskKeys.list>[0], 'status'>) {
  return useQuery({
    queryKey: taskKeys.completed(filters),
    queryFn: async () => {
      const response = await taskAPI.list({
        limit: 100,
        offset: 0,
        ...filters,
        status: 'done',
      });
      const sortedTasks = response.data.tasks.sort((a, b) => {
        const dateA = new Date(a.updated_at).getTime();
        const dateB = new Date(b.updated_at).getTime();
        return dateB - dateA;
      });
      return { ...response.data, tasks: sortedTasks };
    },
    staleTime: 5 * 60 * 1000, // 5 minutes - completed tasks rarely change
  });
}

export function useBulkDelete() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskIds: string[]) => taskAPI.bulkDelete(taskIds),
    // Optimistic update: immediately remove tasks from list
    onMutate: async (taskIds: string[]) => {
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      const previousLists = queryClient.getQueriesData<TaskListResponse>({
        queryKey: taskKeys.lists(),
      });

      const taskIdSet = new Set(taskIds);
      queryClient.setQueriesData<TaskListResponse>(
        { queryKey: taskKeys.lists() },
        (old) => {
          if (!old?.tasks) return old;
          const filtered = old.tasks.filter((task) => !taskIdSet.has(task.id));
          return {
            ...old,
            tasks: filtered,
            total_count: old.total_count - (old.tasks.length - filtered.length),
          };
        }
      );

      return { previousLists };
    },
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: taskKeys.completed() });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
      toast.success(response.data.message);
    },
    onError: (err: unknown, _taskIds, context) => {
      if (context?.previousLists) {
        context.previousLists.forEach(([queryKey, data]) => {
          if (data) {
            queryClient.setQueryData(queryKey, data);
          }
        });
      }
      toast.error(getApiErrorMessage(err, 'Failed to delete tasks', 'Bulk Delete'));
    },
  });
}

export function useBulkRestore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskIds: string[]) => taskAPI.bulkRestore(taskIds),
    onSuccess: (response) => {
      // Invalidate both lists and completed
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: taskKeys.completed() });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });
      toast.success(response.data.message);
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to restore tasks', 'Bulk Restore'));
    },
  });
}
