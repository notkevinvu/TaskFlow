import { useQuery } from '@tanstack/react-query';
import { analyticsAPI, getApiErrorMessage } from '@/lib/api';

export function useAnalyticsSummary(days: number = 30) {
  return useQuery({
    queryKey: ['analytics', 'summary', days],
    queryFn: async () => {
      try {
        const response = await analyticsAPI.getSummary({ days });
        return response.data;
      } catch (error) {
        throw new Error(getApiErrorMessage(error, 'Failed to fetch analytics summary', 'useAnalyticsSummary'));
      }
    },
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
  });
}

export function useAnalyticsTrends(days: number = 30) {
  return useQuery({
    queryKey: ['analytics', 'trends', days],
    queryFn: async () => {
      try {
        const response = await analyticsAPI.getTrends({ days });
        return response.data;
      } catch (error) {
        throw new Error(getApiErrorMessage(error, 'Failed to fetch analytics trends', 'useAnalyticsTrends'));
      }
    },
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
  });
}

export function useProductivityHeatmap(days: number = 90) {
  return useQuery({
    queryKey: ['analytics', 'heatmap', days],
    queryFn: async () => {
      try {
        const response = await analyticsAPI.getHeatmap({ days });
        return response.data;
      } catch (error) {
        throw new Error(getApiErrorMessage(error, 'Failed to fetch productivity heatmap', 'useProductivityHeatmap'));
      }
    },
    staleTime: 10 * 60 * 1000, // Cache for 10 minutes (heatmap data changes slowly)
  });
}

export function useCategoryTrends(days: number = 90) {
  return useQuery({
    queryKey: ['analytics', 'category-trends', days],
    queryFn: async () => {
      try {
        const response = await analyticsAPI.getCategoryTrends({ days });
        return response.data;
      } catch (error) {
        throw new Error(getApiErrorMessage(error, 'Failed to fetch category trends', 'useCategoryTrends'));
      }
    },
    staleTime: 10 * 60 * 1000, // Cache for 10 minutes
  });
}
