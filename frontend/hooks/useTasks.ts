'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { taskAPI, CreateTaskDTO, getApiErrorMessage, AchievementEarnedEvent } from '@/lib/api';
import { toast } from 'sonner';
import { gamificationKeys, getAchievementIcon, getAchievementTitle } from './useGamification';

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
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      // Invalidate gamification data to refresh stats
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });

      const gamification = response.data.gamification;

      // Show achievement toasts for new achievements
      if (gamification?.new_achievements && gamification.new_achievements.length > 0) {
        // Show task completed toast first
        toast.success('Task completed!');

        // Show achievement toasts with a slight delay for each
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
          }, 500 * (index + 1)); // Stagger toasts
        });

        // Show streak extended toast if applicable
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
        // No achievements, just show regular completion toast
        toast.success('Task completed!');

        // Show streak extended toast if applicable
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

// Hook for fetching completed tasks (for archive/completed views)
export function useCompletedTasks(filters?: Omit<TaskFilters, 'status'>) {
  return useQuery({
    queryKey: ['tasks', 'completed', filters],
    queryFn: async () => {
      const response = await taskAPI.list({
        limit: 100,
        offset: 0,
        ...filters,
        status: 'done',
      });
      // Sort by updated_at descending (most recently completed first)
      const sortedTasks = response.data.tasks.sort((a, b) => {
        const dateA = new Date(a.updated_at).getTime();
        const dateB = new Date(b.updated_at).getTime();
        return dateB - dateA;
      });
      return { ...response.data, tasks: sortedTasks };
    },
  });
}

// Hook for bulk deleting tasks
export function useBulkDelete() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskIds: string[]) => taskAPI.bulkDelete(taskIds),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success(response.data.message);
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to delete tasks', 'Bulk Delete'));
    },
  });
}

// Hook for bulk restoring tasks (completed -> todo)
export function useBulkRestore() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskIds: string[]) => taskAPI.bulkRestore(taskIds),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success(response.data.message);
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to restore tasks', 'Bulk Restore'));
    },
  });
}
