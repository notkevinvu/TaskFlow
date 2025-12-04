'use client';

import { useEffect, useRef } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { PriorityBreakdown } from '@/lib/api';
import { tokens } from '@/lib/tokens';

interface PriorityBreakdownPanelProps {
  breakdown: PriorityBreakdown;
  finalScore: number;
}

// Factor labels and descriptions
const FACTOR_CONFIG = {
  user_priority: {
    label: 'Your Priority',
    description: 'The priority you set (1-10 scaled to 0-100)',
    weight: '40%',
    color: tokens.accent.blue.default,
  },
  time_decay: {
    label: 'Time Decay',
    description: 'Increases as task ages (30 days = max)',
    weight: '30%',
    color: tokens.accent.purple.default,
  },
  deadline_urgency: {
    label: 'Deadline Urgency',
    description: 'Increases as due date approaches',
    weight: '20%',
    color: tokens.accent.orange.default,
  },
  bump_penalty: {
    label: 'Bump Penalty',
    description: '+10 points per delay',
    weight: '10%',
    color: tokens.accent.pink.default,
  },
} as const;

export function PriorityBreakdownPanel({ breakdown, finalScore }: PriorityBreakdownPanelProps) {
  const hasTrackedView = useRef(false);

  // Track view event for analytics (once per mount)
  useEffect(() => {
    if (!hasTrackedView.current) {
      hasTrackedView.current = true;
      // Analytics tracking - logs to console in dev, could be sent to backend
      console.debug('[Analytics] priority_panel_view');
    }
  }, []);

  // Prepare chart data (weighted contributions)
  const chartData = [
    {
      name: 'Your Priority',
      value: Math.round(breakdown.user_priority_weighted * 10) / 10,
      color: FACTOR_CONFIG.user_priority.color,
    },
    {
      name: 'Time Decay',
      value: Math.round(breakdown.time_decay_weighted * 10) / 10,
      color: FACTOR_CONFIG.time_decay.color,
    },
    {
      name: 'Deadline Urgency',
      value: Math.round(breakdown.deadline_urgency_weighted * 10) / 10,
      color: FACTOR_CONFIG.deadline_urgency.color,
    },
    {
      name: 'Bump Penalty',
      value: Math.round(breakdown.bump_penalty_weighted * 10) / 10,
      color: FACTOR_CONFIG.bump_penalty.color,
    },
  ].filter(item => item.value > 0); // Only show factors that contribute

  // Show effort boost if not 1.0 (default)
  const hasEffortBoost = breakdown.effort_boost !== 1.0;

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-sm font-medium flex items-center justify-between">
          <span>Priority Breakdown</span>
          <span
            className="text-lg font-bold"
            style={{ color: finalScore >= 70 ? tokens.status.error.default : finalScore >= 40 ? tokens.status.warning.default : tokens.accent.blue.default }}
          >
            {finalScore}
          </span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Donut Chart */}
        {chartData.length > 0 ? (
          <div className="h-40">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={chartData}
                  cx="50%"
                  cy="50%"
                  innerRadius={35}
                  outerRadius={60}
                  paddingAngle={2}
                  dataKey="value"
                >
                  {chartData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px',
                    fontSize: '12px',
                  }}
                  formatter={(value: number) => [`${value} pts`, '']}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <div className="h-40 flex items-center justify-center text-sm text-muted-foreground">
            No priority factors active
          </div>
        )}

        {/* Detailed Breakdown Table */}
        <div className="space-y-2 text-xs">
          {/* Header */}
          <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 text-muted-foreground font-medium border-b pb-1">
            <span>Factor</span>
            <span className="text-right">Raw</span>
            <span className="text-right">Weight</span>
            <span className="text-right">Points</span>
          </div>

          {/* User Priority */}
          <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 items-center">
            <div className="flex items-center gap-2">
              <div
                className="w-2.5 h-2.5 rounded-full"
                style={{ backgroundColor: FACTOR_CONFIG.user_priority.color }}
              />
              <span title={FACTOR_CONFIG.user_priority.description}>
                {FACTOR_CONFIG.user_priority.label}
              </span>
            </div>
            <span className="text-right font-mono">{Math.round(breakdown.user_priority)}</span>
            <span className="text-right text-muted-foreground">{FACTOR_CONFIG.user_priority.weight}</span>
            <span className="text-right font-mono font-medium">
              {Math.round(breakdown.user_priority_weighted * 10) / 10}
            </span>
          </div>

          {/* Time Decay */}
          <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 items-center">
            <div className="flex items-center gap-2">
              <div
                className="w-2.5 h-2.5 rounded-full"
                style={{ backgroundColor: FACTOR_CONFIG.time_decay.color }}
              />
              <span title={FACTOR_CONFIG.time_decay.description}>
                {FACTOR_CONFIG.time_decay.label}
              </span>
            </div>
            <span className="text-right font-mono">{Math.round(breakdown.time_decay)}</span>
            <span className="text-right text-muted-foreground">{FACTOR_CONFIG.time_decay.weight}</span>
            <span className="text-right font-mono font-medium">
              {Math.round(breakdown.time_decay_weighted * 10) / 10}
            </span>
          </div>

          {/* Deadline Urgency */}
          <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 items-center">
            <div className="flex items-center gap-2">
              <div
                className="w-2.5 h-2.5 rounded-full"
                style={{ backgroundColor: FACTOR_CONFIG.deadline_urgency.color }}
              />
              <span title={FACTOR_CONFIG.deadline_urgency.description}>
                {FACTOR_CONFIG.deadline_urgency.label}
              </span>
            </div>
            <span className="text-right font-mono">{Math.round(breakdown.deadline_urgency)}</span>
            <span className="text-right text-muted-foreground">{FACTOR_CONFIG.deadline_urgency.weight}</span>
            <span className="text-right font-mono font-medium">
              {Math.round(breakdown.deadline_urgency_weighted * 10) / 10}
            </span>
          </div>

          {/* Bump Penalty */}
          <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 items-center">
            <div className="flex items-center gap-2">
              <div
                className="w-2.5 h-2.5 rounded-full"
                style={{ backgroundColor: FACTOR_CONFIG.bump_penalty.color }}
              />
              <span title={FACTOR_CONFIG.bump_penalty.description}>
                {FACTOR_CONFIG.bump_penalty.label}
              </span>
            </div>
            <span className="text-right font-mono">{Math.round(breakdown.bump_penalty)}</span>
            <span className="text-right text-muted-foreground">{FACTOR_CONFIG.bump_penalty.weight}</span>
            <span className="text-right font-mono font-medium">
              {Math.round(breakdown.bump_penalty_weighted * 10) / 10}
            </span>
          </div>

          {/* Effort Boost (if applicable) */}
          {hasEffortBoost && (
            <>
              <div className="border-t my-2" />
              <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 items-center">
                <div className="flex items-center gap-2">
                  <div
                    className="w-2.5 h-2.5 rounded-full"
                    style={{ backgroundColor: tokens.accent.green.default }}
                  />
                  <span title="Small tasks get a 1.3x boost, medium tasks get 1.15x">
                    Effort Boost
                  </span>
                </div>
                <span className="text-right font-mono">{breakdown.effort_boost.toFixed(2)}x</span>
                <span className="text-right text-muted-foreground">mult</span>
                <span className="text-right font-mono font-medium text-green-600">
                  +{Math.round((breakdown.effort_boost - 1) * 100)}%
                </span>
              </div>
            </>
          )}

          {/* Total */}
          <div className="border-t pt-2 mt-2">
            <div className="grid grid-cols-[1fr,60px,50px,60px] gap-1 items-center font-medium">
              <span>Final Score</span>
              <span></span>
              <span></span>
              <span
                className="text-right text-base"
                style={{ color: finalScore >= 70 ? tokens.status.error.default : finalScore >= 40 ? tokens.status.warning.default : tokens.text.default }}
              >
                {finalScore}
              </span>
            </div>
          </div>
        </div>

        {/* Formula explanation */}
        <p className="text-xs text-muted-foreground pt-2 border-t">
          Formula: (Priority×0.4 + Decay×0.3 + Urgency×0.2 + Penalty×0.1) × Effort
        </p>
      </CardContent>
    </Card>
  );
}
