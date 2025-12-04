'use client';

import { CategoryStat } from '@/lib/api';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { tokens, ACCENT_COLORS } from '@/lib/tokens';

interface CategoryChartProps {
  data: CategoryStat[];
}

// Use accent tokens for consistent category colors
const COLORS = ACCENT_COLORS.slice(0, 5).map(
  (color) => tokens.accent[color].default
);

export function CategoryChart({ data }: CategoryChartProps) {
  // Format data for chart
  const chartData = data.map((d, index) => ({
    name: d.category || 'Uncategorized',
    value: d.task_count,
    completionRate: Math.round(d.completion_rate * 100),
    fill: COLORS[index % COLORS.length],
  }));

  if (chartData.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Category Breakdown</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center text-muted-foreground">
            No category data available
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Category Breakdown</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ name, percent }) => `${name}: ${((percent || 0) * 100).toFixed(0)}%`}
              outerRadius={100}
              fill={tokens.accent.purple.default}
              dataKey="value"
            >
              {chartData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.fill} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                backgroundColor: 'hsl(var(--card))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '6px',
              }}
              formatter={(value: number, _name: string, props: { payload?: { completionRate?: number; name?: string } }) => [
                `${value} tasks (${props.payload?.completionRate ?? 0}% complete)`,
                props.payload?.name ?? '',
              ]}
            />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
