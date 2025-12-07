'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  gamificationAPI,
  getApiErrorMessage,
  AchievementDefinition,
  AchievementType,
} from '@/lib/api';
import { toast } from 'sonner';

// =============================================================================
// Query Keys
// =============================================================================

export const gamificationKeys = {
  all: ['gamification'] as const,
  dashboard: () => [...gamificationKeys.all, 'dashboard'] as const,
  stats: () => [...gamificationKeys.all, 'stats'] as const,
  timezone: () => [...gamificationKeys.all, 'timezone'] as const,
};

// =============================================================================
// Dashboard Hook - Full gamification data
// =============================================================================

/**
 * Fetch the full gamification dashboard including stats, achievements, and category progress.
 * Use this for the analytics page section where we need all the data.
 */
export function useGamificationDashboard() {
  return useQuery({
    queryKey: gamificationKeys.dashboard(),
    queryFn: async () => {
      const response = await gamificationAPI.getDashboard();
      return response.data;
    },
    staleTime: 30 * 1000, // 30 seconds - gamification data changes on task completion
  });
}

// =============================================================================
// Stats Hook - Lightweight stats only
// =============================================================================

/**
 * Fetch only gamification stats (streaks, productivity score, etc).
 * Use this for the sidebar widget where we don't need full achievements list.
 */
export function useGamificationStats() {
  return useQuery({
    queryKey: gamificationKeys.stats(),
    queryFn: async () => {
      const response = await gamificationAPI.getStats();
      return response.data;
    },
    staleTime: 30 * 1000, // 30 seconds
  });
}

// =============================================================================
// Timezone Hooks
// =============================================================================

/**
 * Fetch user's timezone setting for streak calculation
 */
export function useGamificationTimezone() {
  return useQuery({
    queryKey: gamificationKeys.timezone(),
    queryFn: async () => {
      const response = await gamificationAPI.getTimezone();
      return response.data.timezone;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes - timezone rarely changes
  });
}

/**
 * Update user's timezone for streak calculation
 */
export function useSetGamificationTimezone() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (timezone: string) => gamificationAPI.setTimezone(timezone),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: gamificationKeys.timezone() });
      // Also invalidate stats since streak calculation depends on timezone
      queryClient.invalidateQueries({ queryKey: gamificationKeys.stats() });
      queryClient.invalidateQueries({ queryKey: gamificationKeys.dashboard() });
      toast.success(`Timezone updated to ${response.data.timezone}`);
    },
    onError: (err: unknown) => {
      toast.error(getApiErrorMessage(err, 'Failed to update timezone', 'SetTimezone'));
    },
  });
}

// =============================================================================
// Invalidation Hook - Call after task completion
// =============================================================================

/**
 * Hook to manually invalidate gamification data.
 * Call this after task completion to refresh stats and check for new achievements.
 */
export function useInvalidateGamification() {
  const queryClient = useQueryClient();

  return () => {
    queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
  };
}

// =============================================================================
// Achievement Utilities
// =============================================================================

/**
 * Get the display icon emoji for an achievement type
 */
export function getAchievementIcon(type: AchievementType): string {
  const icons: Record<AchievementType, string> = {
    first_task: '🎯',
    milestone_10: '⭐',
    milestone_50: '🌟',
    milestone_100: '💫',
    streak_3: '🔥',
    streak_7: '🔥🔥',
    streak_14: '🔥🔥🔥',
    streak_30: '🏆',
    category_master: '🎓',
    speed_demon: '⚡',
    consistency_king: '👑',
  };
  return icons[type] || '🏅';
}

/**
 * Get achievement display title (handles dynamic category master title)
 */
export function getAchievementTitle(
  achievement: { achievement_type: AchievementType; metadata?: Record<string, unknown> },
  definitions?: AchievementDefinition[]
): string {
  // For category master, show the specific category
  if (achievement.achievement_type === 'category_master' && achievement.metadata?.category) {
    return `${achievement.metadata.category} Expert`;
  }

  // Look up in definitions if available
  if (definitions) {
    const def = definitions.find((d) => d.type === achievement.achievement_type);
    if (def) return def.title;
  }

  // Fallback titles
  const titles: Record<AchievementType, string> = {
    first_task: 'First Steps',
    milestone_10: 'Getting Started',
    milestone_50: 'Productivity Pro',
    milestone_100: 'Century Champion',
    streak_3: 'Streak Starter',
    streak_7: 'Week Warrior',
    streak_14: 'Fortnight Force',
    streak_30: 'Monthly Master',
    category_master: 'Category Master',
    speed_demon: 'Speed Demon',
    consistency_king: 'Consistency King',
  };
  return titles[achievement.achievement_type] || 'Achievement';
}

/**
 * Get achievement description
 */
export function getAchievementDescription(
  achievement: { achievement_type: AchievementType; metadata?: Record<string, unknown> },
  definitions?: AchievementDefinition[]
): string {
  // For category master, show the specific category
  if (achievement.achievement_type === 'category_master' && achievement.metadata?.category) {
    return `Complete 10 tasks in ${achievement.metadata.category}`;
  }

  // Look up in definitions if available
  if (definitions) {
    const def = definitions.find((d) => d.type === achievement.achievement_type);
    if (def) return def.description;
  }

  // Fallback descriptions
  const descriptions: Record<AchievementType, string> = {
    first_task: 'Complete your first task',
    milestone_10: 'Complete 10 tasks',
    milestone_50: 'Complete 50 tasks',
    milestone_100: 'Complete 100 tasks',
    streak_3: '3-day completion streak',
    streak_7: '7-day completion streak',
    streak_14: '14-day completion streak',
    streak_30: '30-day completion streak',
    category_master: 'Complete 10 tasks in a category',
    speed_demon: 'Complete 5 tasks within 24h of creation',
    consistency_king: 'Complete tasks on 5+ days in a week',
  };
  return descriptions[achievement.achievement_type] || 'Achievement unlocked';
}

/**
 * Get productivity score tier and color
 */
export function getProductivityTier(score: number): {
  tier: string;
  color: string;
  textColor: string;
} {
  if (score >= 90) return { tier: 'Elite', color: 'bg-purple-500', textColor: 'text-purple-500' };
  if (score >= 75) return { tier: 'Excellent', color: 'bg-emerald-500', textColor: 'text-emerald-500' };
  if (score >= 60) return { tier: 'Good', color: 'bg-blue-500', textColor: 'text-blue-500' };
  if (score >= 40) return { tier: 'Average', color: 'bg-yellow-500', textColor: 'text-yellow-500' };
  return { tier: 'Needs Work', color: 'bg-orange-500', textColor: 'text-orange-500' };
}

/**
 * Format streak display with fire emoji based on streak length
 */
export function formatStreakDisplay(streak: number): string {
  if (streak === 0) return '0';
  if (streak >= 30) return `${streak} 🏆`;
  if (streak >= 14) return `${streak} 🔥🔥🔥`;
  if (streak >= 7) return `${streak} 🔥🔥`;
  if (streak >= 3) return `${streak} 🔥`;
  return `${streak}`;
}

/**
 * Calculate progress towards next achievement milestone
 */
export function getNextMilestone(totalCompleted: number): {
  current: number;
  next: number;
  progress: number;
  label: string;
} {
  const milestones = [
    { threshold: 1, label: 'First Steps' },
    { threshold: 10, label: 'Getting Started' },
    { threshold: 50, label: 'Productivity Pro' },
    { threshold: 100, label: 'Century Champion' },
  ];

  for (const milestone of milestones) {
    if (totalCompleted < milestone.threshold) {
      return {
        current: totalCompleted,
        next: milestone.threshold,
        progress: (totalCompleted / milestone.threshold) * 100,
        label: milestone.label,
      };
    }
  }

  // All milestones achieved
  return {
    current: totalCompleted,
    next: totalCompleted,
    progress: 100,
    label: 'All Milestones Complete!',
  };
}
