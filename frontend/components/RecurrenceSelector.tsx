'use client';

import { RecurrencePattern, RecurrenceRule } from '@/lib/api';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Repeat } from 'lucide-react';

interface RecurrenceSelectorProps {
  value: RecurrenceRule | null;
  onChange: (value: RecurrenceRule | null) => void;
  showEndDate?: boolean;
}

const PATTERN_OPTIONS: { value: RecurrencePattern; label: string }[] = [
  { value: 'none', label: 'Does not repeat' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

export function RecurrenceSelector({ value, onChange, showEndDate = false }: RecurrenceSelectorProps) {
  const pattern = value?.pattern ?? 'none';
  const intervalValue = value?.interval_value ?? 1;
  const endDate = value?.end_date ?? '';

  const handlePatternChange = (newPattern: RecurrencePattern) => {
    if (newPattern === 'none') {
      onChange(null);
    } else {
      onChange({
        pattern: newPattern,
        interval_value: intervalValue,
        end_date: endDate || undefined,
      });
    }
  };

  const handleIntervalChange = (newInterval: number) => {
    if (pattern === 'none') return;
    onChange({
      pattern,
      interval_value: Math.max(1, Math.min(365, newInterval)),
      end_date: endDate || undefined,
    });
  };

  const handleEndDateChange = (newEndDate: string) => {
    if (pattern === 'none') return;
    onChange({
      pattern,
      interval_value: intervalValue,
      end_date: newEndDate || undefined,
    });
  };

  const getIntervalLabel = () => {
    switch (pattern) {
      case 'daily': return intervalValue === 1 ? 'day' : 'days';
      case 'weekly': return intervalValue === 1 ? 'week' : 'weeks';
      case 'monthly': return intervalValue === 1 ? 'month' : 'months';
      default: return '';
    }
  };

  return (
    <div className="space-y-3">
      {/* Pattern Selection */}
      <div className="grid gap-2">
        <Label htmlFor="recurrence" className="flex items-center gap-2">
          <Repeat className="h-4 w-4" />
          Repeat
        </Label>
        <Select
          value={pattern}
          onValueChange={(val) => handlePatternChange(val as RecurrencePattern)}
        >
          <SelectTrigger id="recurrence">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {PATTERN_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Interval Selection (shown only when repeating) */}
      {pattern !== 'none' && (
        <div className="grid gap-2">
          <Label htmlFor="interval">Repeat every</Label>
          <div className="flex items-center gap-2">
            <Input
              id="interval"
              type="number"
              min={1}
              max={365}
              value={intervalValue}
              onChange={(e) => handleIntervalChange(parseInt(e.target.value) || 1)}
              className="w-20"
            />
            <span className="text-sm text-muted-foreground">{getIntervalLabel()}</span>
          </div>
        </div>
      )}

      {/* End Date (optional, shown only when repeating) */}
      {pattern !== 'none' && showEndDate && (
        <div className="grid gap-2">
          <Label htmlFor="recurrence_end_date">Ends on (optional)</Label>
          <Input
            id="recurrence_end_date"
            type="date"
            value={endDate}
            onChange={(e) => handleEndDateChange(e.target.value)}
          />
        </div>
      )}
    </div>
  );
}
