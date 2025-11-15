'use client';

import { useTasks } from '@/hooks/useTasks';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

export default function AnalyticsPage() {
  const { data: tasksData, isLoading } = useTasks();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h2 className="text-3xl font-bold">Analytics</h2>
        <div className="grid gap-6">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-64" />
          ))}
        </div>
      </div>
    );
  }

  const tasks = tasksData?.tasks || [];

  // Delay Analysis
  const bumpedTasks = tasks.filter(t => t.bump_count > 0);
  const avgBumpCount = bumpedTasks.length > 0
    ? bumpedTasks.reduce((sum, t) => sum + t.bump_count, 0) / bumpedTasks.length
    : 0;

  const bumpDistribution = {
    '0 bumps': tasks.filter(t => t.bump_count === 0).length,
    '1-2 bumps': tasks.filter(t => t.bump_count >= 1 && t.bump_count <= 2).length,
    '3-5 bumps': tasks.filter(t => t.bump_count >= 3 && t.bump_count <= 5).length,
    '6+ bumps': tasks.filter(t => t.bump_count >= 6).length,
  };

  // Category Breakdown
  const categories: Record<string, { count: number; avgBumps: number }> = {};
  tasks.forEach(task => {
    const cat = task.category || 'Uncategorized';
    if (!categories[cat]) {
      categories[cat] = { count: 0, avgBumps: 0 };
    }
    categories[cat].count++;
    categories[cat].avgBumps += task.bump_count;
  });

  Object.keys(categories).forEach(cat => {
    categories[cat].avgBumps = categories[cat].avgBumps / categories[cat].count;
  });

  const sortedCategories = Object.entries(categories).sort(
    (a, b) => b[1].count - a[1].count
  );

  // Effort Distribution
  const effortDistribution = {
    small: tasks.filter(t => t.estimated_effort === 'small').length,
    medium: tasks.filter(t => t.estimated_effort === 'medium').length,
    large: tasks.filter(t => t.estimated_effort === 'large').length,
    xlarge: tasks.filter(t => t.estimated_effort === 'xlarge').length,
    unknown: tasks.filter(t => !t.estimated_effort).length,
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold">Analytics</h2>
        <p className="text-muted-foreground">
          Insights into your task patterns and productivity
        </p>
      </div>

      {/* Delay Analysis */}
      <Card>
        <CardHeader>
          <CardTitle>Delay Analysis</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">Tasks Delayed</p>
              <p className="text-2xl font-bold">{bumpedTasks.length}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Average Bumps</p>
              <p className="text-2xl font-bold">{avgBumpCount.toFixed(1)}</p>
            </div>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium">Bump Distribution</p>
            {Object.entries(bumpDistribution).map(([range, count]) => (
              <div key={range} className="flex items-center justify-between">
                <span className="text-sm">{range}</span>
                <div className="flex items-center gap-2">
                  <div className="w-64 bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full"
                      style={{
                        width: `${(count / tasks.length) * 100}%`
                      }}
                    />
                  </div>
                  <span className="text-sm font-medium w-8 text-right">{count}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Category Breakdown */}
      <Card>
        <CardHeader>
          <CardTitle>Category Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {sortedCategories.map(([category, stats]) => (
              <div key={category} className="flex items-center justify-between pb-3 border-b last:border-0">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{category}</p>
                    <Badge variant="outline">{stats.count} tasks</Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">
                    Avg bumps: {stats.avgBumps.toFixed(1)}
                  </p>
                </div>
                <div className="w-48 bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-primary h-2 rounded-full"
                    style={{
                      width: `${(stats.count / tasks.length) * 100}%`
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Effort Distribution */}
      <Card>
        <CardHeader>
          <CardTitle>Effort Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Object.entries(effortDistribution).map(([effort, count]) => (
              <div key={effort} className="flex items-center justify-between">
                <span className="text-sm capitalize">{effort}</span>
                <div className="flex items-center gap-2">
                  <div className="w-64 bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full"
                      style={{
                        width: `${tasks.length > 0 ? (count / tasks.length) * 100 : 0}%`
                      }}
                    />
                  </div>
                  <span className="text-sm font-medium w-8 text-right">{count}</span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Velocity (Simple version for now) */}
      <Card>
        <CardHeader>
          <CardTitle>Velocity & Completion</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">Total Tasks</p>
              <p className="text-2xl font-bold">{tasks.length}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">High Priority (75+)</p>
              <p className="text-2xl font-bold">
                {tasks.filter(t => t.priority_score >= 75).length}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Critical (90+)</p>
              <p className="text-2xl font-bold text-red-600">
                {tasks.filter(t => t.priority_score >= 90).length}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
