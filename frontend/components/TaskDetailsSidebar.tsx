'use client';

import { useEffect, useState } from 'react';
import { useTask } from '@/hooks/useTasks';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { X } from 'lucide-react';

interface TaskDetailsSidebarProps {
  taskId: string;
  onClose: () => void;
}

export function TaskDetailsSidebar({ taskId, onClose }: TaskDetailsSidebarProps) {
  const { data: task, isLoading } = useTask(taskId);
  const [isVisible, setIsVisible] = useState(false);

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
      <div className={`fixed top-0 right-0 h-screen w-full sm:w-96 lg:w-96 bg-white shadow-xl z-50 overflow-y-auto flex-shrink-0 transform transition-transform duration-[180ms] ease-in-out lg:border-l ${
        isVisible ? 'translate-x-0' : 'translate-x-full'
      }`}>
        {/* Header */}
        <div className="sticky top-0 bg-white border-b p-4 flex items-center justify-between z-10">
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
                    <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                      ⚠️ Bumped {task.bump_count}x
                    </Badge>
                  )}
                  {task.bump_count >= 3 && (
                    <Badge variant="destructive">
                      AT RISK
                    </Badge>
                  )}
                </div>
              </div>

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
                    <p className="text-sm italic text-muted-foreground">"{task.context}"</p>
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
                    <span className="text-sm font-medium">{task.user_priority}/100</span>
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

              {/* Priority Calculation Info */}
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium">Priority Calculation</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-xs text-muted-foreground">
                  <p>
                    The priority score is calculated using our intelligent algorithm that considers:
                  </p>
                  <ul className="list-disc list-inside space-y-1 ml-2">
                    <li>Your set priority ({task.user_priority}/100)</li>
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
            </>
          ) : (
            <div className="text-center py-8">
              <p className="text-muted-foreground">Task not found</p>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
