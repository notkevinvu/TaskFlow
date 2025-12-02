'use client';

import { useRouter } from 'next/navigation';
import {
  AlertTriangle,
  TrendingUp,
  Zap,
  Calendar,
  Target,
  Layers,
  ArrowRight,
} from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { Insight, InsightType, InsightPriority } from '@/lib/api';

// Map insight types to icons and colors
const insightConfig: Record<
  InsightType,
  { icon: typeof AlertTriangle; color: string; bgColor: string }
> = {
  avoidance_pattern: {
    icon: AlertTriangle,
    color: 'text-amber-600',
    bgColor: 'bg-amber-50 dark:bg-amber-950/30',
  },
  peak_performance: {
    icon: TrendingUp,
    color: 'text-green-600',
    bgColor: 'bg-green-50 dark:bg-green-950/30',
  },
  quick_wins: {
    icon: Zap,
    color: 'text-blue-600',
    bgColor: 'bg-blue-50 dark:bg-blue-950/30',
  },
  deadline_clustering: {
    icon: Calendar,
    color: 'text-purple-600',
    bgColor: 'bg-purple-50 dark:bg-purple-950/30',
  },
  at_risk_alert: {
    icon: Target,
    color: 'text-red-600',
    bgColor: 'bg-red-50 dark:bg-red-950/30',
  },
  category_overload: {
    icon: Layers,
    color: 'text-orange-600',
    bgColor: 'bg-orange-50 dark:bg-orange-950/30',
  },
};

// Map priority to badge variant
const priorityConfig: Record<
  InsightPriority,
  { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }
> = {
  1: { label: 'Low', variant: 'outline' },
  2: { label: 'Medium', variant: 'secondary' },
  3: { label: 'High', variant: 'default' },
  4: { label: 'Urgent', variant: 'destructive' },
  5: { label: 'Critical', variant: 'destructive' },
};

interface InsightCardProps {
  insight: Insight;
}

export function InsightCard({ insight }: InsightCardProps) {
  const router = useRouter();
  const config = insightConfig[insight.type];
  const priorityInfo = priorityConfig[insight.priority];
  const Icon = config.icon;

  const handleAction = () => {
    if (insight.action_url) {
      router.push(insight.action_url);
    }
  };

  return (
    <Card className={`${config.bgColor} border-0 shadow-sm`}>
      <CardContent className="p-4">
        <div className="flex items-start gap-3">
          {/* Icon */}
          <div className={`p-2 rounded-lg bg-white dark:bg-gray-800 shadow-sm`}>
            <Icon className={`h-5 w-5 ${config.color}`} />
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="font-semibold text-sm">{insight.title}</h3>
              <Badge variant={priorityInfo.variant} className="text-xs">
                {priorityInfo.label}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground leading-relaxed">
              {insight.message}
            </p>

            {/* Action button if action_url exists */}
            {insight.action_url && (
              <Button
                variant="ghost"
                size="sm"
                className="mt-2 -ml-2 h-7 text-xs"
                onClick={handleAction}
              >
                Take action
                <ArrowRight className="h-3 w-3 ml-1" />
              </Button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
