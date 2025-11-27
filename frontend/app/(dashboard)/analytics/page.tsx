'use client';

import { useState } from 'react';
import { useAnalyticsSummary, useAnalyticsTrends } from '@/hooks/useAnalytics';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CompletionChart } from '@/components/charts/CompletionChart';
import { CategoryChart } from '@/components/charts/CategoryChart';
import { PriorityChart } from '@/components/charts/PriorityChart';
import { BumpChart } from '@/components/charts/BumpChart';

export default function AnalyticsPage() {
  const [timePeriod, setTimePeriod] = useState<number>(30);

  const { data: summary, isLoading: summaryLoading, error: summaryError } = useAnalyticsSummary(timePeriod);
  const { data: trends, isLoading: trendsLoading, error: trendsError } = useAnalyticsTrends(timePeriod);

  const isLoading = summaryLoading || trendsLoading;

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

  // Show error state with details
  if (summaryError || trendsError) {
    const error = summaryError || trendsError;
    let errorMessage = '';
    let backendDetails = '';

    if (error) {
      errorMessage = error instanceof Error ? error.message : JSON.stringify(error);

      // Try to extract backend error details
      if ('response' in error && error.response && typeof error.response === 'object') {
        const response = error.response as any;
        if (response.data) {
          backendDetails = JSON.stringify(response.data, null, 2);
        }
      }
    }

    return (
      <div className="space-y-6">
        <h2 className="text-3xl font-bold">Analytics</h2>
        <Card>
          <CardContent className="pt-6">
            <div className="text-center space-y-2">
              <p className="text-muted-foreground">
                Unable to load analytics data. Please try again later.
              </p>
              <details className="text-left text-xs text-red-600 mt-2">
                <summary className="cursor-pointer">Error details</summary>
                <div className="mt-2 space-y-2">
                  <div>
                    <strong>Error:</strong>
                    <pre className="mt-1 p-2 bg-red-50 dark:bg-red-950 rounded overflow-auto text-xs">
                      {errorMessage}
                    </pre>
                  </div>
                  {backendDetails && (
                    <div>
                      <strong>Backend response:</strong>
                      <pre className="mt-1 p-2 bg-red-50 dark:bg-red-950 rounded overflow-auto text-xs">
                        {backendDetails}
                      </pre>
                    </div>
                  )}
                </div>
              </details>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!summary || !trends) {
    return (
      <div className="space-y-6">
        <h2 className="text-3xl font-bold">Analytics</h2>
        <Card>
          <CardContent className="pt-6">
            <p className="text-center text-muted-foreground">
              Loading analytics data...
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const completionStats = summary.completion_stats;
  const bumpAnalytics = summary.bump_analytics;
  const categoryBreakdown = summary.category_breakdown;
  const priorityDistribution = summary.priority_distribution;
  const velocityMetrics = trends.velocity_metrics;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold">Analytics</h2>
          <p className="text-muted-foreground">
            Insights into your task patterns and productivity
          </p>
        </div>

        {/* Time Period Selector */}
        <Select
          value={timePeriod.toString()}
          onValueChange={(value) => setTimePeriod(parseInt(value))}
        >
          <SelectTrigger className="w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="7">Last 7 days</SelectItem>
            <SelectItem value="14">Last 14 days</SelectItem>
            <SelectItem value="30">Last 30 days</SelectItem>
            <SelectItem value="60">Last 60 days</SelectItem>
            <SelectItem value="90">Last 90 days</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Summary Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Tasks
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{completionStats.total_tasks}</div>
            <p className="text-xs text-muted-foreground">
              In selected period
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Completed
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {completionStats.completed_tasks}
            </div>
            <p className="text-xs text-muted-foreground">
              {completionStats.completion_rate.toFixed(1)}% completion rate
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Pending
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">
              {completionStats.pending_tasks}
            </div>
            <p className="text-xs text-muted-foreground">
              Still in progress
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              At Risk
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {bumpAnalytics.at_risk_count}
            </div>
            <p className="text-xs text-muted-foreground">
              Avg bumps: {bumpAnalytics.average_bump_count.toFixed(1)}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts Grid */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Completion Trends */}
        <div className="lg:col-span-2">
          <CompletionChart data={velocityMetrics} />
        </div>

        {/* Priority Distribution */}
        <PriorityChart data={priorityDistribution} />

        {/* Category Breakdown */}
        <CategoryChart data={categoryBreakdown} />

        {/* Bump Analysis */}
        <div className="lg:col-span-2">
          <BumpChart data={bumpAnalytics} />
        </div>
      </div>
    </div>
  );
}
