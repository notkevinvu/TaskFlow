'use client';

import { BumpAnalytics } from '@/lib/api';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { tokens } from '@/lib/tokens';

interface BumpChartProps {
  data: BumpAnalytics;
}

export function BumpChart({ data }: BumpChartProps) {
  // Format bump distribution data for chart
  const chartData = Object.entries(data.bump_distribution)
    .sort((a, b) => {
      // Sort by bump count range
      const getMin = (range: string) => {
        if (range === '0 bumps') return 0;
        if (range === '6+ bumps') return 6;
        const match = range.match(/(\d+)-(\d+)/);
        return match ? parseInt(match[1]) : 0;
      };
      return getMin(a[0]) - getMin(b[0]);
    })
    .map(([range, count]) => ({
      range,
      count,
    }));

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Bump Analysis</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center text-muted-foreground">
            No bump data available
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Bump Analysis</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Stats */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">Average Bumps</p>
              <p className="text-2xl font-bold">{data.average_bump_count.toFixed(1)}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">At Risk (3+ bumps)</p>
              <p className="text-2xl font-bold" style={{ color: tokens.status.error.default }}>{data.at_risk_count}</p>
            </div>
          </div>

          {/* Chart */}
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis
                dataKey="range"
                className="text-xs"
                tick={{ fontSize: 12 }}
              />
              <YAxis
                className="text-xs"
                tick={{ fontSize: 12 }}
                allowDecimals={false}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'hsl(var(--card))',
                  border: '1px solid hsl(var(--border))',
                  borderRadius: '6px',
                }}
                formatter={(value: number) => [`${value} tasks`, 'Count']}
              />
              <Legend />
              <Bar
                dataKey="count"
                fill={tokens.accent.blue.default}
                radius={[8, 8, 0, 0]}
                name="Task Count"
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
