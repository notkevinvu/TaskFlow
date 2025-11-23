'use client';

import { useState } from 'react';
import { useTasks, useBumpTask, useCompleteTask, useDeleteTask, useAtRiskTasks } from '@/hooks/useTasks';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { TaskDetailsSidebar } from "@/components/TaskDetailsSidebar";
import { CreateTaskDialog } from "@/components/CreateTaskDialog";
import { EditTaskDialog } from "@/components/EditTaskDialog";
import { Calendar } from "@/components/Calendar";
import { Plus, Trash2, Pencil } from "lucide-react";
import { Task, CalendarDayData } from "@/lib/api";

export default function DashboardPage() {
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const { data: tasksData, isLoading } = useTasks();
  const { data: atRiskData } = useAtRiskTasks();
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();
  const deleteTask = useDeleteTask();

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold">Today's Priorities</h2>
          <p className="text-muted-foreground">
            Loading your intelligently prioritized tasks...
          </p>
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-24" />
          ))}
        </div>
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  const tasks = tasksData?.tasks || [];
  const atRiskCount = atRiskData?.count || 0;
  const quickWins = tasks.filter(
    (t) => t.estimated_effort === 'small' && t.bump_count === 0
  ).length;
  const totalTasks = tasks.length;

  return (
    <div className={`transition-all duration-[180ms] ${selectedTaskId ? 'lg:pr-96' : ''}`}>
      {/* Main Content */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-3xl font-bold">Today's Priorities</h2>
            <p className="text-muted-foreground">
              Tasks sorted by intelligent priority algorithm
            </p>
          </div>
          <Button
            onClick={() => setCreateDialogOpen(true)}
            className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
          >
            <Plus className="mr-2 h-4 w-4" />
            Quick Add
          </Button>
        </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Tasks
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalTasks}</div>
            <p className="text-xs text-muted-foreground">
              Active tasks to complete
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              At Risk
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">{atRiskCount}</div>
            <p className="text-xs text-muted-foreground">
              Tasks bumped 3+ times
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Quick Wins
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">{quickWins}</div>
            <p className="text-xs text-muted-foreground">
              Small tasks ready to complete
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Calendar and Task List - Responsive Layout */}
      <div className="grid grid-cols-1 xl:grid-cols-[450px,1fr] gap-6">
        {/* Calendar */}
        <div className="xl:sticky xl:top-6 xl:self-start">
          <Calendar
            onDayClick={(date, dayData) => {
              // For now, just show console log
              // Later we'll add popover/dialog here
              console.log('Day clicked:', date, dayData);
            }}
          />
        </div>

        {/* Task List */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Your Tasks</h3>
          {tasks.length === 0 ? (
            <Card>
              <CardContent className="pt-6 text-center">
                <p className="text-muted-foreground">
                  No tasks yet. Create your first task to get started!
                </p>
              </CardContent>
            </Card>
          ) : (
            tasks.map((task) => (
            <Card
              key={task.id}
              className="hover:shadow-md transition-shadow cursor-pointer"
              onClick={() => setSelectedTaskId(task.id)}
            >
              <CardContent className="pt-6">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h4 className="font-semibold">{task.title}</h4>
                      <Badge
                        variant={
                          task.priority_score >= 90
                            ? "destructive"
                            : task.priority_score >= 75
                            ? "default"
                            : "secondary"
                        }
                      >
                        {Math.round(task.priority_score)}
                      </Badge>
                      {task.bump_count > 0 && (
                        <Badge variant="outline" className="text-yellow-600 border-yellow-600">
                          ⚠️ Bumped {task.bump_count}x
                        </Badge>
                      )}
                      {task.estimated_effort && (
                        <Badge variant="outline" className="capitalize">
                          {task.estimated_effort}
                        </Badge>
                      )}
                    </div>
                    {task.description && (
                      <p className="text-sm text-muted-foreground mb-2">
                        {task.description}
                      </p>
                    )}
                    {task.context && (
                      <p className="text-sm italic text-muted-foreground mb-2">
                        "{task.context}"
                      </p>
                    )}
                    <div className="flex gap-4 text-sm text-muted-foreground">
                      {task.category && <span>📁 {task.category}</span>}
                      {task.due_date && (
                        <span>📅 Due: {new Date(task.due_date).toLocaleDateString()}</span>
                      )}
                      {task.related_people && task.related_people.length > 0 && (
                        <span>👥 {task.related_people.join(', ')}</span>
                      )}
                    </div>
                  </div>
                  <div className="flex gap-2 ml-4">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditingTask(task);
                      }}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        bumpTask.mutate({ id: task.id });
                      }}
                      disabled={bumpTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      Bump
                    </Button>
                    <Button
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        completeTask.mutate(task.id);
                      }}
                      disabled={completeTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
                    >
                      Complete
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        if (window.confirm('Are you sure you want to delete this task?')) {
                          deleteTask.mutate(task.id);
                          if (selectedTaskId === task.id) {
                            setSelectedTaskId(null);
                          }
                        }
                      }}
                      disabled={deleteTask.isPending}
                      className="transition-all hover:scale-105 hover:shadow-lg cursor-pointer"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
          )}
        </div>
      </div>
    </div>

    {/* Task Details Sidebar */}
    {selectedTaskId && (
      <TaskDetailsSidebar
        taskId={selectedTaskId}
        onClose={() => setSelectedTaskId(null)}
      />
    )}

    {/* Create Task Dialog */}
    <CreateTaskDialog
      open={createDialogOpen}
      onOpenChange={setCreateDialogOpen}
    />

    {/* Edit Task Dialog */}
    {editingTask && (
      <EditTaskDialog
        open={!!editingTask}
        onOpenChange={(open) => !open && setEditingTask(null)}
        task={editingTask}
      />
    )}
  </div>
);
}
