'use client';

import { useEffect, useState } from 'react';
import { useTask, useBumpTask, useCompleteTask, useDeleteTask } from '@/hooks/useTasks';
import { useCanCompleteParent } from '@/hooks/useSubtasks';
import { useCanCompleteDependencies } from '@/hooks/useDependencies';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { X, Pencil, Trash2, Repeat, ListChecks, Lock } from 'lucide-react';
import { EditTaskDialog } from '@/components/EditTaskDialog';
import { PriorityBreakdownPanel } from '@/components/PriorityBreakdownPanel';
import { SubtaskList } from '@/components/SubtaskList';
import { DependencySection } from '@/components/DependencySection';
import { tokens } from '@/lib/tokens';

interface TaskDetailsSidebarProps {
  taskId: string;
  onClose: () => void;
}

export function TaskDetailsSidebar({ taskId, onClose }: TaskDetailsSidebarProps) {
  const { data: task, isLoading } = useTask(taskId);
  const [isVisible, setIsVisible] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();
  const deleteTask = useDeleteTask();

  // Check if this task can be completed (all subtasks done)
  const { data: canCompleteSubtasks } = useCanCompleteParent(taskId);
  // Check if this task can be completed (no incomplete blockers)
  const { data: dependencyStatus } = useCanCompleteDependencies(taskId);
  const isRegularTask = task?.task_type === 'regular';
  const hasSubtaskBlocker = isRegularTask && canCompleteSubtasks === false;
  const hasDependencyBlocker = isRegularTask && dependencyStatus?.is_blocked === true;
  const hasAnyBlocker = hasSubtaskBlocker || hasDependencyBlocker;

  useEffect(() => {
    // Trigger slide-in animation after component mounts
    const timer = setTimeout(() => setIsVisible(true), 10);
    return () => clearTimeout(timer);
  }, []);

  return (
    <>
      {/* Overlay for mobile/tablet */}
      <div
        className={`fixed inset-0 bg-black/50 z-40 lg:hidden transition-opacity duration-150 ${
          isVisible ? 'opacity-100' : 'opacity-0'
        }`}
        onClick={onClose}
      />

      {/* Sidebar */}
      <div
        className={`fixed top-0 right-0 h-screen w-full sm:w-96 lg:w-96 shadow-xl z-50 overflow-y-auto flex-shrink-0 transform transition-transform duration-[180ms] ease-in-out lg:border-l border-border ${
          isVisible ? 'translate-x-0' : 'translate-x-full'
        }`}
        style={{ backgroundColor: tokens.surface.elevated }}
      >
        {/* Header */}
        <div
          className="sticky top-0 border-b border-border p-4 flex items-center justify-between z-10"
          style={{ backgroundColor: tokens.surface.elevated }}
        >
          <h2 className="text-lg font-semibold">Task Details</h2>
          <Button
            variant="ghost"
            size="icon"
            onClick={onClose}
            className="h-8 w-8"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          {isLoading ? (
            <>
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-24 w-full" />
              <Skeleton className="h-32 w-full" />
            </>
          ) : task ? (
            <>
              {/* Title and Priority */}
              <div>
                <div className="flex items-start gap-2 mb-2">
                  <h3 className="text-xl font-bold flex-1">{task.title}</h3>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Badge
                    variant={
                      task.priority_score >= 90
                        ? "destructive"
                        : task.priority_score >= 75
                        ? "default"
                        : "secondary"
                    }
                    className="text-sm"
                  >
                    Priority: {Math.round(task.priority_score)}
                  </Badge>
                  {task.bump_count > 0 && (
                    <Badge
                      variant="outline"
                      style={{
                        color: tokens.status.warning.default,
                        borderColor: tokens.status.warning.default,
                      }}
                    >
                      ⚠️ Bumped {task.bump_count}x
                    </Badge>
                  )}
                  {task.bump_count >= 3 && (
                    <Badge variant="destructive">
                      AT RISK
                    </Badge>
                  )}
                  {task.series_id && (
                    <Badge variant="outline" className="text-purple-600 dark:text-purple-400 border-purple-300 dark:border-purple-600 bg-purple-50 dark:bg-purple-950">
                      <Repeat className="h-3 w-3 mr-1" />
                      Recurring
                    </Badge>
                  )}
                </div>
              </div>

              {/* Action Buttons */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Actions</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setEditDialogOpen(true)}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      <Pencil className="mr-2 h-4 w-4" />
                      Edit
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => bumpTask.mutate({ id: taskId })}
                      disabled={bumpTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      Bump
                    </Button>
                    <Button
                      size="sm"
                      onClick={() => {
                        completeTask.mutate(taskId);
                        onClose();
                      }}
                      disabled={completeTask.isPending || hasAnyBlocker}
                      title={
                        hasDependencyBlocker
                          ? 'Complete all blocking tasks first'
                          : hasSubtaskBlocker
                          ? 'Complete all subtasks first'
                          : undefined
                      }
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      {hasDependencyBlocker && <Lock className="mr-1 h-3 w-3" />}
                      {hasSubtaskBlocker && !hasDependencyBlocker && <ListChecks className="mr-1 h-3 w-3" />}
                      Complete
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => {
                        if (window.confirm('Are you sure you want to delete this task?')) {
                          deleteTask.mutate(taskId);
                          onClose();
                        }
                      }}
                      disabled={deleteTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-lg cursor-pointer"
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </Button>
                  </div>
                </CardContent>
              </Card>

              {/* Subtasks - only for regular tasks */}
              {task.task_type === 'regular' && (
                <SubtaskList
                  parentTaskId={taskId}
                  parentTask={task}
                  onParentCompleted={onClose}
                />
              )}

              {/* Dependencies - only for regular tasks */}
              {task.task_type === 'regular' && (
                <DependencySection
                  taskId={taskId}
                  task={task}
                />
              )}

              {/* Description */}
              {task.description && (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-medium">Description</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground">{task.description}</p>
                  </CardContent>
                </Card>
              )}

              {/* Context */}
              {task.context && (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-medium">Context</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm italic text-muted-foreground">&quot;{task.context}&quot;</p>
                  </CardContent>
                </Card>
              )}

              {/* Task Properties */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Properties</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {task.category && (
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Category</span>
                      <Badge variant="outline">{task.category}</Badge>
                    </div>
                  )}
                  {task.estimated_effort && (
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Effort</span>
                      <Badge variant="outline" className="capitalize">{task.estimated_effort}</Badge>
                    </div>
                  )}
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">Status</span>
                    <Badge variant="outline" className="capitalize">{task.status}</Badge>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">User Priority</span>
                    <span className="text-sm font-medium">{task.user_priority}/10</span>
                  </div>
                </CardContent>
              </Card>

              {/* Dates */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Timeline</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {task.due_date && (
                    <div className="flex justify-between items-center">
                      <span className="text-sm text-muted-foreground">Due Date</span>
                      <span className="text-sm font-medium">
                        {new Date(task.due_date).toLocaleDateString()}
                      </span>
                    </div>
                  )}
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">Created</span>
                    <span className="text-sm font-medium">
                      {new Date(task.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">Last Updated</span>
                    <span className="text-sm font-medium">
                      {new Date(task.updated_at).toLocaleDateString()}
                    </span>
                  </div>
                </CardContent>
              </Card>

              {/* Related People */}
              {task.related_people && task.related_people.length > 0 && (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-medium">Related People</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex flex-wrap gap-2">
                      {task.related_people.map((person) => (
                        <Badge key={person} variant="secondary">
                          👤 {person}
                        </Badge>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              )}

              {/* Priority Breakdown Panel */}
              {task.priority_breakdown ? (
                <PriorityBreakdownPanel
                  breakdown={task.priority_breakdown}
                  finalScore={Math.round(task.priority_score)}
                />
              ) : (
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-medium">Priority Calculation</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-xs text-muted-foreground">
                    <p>
                      The priority score is calculated using our intelligent algorithm that considers:
                    </p>
                    <ul className="list-disc list-inside space-y-1 ml-2">
                      <li>Your set priority ({task.user_priority}/10)</li>
                      <li>Time since creation (time decay)</li>
                      <li>Number of delays ({task.bump_count} bumps)</li>
                      <li>Due date proximity</li>
                      <li>Estimated effort</li>
                    </ul>
                    <p className="pt-2">
                      Final Score: <span className="font-bold text-foreground">{Math.round(task.priority_score)}</span>
                    </p>
                  </CardContent>
                </Card>
              )}
            </>
          ) : (
            <div className="text-center py-8">
              <p className="text-muted-foreground">Task not found</p>
            </div>
          )}
        </div>
      </div>

      {/* Edit Task Dialog */}
      {task && (
        <EditTaskDialog
          open={editDialogOpen}
          onOpenChange={setEditDialogOpen}
          task={task}
        />
      )}
    </>
  );
}
