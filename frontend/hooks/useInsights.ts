import { useQuery } from '@tanstack/react-query';
import { insightsAPI, getApiErrorMessage } from '@/lib/api';
import type { InsightResponse, TimeEstimate, CategorySuggestionResponse } from '@/lib/api';

/**
 * Hook to fetch smart insights and suggestions for the current user.
 * Returns actionable insights based on task patterns and behavior.
 */
export function useInsights() {
  return useQuery<InsightResponse>({
    queryKey: ['insights'],
    queryFn: async () => {
      try {
        const response = await insightsAPI.getInsights();
        return response.data;
      } catch (error) {
        throw new Error(getApiErrorMessage(error, 'Failed to fetch insights', 'useInsights'));
      }
    },
    staleTime: 5 * 60 * 1000, // Consider insights fresh for 5 minutes
    refetchOnWindowFocus: false, // Don't refetch on tab focus (insights don't change that often)
  });
}

/**
 * Hook to fetch time estimate for a specific task.
 * Uses historical data to predict how long a task will take.
 */
export function useTimeEstimate(taskId: string | undefined) {
  return useQuery<TimeEstimate>({
    queryKey: ['timeEstimate', taskId],
    queryFn: async () => {
      if (!taskId) throw new Error('Task ID is required');
      try {
        const response = await insightsAPI.getTimeEstimate(taskId);
        return response.data;
      } catch (error) {
        throw new Error(getApiErrorMessage(error, 'Failed to fetch time estimate', 'useTimeEstimate'));
      }
    },
    enabled: !!taskId, // Only run query if taskId is provided
    staleTime: 10 * 60 * 1000, // Cache estimate for 10 minutes
  });
}

/**
 * Fetch category suggestions for task content.
 * This is a one-time fetch, not a reactive query, so we use a plain function.
 */
export async function suggestCategory(
  title: string,
  description?: string
): Promise<CategorySuggestionResponse> {
  try {
    const response = await insightsAPI.suggestCategory({ title, description });
    return response.data;
  } catch (error) {
    console.error('[suggestCategory]', error);
    return { suggestions: [] };
  }
}
