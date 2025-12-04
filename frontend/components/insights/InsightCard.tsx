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
import { tokens } from '@/lib/tokens';

// Map insight types to icons and token-based colors
const insightConfig: Record<
  InsightType,
  { icon: typeof AlertTriangle; colorToken: string; bgToken: string }
> = {
  avoidance_pattern: {
    icon: AlertTriangle,
    colorToken: tokens.status.warning.default,
    bgToken: tokens.status.warning.muted,
  },
  peak_performance: {
    icon: TrendingUp,
    colorToken: tokens.status.success.default,
    bgToken: tokens.status.success.muted,
  },
  quick_wins: {
    icon: Zap,
    colorToken: tokens.status.info.default,
    bgToken: tokens.status.info.muted,
  },
  deadline_clustering: {
    icon: Calendar,
    colorToken: tokens.accent.purple.default,
    bgToken: tokens.accent.purple.muted,
  },
  at_risk_alert: {
    icon: Target,
    colorToken: tokens.status.error.default,
    bgToken: tokens.status.error.muted,
  },
  category_overload: {
    icon: Layers,
    colorToken: tokens.accent.orange.default,
    bgToken: tokens.accent.orange.muted,
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
    <Card className="border-0 shadow-sm" style={{ backgroundColor: config.bgToken }}>
      <CardContent className="p-4">
        <div className="flex items-start gap-3">
          {/* Icon */}
          <div className="p-2 rounded-lg shadow-sm" style={{ backgroundColor: tokens.surface.elevated }}>
            <Icon className="h-5 w-5" style={{ color: config.colorToken }} />
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
