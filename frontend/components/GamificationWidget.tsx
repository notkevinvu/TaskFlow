'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from '@/components/ui/tooltip';
import {
  useGamificationStats,
  getProductivityTier,
  formatStreakDisplay,
  getNextMilestone,
  getAchievementIcon,
} from '@/hooks/useGamification';
import { Flame, Trophy, TrendingUp, Target } from 'lucide-react';
import Link from 'next/link';

/**
 * GamificationWidget - A compact sidebar widget showing key gamification metrics
 *
 * Displays:
 * - Current streak with fire emoji scaling
 * - Productivity score with color-coded tier
 * - Progress towards next milestone
 * - Link to full analytics page for more details
 */
export function GamificationWidget() {
  const { data: stats, isLoading, error } = useGamificationStats();

  if (error) {
    // Silently fail - gamification is non-critical
    return null;
  }

  if (isLoading) {
    return (
      <Card className="border-0 shadow-none bg-transparent">
        <CardHeader className="pb-2 px-2 pt-2">
          <CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-1">
            <Trophy className="h-3 w-3" />
            Progress
          </CardTitle>
        </CardHeader>
        <CardContent className="px-2 pb-2 space-y-3">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-8 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (!stats) {
    return null;
  }

  const tier = getProductivityTier(stats.productivity_score);
  const milestone = getNextMilestone(stats.total_completed);

  return (
    <TooltipProvider>
      <div className="space-y-3">
        <div className="flex items-center justify-between px-2">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-1">
            <Trophy className="h-3 w-3" />
            Progress
          </p>
          <Link
            href="/analytics#gamification"
            className="text-xs text-primary hover:underline"
          >
            View all
          </Link>
        </div>

        {/* Main Stats Grid */}
        <div className="grid grid-cols-2 gap-2 px-2">
          {/* Streak Card */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Card className="py-2 px-3 cursor-default hover:bg-accent/50 transition-colors">
                <div className="flex items-center gap-2">
                  <div className="p-1.5 rounded-md bg-orange-100 dark:bg-orange-900/30">
                    <Flame className="h-4 w-4 text-orange-500" />
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Streak</p>
                    <p className="text-lg font-bold leading-tight">
                      {formatStreakDisplay(stats.current_streak)}
                    </p>
                  </div>
                </div>
              </Card>
            </TooltipTrigger>
            <TooltipContent side="top">
              <p>
                {stats.current_streak === 0
                  ? 'Complete a task today to start your streak!'
                  : `${stats.current_streak} day streak! Best: ${stats.longest_streak} days`}
              </p>
            </TooltipContent>
          </Tooltip>

          {/* Productivity Score Card */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Card className="py-2 px-3 cursor-default hover:bg-accent/50 transition-colors">
                <div className="flex items-center gap-2">
                  <div className={`p-1.5 rounded-md ${tier.color.replace('bg-', 'bg-').replace('-500', '-100')} dark:${tier.color.replace('bg-', 'bg-').replace('-500', '-900/30')}`}>
                    <TrendingUp className={`h-4 w-4 ${tier.textColor}`} />
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Score</p>
                    <p className="text-lg font-bold leading-tight">
                      {Math.round(stats.productivity_score)}
                    </p>
                  </div>
                </div>
              </Card>
            </TooltipTrigger>
            <TooltipContent side="top">
              <p>{tier.tier} productivity level</p>
              <p className="text-xs text-muted-foreground">
                Completion: {Math.round(stats.completion_rate)}% | On-time: {Math.round(stats.on_time_percentage)}%
              </p>
            </TooltipContent>
          </Tooltip>
        </div>

        {/* Milestone Progress */}
        {milestone.progress < 100 && (
          <div className="px-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="space-y-1.5 cursor-default">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-muted-foreground flex items-center gap-1">
                      <Target className="h-3 w-3" />
                      {milestone.label}
                    </span>
                    <span className="font-medium">
                      {milestone.current}/{milestone.next}
                    </span>
                  </div>
                  <div className="h-1.5 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full transition-all duration-500"
                      style={{ width: `${milestone.progress}%` }}
                    />
                  </div>
                </div>
              </TooltipTrigger>
              <TooltipContent side="top">
                <p>{milestone.next - milestone.current} more tasks to unlock &quot;{milestone.label}&quot;</p>
              </TooltipContent>
            </Tooltip>
          </div>
        )}

        {/* Total Completed Badge */}
        <div className="flex justify-center px-2">
          <Badge variant="secondary" className="text-xs">
            {stats.total_completed} tasks completed
          </Badge>
        </div>
      </div>
    </TooltipProvider>
  );
}
