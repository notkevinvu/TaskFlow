'use client';

import { useState } from 'react';
import { Task, DueDateCalculation } from '@/lib/api';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { formatRecurrencePattern } from '@/hooks/useRecurrence';
import { CalendarClock, CalendarCheck2 } from 'lucide-react';

export interface CompletionOptions {
  dueDateCalculation: DueDateCalculation;
  saveAsDefault: boolean;
  saveForCategory: boolean;
  stopRecurrence: boolean;
  skipNextOccurrence: boolean;
}

interface RecurrenceCompletionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  task: Task;
  series?: {
    pattern: string;
    interval_value: number;
  };
  onComplete: (options: CompletionOptions) => void;
  isPending?: boolean;
}

export function RecurrenceCompletionDialog({
  open,
  onOpenChange,
  task,
  series,
  onComplete,
  isPending,
}: RecurrenceCompletionDialogProps) {
  const [dueDateCalculation, setDueDateCalculation] = useState<DueDateCalculation>('from_original');
  const [saveAsDefault, setSaveAsDefault] = useState(false);
  const [saveForCategory, setSaveForCategory] = useState(false);
  const [stopRecurrence, setStopRecurrence] = useState(false);
  const [skipNextOccurrence, setSkipNextOccurrence] = useState(false);

  const handleComplete = () => {
    onComplete({
      dueDateCalculation,
      saveAsDefault,
      saveForCategory,
      stopRecurrence,
      skipNextOccurrence,
    });
  };

  const recurrenceText = series
    ? formatRecurrencePattern(series.pattern, series.interval_value)
    : 'Recurring';

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[450px]">
        <DialogHeader>
          <DialogTitle>Complete Recurring Task</DialogTitle>
          <DialogDescription>
            &quot;{task.title}&quot; repeats {recurrenceText.toLowerCase()}.
            How should the next due date be calculated?
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Due Date Calculation Options */}
          <RadioGroup
            value={dueDateCalculation}
            onValueChange={(val) => setDueDateCalculation(val as DueDateCalculation)}
            className="space-y-3"
          >
            <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors">
              <RadioGroupItem value="from_original" id="from_original" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="from_original" className="flex items-center gap-2 cursor-pointer">
                  <CalendarClock className="h-4 w-4 text-muted-foreground" />
                  Based on original due date
                </Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Next due date is calculated from the current due date, regardless of when you complete it.
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors">
              <RadioGroupItem value="from_completion" id="from_completion" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="from_completion" className="flex items-center gap-2 cursor-pointer">
                  <CalendarCheck2 className="h-4 w-4 text-muted-foreground" />
                  Based on completion date
                </Label>
                <p className="text-sm text-muted-foreground mt-1">
                  Next due date is calculated from today. Use this if you completed the task late.
                </p>
              </div>
            </div>
          </RadioGroup>

          {/* Save Preference Options */}
          <div className="space-y-3 border-t pt-4">
            <p className="text-sm font-medium">Save this preference?</p>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="save_default"
                  checked={saveAsDefault}
                  onCheckedChange={(checked) => {
                    setSaveAsDefault(checked === true);
                    if (checked) setSaveForCategory(false);
                  }}
                />
                <Label htmlFor="save_default" className="text-sm cursor-pointer">
                  Set as my default for all recurring tasks
                </Label>
              </div>
              {task.category && (
                <div className="flex items-center space-x-2">
                  <Checkbox
                    id="save_category"
                    checked={saveForCategory}
                    onCheckedChange={(checked) => {
                      setSaveForCategory(checked === true);
                      if (checked) setSaveAsDefault(false);
                    }}
                  />
                  <Label htmlFor="save_category" className="text-sm cursor-pointer">
                    Set as default for &quot;{task.category}&quot; category
                  </Label>
                </div>
              )}
            </div>
          </div>

          {/* Additional Options */}
          <div className="space-y-3 border-t pt-4">
            <p className="text-sm font-medium">Additional options</p>
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="skip_next"
                  checked={skipNextOccurrence}
                  onCheckedChange={(checked) => {
                    setSkipNextOccurrence(checked === true);
                    if (checked) setStopRecurrence(false);
                  }}
                  disabled={stopRecurrence}
                />
                <Label htmlFor="skip_next" className="text-sm cursor-pointer">
                  Skip the next occurrence (one-time)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <Checkbox
                  id="stop_recurrence"
                  checked={stopRecurrence}
                  onCheckedChange={(checked) => {
                    setStopRecurrence(checked === true);
                    if (checked) setSkipNextOccurrence(false);
                  }}
                />
                <Label htmlFor="stop_recurrence" className="text-sm cursor-pointer text-destructive">
                  Stop this recurring task entirely
                </Label>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button onClick={handleComplete} disabled={isPending}>
            {isPending ? 'Completing...' : 'Complete Task'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
