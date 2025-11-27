import { useQuery } from '@tanstack/react-query';
import { analyticsAPI } from '@/lib/api';

export function useAnalyticsSummary(days: number = 30) {
  return useQuery({
    queryKey: ['analytics', 'summary', days],
    queryFn: async () => {
      console.log('[useAnalyticsSummary] Fetching summary for days:', days);
      try {
        const response = await analyticsAPI.getSummary({ days });
        console.log('[useAnalyticsSummary] Response:', response.data);
        return response.data;
      } catch (error) {
        console.error('[useAnalyticsSummary] Error:', error);
        throw error;
      }
    },
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
  });
}

export function useAnalyticsTrends(days: number = 30) {
  return useQuery({
    queryKey: ['analytics', 'trends', days],
    queryFn: async () => {
      console.log('[useAnalyticsTrends] Fetching trends for days:', days);
      try {
        const response = await analyticsAPI.getTrends({ days });
        console.log('[useAnalyticsTrends] Response:', response.data);
        return response.data;
      } catch (error) {
        console.error('[useAnalyticsTrends] Error:', error);
        throw error;
      }
    },
    staleTime: 5 * 60 * 1000, // Consider data fresh for 5 minutes
  });
}
