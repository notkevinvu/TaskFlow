'use client';

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Progress } from '@/components/ui/progress';
import {
  useGamificationDashboard,
  getAchievementIcon,
  getAchievementTitle,
  getAchievementDescription,
  getProductivityTier,
  formatStreakDisplay,
} from '@/hooks/useGamification';
import { UserAchievement, AchievementDefinition, AchievementType } from '@/lib/api';
import { Trophy, Flame, Target, TrendingUp, Clock, Zap, Award, Crown } from 'lucide-react';
import { tokens } from '@/lib/tokens';

/**
 * GamificationSection - Full gamification dashboard for analytics page
 *
 * Shows:
 * - Productivity score breakdown
 * - Current/longest streak
 * - All earned achievements
 * - Available achievements to unlock
 * - Category mastery progress
 */
export function GamificationSection() {
  const { data, isLoading, error } = useGamificationDashboard();

  if (error) {
    return (
      <Card id="gamification">
        <CardContent className="pt-6">
          <p className="text-center text-muted-foreground">
            Unable to load gamification data. Please try again later.
          </p>
        </CardContent>
      </Card>
    );
  }

  if (isLoading) {
    return (
      <div id="gamification" className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Trophy className="h-5 w-5" />
              Gamification
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-4">
              {[...Array(4)].map((_, i) => (
                <Skeleton key={i} className="h-24" />
              ))}
            </div>
            <Skeleton className="h-48" />
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!data) {
    return null;
  }

  const { stats, recent_achievements, all_achievements, available_achievements, category_progress } = data;
  const tier = getProductivityTier(stats.productivity_score);

  // Determine which achievements are NOT yet earned
  const earnedTypes = new Set(all_achievements.map((a) => a.achievement_type));
  const unearnedAchievements = available_achievements.filter(
    (def) => !earnedTypes.has(def.type)
  );

  return (
    <div id="gamification" className="space-y-6">
      {/* Header Card with Productivity Score */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Trophy className="h-5 w-5 text-yellow-500" />
            Your Progress
          </CardTitle>
          <CardDescription>
            Track your achievements, streaks, and productivity metrics
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Main Stats Grid */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
            {/* Productivity Score */}
            <Card className="bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-950 dark:to-purple-900 border-purple-200 dark:border-purple-800">
              <CardContent className="pt-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-purple-500/20">
                    <TrendingUp className="h-6 w-6 text-purple-600 dark:text-purple-400" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-purple-600 dark:text-purple-400">
                      Productivity Score
                    </p>
                    <p className="text-3xl font-bold text-purple-700 dark:text-purple-300">
                      {Math.round(stats.productivity_score)}
                    </p>
                    <Badge className={`${tier.color} text-white text-xs mt-1`}>
                      {tier.tier}
                    </Badge>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Current Streak */}
            <Card className="bg-gradient-to-br from-orange-50 to-orange-100 dark:from-orange-950 dark:to-orange-900 border-orange-200 dark:border-orange-800">
              <CardContent className="pt-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-orange-500/20">
                    <Flame className="h-6 w-6 text-orange-600 dark:text-orange-400" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-orange-600 dark:text-orange-400">
                      Current Streak
                    </p>
                    <p className="text-3xl font-bold text-orange-700 dark:text-orange-300">
                      {formatStreakDisplay(stats.current_streak)}
                    </p>
                    <p className="text-xs text-orange-600/70 dark:text-orange-400/70">
                      Best: {stats.longest_streak} days
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Tasks Completed */}
            <Card className="bg-gradient-to-br from-emerald-50 to-emerald-100 dark:from-emerald-950 dark:to-emerald-900 border-emerald-200 dark:border-emerald-800">
              <CardContent className="pt-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-emerald-500/20">
                    <Target className="h-6 w-6 text-emerald-600 dark:text-emerald-400" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-emerald-600 dark:text-emerald-400">
                      Total Completed
                    </p>
                    <p className="text-3xl font-bold text-emerald-700 dark:text-emerald-300">
                      {stats.total_completed}
                    </p>
                    <p className="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                      tasks done
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Achievements Count */}
            <Card className="bg-gradient-to-br from-yellow-50 to-yellow-100 dark:from-yellow-950 dark:to-yellow-900 border-yellow-200 dark:border-yellow-800">
              <CardContent className="pt-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-yellow-500/20">
                    <Award className="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-yellow-600 dark:text-yellow-400">
                      Achievements
                    </p>
                    <p className="text-3xl font-bold text-yellow-700 dark:text-yellow-300">
                      {all_achievements.length}
                    </p>
                    <p className="text-xs text-yellow-600/70 dark:text-yellow-400/70">
                      of {available_achievements.length} unlocked
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Score Breakdown */}
          <Card className="mb-6">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Score Breakdown</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-4">
                <ScoreComponent
                  label="Completion Rate"
                  value={stats.completion_rate}
                  weight={30}
                  icon={<Target className="h-4 w-4" />}
                />
                <ScoreComponent
                  label="Streak Score"
                  value={stats.streak_score}
                  weight={25}
                  icon={<Flame className="h-4 w-4" />}
                />
                <ScoreComponent
                  label="On-Time Rate"
                  value={stats.on_time_percentage}
                  weight={25}
                  icon={<Clock className="h-4 w-4" />}
                />
                <ScoreComponent
                  label="Effort Balance"
                  value={stats.effort_mix_score}
                  weight={20}
                  icon={<Zap className="h-4 w-4" />}
                />
              </div>
            </CardContent>
          </Card>
        </CardContent>
      </Card>

      {/* Achievements Grid */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Earned Achievements */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Crown className="h-5 w-5 text-yellow-500" />
              Earned Achievements
            </CardTitle>
            <CardDescription>
              {all_achievements.length} achievements unlocked
            </CardDescription>
          </CardHeader>
          <CardContent>
            {all_achievements.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                Complete tasks to unlock achievements!
              </p>
            ) : (
              <div className="grid gap-3 max-h-[400px] overflow-y-auto pr-2">
                {all_achievements.map((achievement) => (
                  <AchievementCard
                    key={achievement.id}
                    achievement={achievement}
                    definitions={available_achievements}
                    earned
                  />
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Locked Achievements */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Trophy className="h-5 w-5 text-gray-400" />
              Available to Unlock
            </CardTitle>
            <CardDescription>
              {unearnedAchievements.length} achievements remaining
            </CardDescription>
          </CardHeader>
          <CardContent>
            {unearnedAchievements.length === 0 ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                All achievements unlocked! Amazing work!
              </p>
            ) : (
              <div className="grid gap-3 max-h-[400px] overflow-y-auto pr-2">
                {unearnedAchievements.map((def) => (
                  <LockedAchievementCard key={def.type} definition={def} />
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Category Mastery */}
      {category_progress.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Target className="h-5 w-5 text-blue-500" />
              Category Mastery
            </CardTitle>
            <CardDescription>
              Complete 10 tasks in a category to become an expert
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {category_progress.map((cat) => (
                <CategoryMasteryCard key={cat.id} mastery={cat} />
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

// Helper Components

function ScoreComponent({
  label,
  value,
  weight,
  icon,
}: {
  label: string;
  value: number;
  weight: number;
  icon: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-sm">
        <span className="flex items-center gap-1 text-muted-foreground">
          {icon}
          {label}
        </span>
        <span className="font-medium">{Math.round(value)}%</span>
      </div>
      <Progress value={value} className="h-2" />
      <p className="text-xs text-muted-foreground">
        Weight: {weight}%
      </p>
    </div>
  );
}

function AchievementCard({
  achievement,
  definitions,
  earned,
}: {
  achievement: UserAchievement;
  definitions: AchievementDefinition[];
  earned: boolean;
}) {
  const title = getAchievementTitle(achievement, definitions);
  const description = getAchievementDescription(achievement, definitions);
  const icon = getAchievementIcon(achievement.achievement_type);

  return (
    <div className="flex items-center gap-3 p-3 rounded-lg bg-gradient-to-r from-yellow-50 to-amber-50 dark:from-yellow-950/30 dark:to-amber-950/30 border border-yellow-200 dark:border-yellow-800">
      <div className="text-2xl">{icon}</div>
      <div className="flex-1 min-w-0">
        <p className="font-medium text-sm truncate">{title}</p>
        <p className="text-xs text-muted-foreground truncate">{description}</p>
      </div>
      <div className="text-xs text-muted-foreground shrink-0">
        {new Date(achievement.earned_at).toLocaleDateString()}
      </div>
    </div>
  );
}

function LockedAchievementCard({
  definition,
}: {
  definition: AchievementDefinition;
}) {
  const icon = getAchievementIcon(definition.type);

  return (
    <div className="flex items-center gap-3 p-3 rounded-lg bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-800 opacity-60">
      <div className="text-2xl grayscale">{icon}</div>
      <div className="flex-1 min-w-0">
        <p className="font-medium text-sm truncate">{definition.title}</p>
        <p className="text-xs text-muted-foreground truncate">{definition.description}</p>
      </div>
      <Badge variant="outline" className="text-xs shrink-0">
        Locked
      </Badge>
    </div>
  );
}

function CategoryMasteryCard({
  mastery,
}: {
  mastery: { category: string; completed_count: number };
}) {
  const progress = Math.min((mastery.completed_count / 10) * 100, 100);
  const isComplete = mastery.completed_count >= 10;

  return (
    <Card className={isComplete ? 'border-emerald-300 dark:border-emerald-700' : ''}>
      <CardContent className="pt-4">
        <div className="flex items-center justify-between mb-2">
          <span className="font-medium text-sm">{mastery.category}</span>
          {isComplete && (
            <Badge className="bg-emerald-500 text-white text-xs">
              Expert
            </Badge>
          )}
        </div>
        <Progress value={progress} className="h-2 mb-1" />
        <p className="text-xs text-muted-foreground">
          {mastery.completed_count}/10 tasks
        </p>
      </CardContent>
    </Card>
  );
}
