'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  seriesAPI,
  recurrencePreferencesAPI,
  UpdateTaskSeriesDTO,
  DueDateCalculation,
  getApiErrorMessage,
} from '@/lib/api';
import { toast } from 'sonner';

// =============================================================================
// Task Series Hooks
// =============================================================================

/**
 * Fetch all task series for the current user.
 * @param activeOnly - If true, only fetch active (non-deactivated) series
 */
export function useTaskSeries(activeOnly: boolean = true) {
  return useQuery({
    queryKey: ['series', { activeOnly }],
    queryFn: async () => {
      const response = await seriesAPI.list({ active_only: activeOnly });
      return response.data;
    },
  });
}

/**
 * Fetch the history of a specific task series (all tasks in the series).
 * @param seriesId - The ID of the series to fetch history for
 */
export function useSeriesHistory(seriesId: string | null) {
  return useQuery({
    queryKey: ['series', seriesId, 'history'],
    queryFn: async () => {
      if (!seriesId) return null;
      const response = await seriesAPI.getHistory(seriesId);
      return response.data;
    },
    enabled: !!seriesId,
  });
}

/**
 * Update a task series settings (pattern, interval, end date, etc.).
 */
export function useUpdateSeries() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ seriesId, data }: { seriesId: string; data: UpdateTaskSeriesDTO }) =>
      seriesAPI.update(seriesId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['series'] });
      toast.success('Recurrence settings updated');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to update recurrence', 'Series Update'));
    },
  });
}

/**
 * Deactivate (stop) a recurring task series.
 * Future tasks will no longer be generated.
 */
export function useDeactivateSeries() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (seriesId: string) => seriesAPI.deactivate(seriesId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['series'] });
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Recurring task series stopped');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to stop recurrence', 'Series Deactivate'));
    },
  });
}

// =============================================================================
// Recurrence Preferences Hooks
// =============================================================================

/**
 * Fetch all user preferences for recurring tasks.
 * Includes default preference and category-specific preferences.
 */
export function useRecurrencePreferences() {
  return useQuery({
    queryKey: ['preferences', 'recurrence'],
    queryFn: async () => {
      const response = await recurrencePreferencesAPI.getAll();
      return response.data;
    },
  });
}

/**
 * Get the effective due date calculation mode for a specific category.
 * This considers the preference hierarchy: category > user default > system default.
 * @param category - Optional category to check (if omitted, returns user default)
 */
export function useEffectiveDueDateCalculation(category?: string) {
  return useQuery({
    queryKey: ['preferences', 'recurrence', 'effective', category],
    queryFn: async () => {
      const response = await recurrencePreferencesAPI.getEffective(category);
      return response.data;
    },
  });
}

/**
 * Set the user's default due date calculation preference.
 */
export function useSetDefaultPreference() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (dueDateCalculation: DueDateCalculation) =>
      recurrencePreferencesAPI.setDefault(dueDateCalculation),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['preferences', 'recurrence'] });
      toast.success('Default preference saved');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to save preference', 'Set Default Preference'));
    },
  });
}

/**
 * Set a category-specific due date calculation preference.
 */
export function useSetCategoryPreference() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ category, dueDateCalculation }: { category: string; dueDateCalculation: DueDateCalculation }) =>
      recurrencePreferencesAPI.setCategoryPreference(category, dueDateCalculation),
    onSuccess: (_, { category }) => {
      queryClient.invalidateQueries({ queryKey: ['preferences', 'recurrence'] });
      toast.success(`Preference saved for "${category}"`);
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to save category preference', 'Set Category Preference'));
    },
  });
}

/**
 * Delete a category-specific preference (falls back to user default).
 */
export function useDeleteCategoryPreference() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (category: string) =>
      recurrencePreferencesAPI.deleteCategoryPreference(category),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['preferences', 'recurrence'] });
      toast.success('Category preference removed');
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to remove preference', 'Delete Category Preference'));
    },
  });
}

// =============================================================================
// Helper Functions
// =============================================================================

/**
 * Format a recurrence pattern for display.
 * @param pattern - The recurrence pattern ('daily', 'weekly', 'monthly')
 * @param interval - The interval value (e.g., 2 for "every 2 weeks")
 * @returns A human-readable string like "Every 2 weeks"
 */
export function formatRecurrencePattern(pattern: string, interval: number = 1): string {
  if (pattern === 'none') return 'Does not repeat';

  const unit = pattern === 'daily' ? 'day' : pattern === 'weekly' ? 'week' : 'month';
  const plural = interval === 1 ? unit : `${unit}s`;

  return interval === 1 ? `Every ${unit}` : `Every ${interval} ${plural}`;
}

/**
 * Format due date calculation mode for display.
 * @param mode - The calculation mode ('from_original' or 'from_completion')
 * @returns A human-readable string
 */
export function formatDueDateCalculation(mode: DueDateCalculation): string {
  return mode === 'from_completion'
    ? 'Based on completion date'
    : 'Based on original due date';
}
