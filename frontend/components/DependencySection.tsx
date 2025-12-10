'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { ChevronDown, ChevronRight, Plus, Lock, X, ArrowRight, Link2, Pause, Ban } from 'lucide-react';
import { useDependencyInfo, useAddBlocker, useRemoveBlocker } from '@/hooks/useDependencies';
import { useTasks } from '@/hooks/useTasks';
import { Task, DependencyWithTask } from '@/lib/api';
import { tokens } from '@/lib/tokens';

interface DependencySectionProps {
  taskId: string;
  task: Task;
}

export function DependencySection({ taskId, task }: DependencySectionProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [selectedBlockerId, setSelectedBlockerId] = useState<string>('');

  const { data: dependencyInfo, isLoading } = useDependencyInfo(taskId);
  const { data: allTasks } = useTasks();
  const addBlocker = useAddBlocker();
  const removeBlocker = useRemoveBlocker();

  // Only show for regular tasks (not subtasks or recurring)
  if (task.task_type !== 'regular') {
    return null;
  }

  const handleAddBlocker = () => {
    if (!selectedBlockerId) return;

    addBlocker.mutate(
      { taskId, blockedById: selectedBlockerId },
      {
        onSuccess: () => {
          setSelectedBlockerId('');
          setShowAddForm(false);
          setIsExpanded(true);
        },
      }
    );
  };

  const handleRemoveBlocker = (blockedById: string) => {
    removeBlocker.mutate({ taskId, blockedById });
  };

  const hasBlockers = dependencyInfo && dependencyInfo.blockers.length > 0;
  const hasBlocking = dependencyInfo && dependencyInfo.blocking.length > 0;
  const hasDependencies = hasBlockers || hasBlocking;
  const isBlocked = dependencyInfo?.is_blocked ?? false;

  // Get available tasks for the blocker dropdown
  // Filter out: current task, tasks already blocking this one, subtasks, and done tasks
  const availableTasks =
    allTasks?.tasks?.filter((t: Task) => {
      if (t.id === taskId) return false; // Can't block self
      if (t.task_type !== 'regular') return false; // Only regular tasks
      if (t.status === 'done') return false; // No completed tasks
      if (dependencyInfo?.blockers.some((b) => b.task_id === t.id)) return false; // Already blocking
      return true;
    }) ?? [];

  return (
    <Card className={isBlocked ? 'border-amber-200 bg-amber-50/30 dark:border-amber-800 dark:bg-amber-950/20' : ''}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Link2 className="h-4 w-4" />
              Dependencies
            </CardTitle>
            {isLoading ? (
              <Skeleton className="h-5 w-12" />
            ) : (
              <>
                {hasBlockers && (
                  <Badge
                    variant={isBlocked ? 'default' : 'outline'}
                    className={`text-xs ${isBlocked ? 'bg-amber-500 hover:bg-amber-600' : ''}`}
                  >
                    <Lock className="h-3 w-3 mr-1" />
                    {dependencyInfo.blockers.filter((b) => b.status !== 'done').length} blocking
                  </Badge>
                )}
                {hasBlocking && (
                  <Badge variant="outline" className="text-xs">
                    {dependencyInfo.blocking.length} blocked by this
                  </Badge>
                )}
              </>
            )}
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0"
              onClick={() => setShowAddForm(!showAddForm)}
            >
              <Plus className="h-4 w-4" />
            </Button>
            {hasDependencies && (
              <Button
                variant="ghost"
                size="sm"
                className="h-7 w-7 p-0"
                onClick={() => setIsExpanded(!isExpanded)}
              >
                {isExpanded ? (
                  <ChevronDown className="h-4 w-4" />
                ) : (
                  <ChevronRight className="h-4 w-4" />
                )}
              </Button>
            )}
          </div>
        </div>

        {/* Blocked warning */}
        {isBlocked && (
          <p className="text-xs text-amber-600 dark:text-amber-400 mt-2">
            This task cannot be completed until all blockers are resolved.
          </p>
        )}
      </CardHeader>

      <CardContent className="space-y-2">
        {/* Add blocker form */}
        {showAddForm && (
          <div className="flex gap-2 pb-2">
            <Select value={selectedBlockerId} onValueChange={setSelectedBlockerId}>
              <SelectTrigger className="h-8 text-sm flex-1">
                <SelectValue placeholder="Select a task to block this one..." />
              </SelectTrigger>
              <SelectContent>
                {availableTasks.length === 0 ? (
                  <SelectItem value="none" disabled>
                    No available tasks
                  </SelectItem>
                ) : (
                  availableTasks.map((t) => (
                    <SelectItem key={t.id} value={t.id}>
                      <span className="truncate">{t.title}</span>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            <Button
              size="sm"
              className="h-8"
              onClick={handleAddBlocker}
              disabled={!selectedBlockerId || addBlocker.isPending}
            >
              Add
            </Button>
          </div>
        )}

        {/* Dependencies list */}
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : hasDependencies && isExpanded ? (
          <div className="space-y-3">
            {/* Blockers section */}
            {hasBlockers && (
              <div>
                <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1">
                  <Lock className="h-3 w-3" /> Blocked by:
                </p>
                <div className="space-y-1">
                  {dependencyInfo.blockers.map((blocker) => (
                    <DependencyItem
                      key={blocker.task_id}
                      dependency={blocker}
                      onRemove={() => handleRemoveBlocker(blocker.task_id)}
                      isPending={removeBlocker.isPending}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Blocking section */}
            {hasBlocking && (
              <div>
                <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1">
                  <ArrowRight className="h-3 w-3" /> Blocking:
                </p>
                <div className="space-y-1">
                  {dependencyInfo.blocking.map((blocked) => (
                    <div
                      key={blocked.task_id}
                      className="flex items-center gap-2 p-2 rounded-md bg-muted/30"
                    >
                      <span className="flex-1 text-sm truncate">{blocked.title}</span>
                      <StatusBadge status={blocked.status} />
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : !hasDependencies ? (
          <p className="text-xs text-muted-foreground text-center py-2">
            No dependencies. Click + to add a blocker.
          </p>
        ) : null}

        {/* Collapsed summary */}
        {hasDependencies && !isExpanded && (
          <button
            onClick={() => setIsExpanded(true)}
            className="text-xs text-muted-foreground hover:text-foreground transition-colors w-full text-left cursor-pointer"
          >
            Click to expand dependencies...
          </button>
        )}
      </CardContent>
    </Card>
  );
}

// Helper component for dependency items with remove button
function DependencyItem({
  dependency,
  onRemove,
  isPending,
}: {
  dependency: DependencyWithTask;
  onRemove: () => void;
  isPending: boolean;
}) {
  const isDone = dependency.status === 'done';

  return (
    <div
      className={`flex items-center gap-2 p-2 rounded-md transition-colors ${
        isDone ? 'bg-green-50 dark:bg-green-950/30' : 'bg-amber-50 dark:bg-amber-950/30'
      }`}
    >
      <span
        className={`flex-1 text-sm truncate ${
          isDone ? 'line-through text-muted-foreground' : ''
        }`}
      >
        {dependency.title}
      </span>
      <StatusBadge status={dependency.status} />
      <Button
        variant="ghost"
        size="sm"
        className="h-6 w-6 p-0"
        onClick={onRemove}
        disabled={isPending}
      >
        <X className="h-3 w-3" />
      </Button>
    </div>
  );
}

// Helper component for status badges
function StatusBadge({ status }: { status: 'todo' | 'in_progress' | 'done' | 'on_hold' | 'blocked' }) {
  if (status === 'done') {
    return (
      <Badge
        variant="outline"
        className="text-xs"
        style={{ borderColor: tokens.status.success.default, color: tokens.status.success.default }}
      >
        Done
      </Badge>
    );
  }
  if (status === 'in_progress') {
    return (
      <Badge variant="outline" className="text-xs">
        In Progress
      </Badge>
    );
  }
  if (status === 'on_hold') {
    return (
      <Badge
        variant="outline"
        className="text-xs text-purple-600 dark:text-purple-400 border-purple-300 dark:border-purple-600"
      >
        <Pause className="h-2.5 w-2.5 mr-1" />
        On Hold
      </Badge>
    );
  }
  if (status === 'blocked') {
    return (
      <Badge
        variant="outline"
        className="text-xs"
        style={{ borderColor: tokens.status.error.default, color: tokens.status.error.default }}
      >
        <Ban className="h-2.5 w-2.5 mr-1" />
        Blocked
      </Badge>
    );
  }
  return (
    <Badge
      variant="outline"
      className="text-xs"
      style={{ borderColor: tokens.status.warning.default, color: tokens.status.warning.default }}
    >
      Todo
    </Badge>
  );
}

// Export a helper to check if a task is blocked (for use in task cards)
export function isTaskBlocked(task: Task): boolean {
  // This is a simple check - the actual blocked status comes from the API
  // This is used for optimistic rendering before API data loads
  return false;
}
