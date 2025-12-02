'use client';

import { Task } from '@/lib/api';
import { Badge } from '@/components/ui/badge';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { format } from 'date-fns';

interface CalendarTaskPopoverProps {
  date: Date;
  tasks: Task[];
  trigger: React.ReactNode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTaskClick: (taskId: string) => void;
}

export function CalendarTaskPopover({
  date,
  tasks,
  trigger,
  open,
  onOpenChange,
  onTaskClick,
}: CalendarTaskPopoverProps) {
  return (
    <Popover open={open} onOpenChange={onOpenChange}>
      <PopoverTrigger asChild>
        {trigger}
      </PopoverTrigger>
      <PopoverContent
        className="w-80 p-3"
        align="start"
        side="right"
        sideOffset={8}
        collisionPadding={16}
        avoidCollisions={true}
      >
        <div className="space-y-3">
          {/* Header */}
          <div className="border-b pb-2">
            <h4 className="font-semibold text-sm">
              {format(date, 'EEEE, MMM d, yyyy')}
            </h4>
            <p className="text-xs text-muted-foreground">
              {tasks.length} {tasks.length === 1 ? 'task' : 'tasks'} due
            </p>
          </div>

          {/* Task List */}
          <div className="space-y-2 max-h-96 overflow-y-auto">
            {tasks.map((task) => (
              <div
                key={task.id}
                className="p-2 rounded-md border hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer transition-colors"
                onClick={() => {
                  onTaskClick(task.id);
                  onOpenChange(false);
                }}
              >
                <div className="flex items-start gap-2">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-start gap-1.5 mb-1">
                      <p className="text-sm font-medium line-clamp-2 flex-1">
                        {task.title}
                      </p>
                      <Badge
                        variant={
                          task.priority_score >= 90
                            ? 'destructive'
                            : task.priority_score >= 75
                            ? 'default'
                            : 'secondary'
                        }
                        className="text-[10px] h-4 px-1 flex-shrink-0"
                      >
                        {Math.round(task.priority_score)}
                      </Badge>
                    </div>
                    {task.description && (
                      <p className="text-xs text-muted-foreground line-clamp-2">
                        {task.description}
                      </p>
                    )}
                    {task.category && (
                      <p className="text-xs text-muted-foreground mt-1">
                        📁 {task.category}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
