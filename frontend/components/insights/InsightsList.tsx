'use client';

import { Lightbulb, RefreshCw } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { InsightCard } from './InsightCard';
import { useInsights } from '@/hooks/useInsights';
import { tokens } from '@/lib/tokens';

interface InsightsListProps {
  className?: string;
}

export function InsightsList({ className }: InsightsListProps) {
  const { data, isLoading, error, refetch, isFetching } = useInsights();

  if (isLoading) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Lightbulb className="h-5 w-5" />
            Smart Insights
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-24 w-full" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className={className}>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Lightbulb className="h-5 w-5" />
            Smart Insights
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-6 text-muted-foreground">
            <p className="text-sm">Failed to load insights</p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => refetch()}
              className="mt-2"
            >
              Try again
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  const insights = data?.insights || [];

  return (
    <Card className={className}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <Lightbulb className="h-5 w-5" style={{ color: tokens.status.warning.default }} />
            Smart Insights
          </CardTitle>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => refetch()}
            disabled={isFetching}
            className="h-8 w-8"
          >
            <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
          </Button>
        </div>
        <p className="text-sm text-muted-foreground">
          Actionable suggestions based on your task patterns
        </p>
      </CardHeader>
      <CardContent>
        {insights.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <Lightbulb className="h-12 w-12 mx-auto mb-3 opacity-20" />
            <p className="text-sm">No insights available yet</p>
            <p className="text-xs mt-1">
              Keep using TaskFlow to generate personalized suggestions
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {insights.map((insight, index) => (
              <InsightCard key={`${insight.type}-${index}`} insight={insight} />
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
