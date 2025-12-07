'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { Checkbox } from '@/components/ui/checkbox';
import { ChevronDown, ChevronRight, Plus, ListTodo } from 'lucide-react';
import { useSubtasks, useSubtaskInfo, useCreateSubtask, useCompleteSubtask } from '@/hooks/useSubtasks';
import { useCompleteTask } from '@/hooks/useTasks';
import { Task } from '@/lib/api';
import { tokens } from '@/lib/tokens';

interface SubtaskListProps {
  parentTaskId: string;
  parentTask: Task;
  onParentCompleted?: () => void;
}

export function SubtaskList({ parentTaskId, parentTask, onParentCompleted }: SubtaskListProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [newSubtaskTitle, setNewSubtaskTitle] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);

  const { data: subtasksData, isLoading: subtasksLoading } = useSubtasks(parentTaskId);
  const { data: subtaskInfo, isLoading: infoLoading } = useSubtaskInfo(parentTaskId);
  const createSubtask = useCreateSubtask();
  const completeSubtask = useCompleteSubtask();
  const completeParentTask = useCompleteTask();

  // Only show for regular tasks (not subtasks or recurring)
  if (parentTask.task_type !== 'regular') {
    return null;
  }

  const handleAddSubtask = () => {
    if (!newSubtaskTitle.trim()) return;

    createSubtask.mutate(
      { parentTaskId, data: { title: newSubtaskTitle.trim() } },
      {
        onSuccess: () => {
          setNewSubtaskTitle('');
          setShowAddForm(false);
          setIsExpanded(true);
        },
      }
    );
  };

  const handleCompleteSubtask = (subtaskId: string) => {
    completeSubtask.mutate(subtaskId, {
      onSuccess: (response) => {
        // If all subtasks are complete, ask about completing parent
        if (response.data.all_subtasks_complete && response.data.parent_task) {
          const shouldComplete = window.confirm(
            'All subtasks are complete! Would you like to mark the parent task as done?'
          );
          if (shouldComplete) {
            completeParentTask.mutate(parentTaskId, {
              onSuccess: () => {
                onParentCompleted?.();
              },
            });
          }
        }
      },
    });
  };

  const hasSubtasks = subtaskInfo && subtaskInfo.total_count > 0;
  const completionPercentage = subtaskInfo ? Math.round(subtaskInfo.completion_rate * 100) : 0;

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <ListTodo className="h-4 w-4" />
              Subtasks
            </CardTitle>
            {infoLoading ? (
              <Skeleton className="h-5 w-12" />
            ) : subtaskInfo && subtaskInfo.total_count > 0 ? (
              <Badge variant="outline" className="text-xs">
                {subtaskInfo.completed_count}/{subtaskInfo.total_count}
              </Badge>
            ) : null}
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
            {hasSubtasks && (
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

        {/* Progress bar */}
        {hasSubtasks && (
          <div className="mt-2">
            <div className="flex justify-between text-xs text-muted-foreground mb-1">
              <span>Progress</span>
              <span>{completionPercentage}%</span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full transition-all duration-300 rounded-full"
                style={{
                  width: `${completionPercentage}%`,
                  backgroundColor:
                    completionPercentage === 100
                      ? tokens.status.success.default
                      : completionPercentage > 50
                      ? tokens.accent.blue.default
                      : tokens.status.warning.default,
                }}
              />
            </div>
          </div>
        )}
      </CardHeader>

      <CardContent className="space-y-2">
        {/* Quick add form */}
        {showAddForm && (
          <div className="flex gap-2 pb-2">
            <Input
              placeholder="Add a subtask..."
              value={newSubtaskTitle}
              onChange={(e) => setNewSubtaskTitle(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleAddSubtask();
                if (e.key === 'Escape') {
                  setShowAddForm(false);
                  setNewSubtaskTitle('');
                }
              }}
              className="h-8 text-sm"
              autoFocus
            />
            <Button
              size="sm"
              className="h-8"
              onClick={handleAddSubtask}
              disabled={!newSubtaskTitle.trim() || createSubtask.isPending}
            >
              Add
            </Button>
          </div>
        )}

        {/* Subtasks list */}
        {subtasksLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : hasSubtasks && isExpanded ? (
          <div className="space-y-1">
            {subtasksData?.subtasks.map((subtask) => (
              <div
                key={subtask.id}
                className="flex items-center gap-2 p-2 rounded-md hover:bg-muted/50 transition-colors"
              >
                <Checkbox
                  checked={subtask.status === 'done'}
                  disabled={subtask.status === 'done' || completeSubtask.isPending}
                  onCheckedChange={() => handleCompleteSubtask(subtask.id)}
                />
                <span
                  className={`flex-1 text-sm ${
                    subtask.status === 'done'
                      ? 'line-through text-muted-foreground'
                      : ''
                  }`}
                >
                  {subtask.title}
                </span>
                {subtask.status === 'in_progress' && (
                  <Badge variant="outline" className="text-xs">
                    In Progress
                  </Badge>
                )}
              </div>
            ))}
          </div>
        ) : !hasSubtasks ? (
          <p className="text-xs text-muted-foreground text-center py-2">
            No subtasks yet. Click + to add one.
          </p>
        ) : null}

        {/* Collapsed summary */}
        {hasSubtasks && !isExpanded && (
          <button
            onClick={() => setIsExpanded(true)}
            className="text-xs text-muted-foreground hover:text-foreground transition-colors w-full text-left cursor-pointer"
          >
            Click to expand {subtaskInfo?.total_count} subtask
            {subtaskInfo?.total_count !== 1 ? 's' : ''}...
          </button>
        )}
      </CardContent>
    </Card>
  );
}
