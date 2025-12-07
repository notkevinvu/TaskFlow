'use client';

import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  TooltipProvider,
} from '@/components/ui/tooltip';
import {
  usePomodoro,
  formatTime,
  getModeDisplayName,
  POMODORO_DURATIONS,
  PomodoroMode,
} from '@/contexts/PomodoroContext';
import {
  Play,
  Pause,
  RotateCcw,
  SkipForward,
  Timer,
  Coffee,
  X,
  Link2,
} from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import { useTasks } from '@/hooks/useTasks';
import { useEffect } from 'react';

/**
 * PomodoroWidget - A compact sidebar widget for the Pomodoro timer
 *
 * Features:
 * - Visual countdown timer with progress ring
 * - Play/pause/reset controls
 * - Mode switching (work/short break/long break)
 * - Task linking from URL or manual selection
 * - Session counter with progress to long break
 */
export function PomodoroWidget() {
  const { state, actions } = usePomodoro();
  const searchParams = useSearchParams();
  const selectedTaskId = searchParams.get('taskId');
  const { data: tasksData } = useTasks();

  // Auto-link task when selected in dashboard
  useEffect(() => {
    if (selectedTaskId && !state.linkedTaskId && tasksData?.tasks) {
      const task = tasksData.tasks.find((t) => t.id === selectedTaskId);
      if (task) {
        actions.linkTask(task.id, task.title);
      }
    }
  }, [selectedTaskId, state.linkedTaskId, tasksData?.tasks, actions]);

  const totalDuration = POMODORO_DURATIONS[state.mode];
  const progress = ((totalDuration - state.timeRemaining) / totalDuration) * 100;

  const isWork = state.mode === 'work';
  const modeColor = isWork
    ? 'text-red-500'
    : state.mode === 'shortBreak'
      ? 'text-green-500'
      : 'text-blue-500';
  const modeBgColor = isWork
    ? 'bg-red-100 dark:bg-red-900/30'
    : state.mode === 'shortBreak'
      ? 'bg-green-100 dark:bg-green-900/30'
      : 'bg-blue-100 dark:bg-blue-900/30';
  const progressColor = isWork
    ? 'bg-red-500'
    : state.mode === 'shortBreak'
      ? 'bg-green-500'
      : 'bg-blue-500';

  const handleModeClick = (mode: PomodoroMode) => {
    if (!state.isRunning) {
      actions.setMode(mode);
    }
  };

  return (
    <TooltipProvider>
      <div className="space-y-3">
        <div className="flex items-center justify-between px-2">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-1">
            <Timer className="h-3 w-3" />
            Pomodoro
          </p>
          {state.completedSessions > 0 && (
            <Badge variant="secondary" className="text-xs">
              {state.completedSessions} sessions
            </Badge>
          )}
        </div>

        {/* Timer Display */}
        <Card className="py-4 px-4">
          <div className="space-y-3">
            {/* Mode tabs */}
            <div className="flex gap-1 justify-center">
              {(['work', 'shortBreak', 'longBreak'] as PomodoroMode[]).map(
                (mode) => (
                  <Tooltip key={mode}>
                    <TooltipTrigger asChild>
                      <button
                        onClick={() => handleModeClick(mode)}
                        disabled={state.isRunning}
                        className={`
                          px-2 py-1 text-xs rounded-md transition-colors
                          ${
                            state.mode === mode
                              ? `${modeBgColor} ${modeColor} font-medium`
                              : 'text-muted-foreground hover:bg-accent disabled:opacity-50'
                          }
                        `}
                      >
                        {mode === 'work' ? 'Focus' : mode === 'shortBreak' ? 'Short' : 'Long'}
                      </button>
                    </TooltipTrigger>
                    <TooltipContent side="top">
                      <p>
                        {getModeDisplayName(mode)} (
                        {Math.floor(POMODORO_DURATIONS[mode] / 60)} min)
                      </p>
                    </TooltipContent>
                  </Tooltip>
                )
              )}
            </div>

            {/* Time display */}
            <div className="text-center">
              <div className={`text-4xl font-mono font-bold ${modeColor}`}>
                {formatTime(state.timeRemaining)}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                {getModeDisplayName(state.mode)}
                {state.isPaused && ' (Paused)'}
              </p>
            </div>

            {/* Progress bar */}
            <div className="relative h-2 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
              <div
                className={`h-full ${progressColor} rounded-full transition-all duration-1000`}
                style={{ width: `${progress}%` }}
              />
            </div>

            {/* Controls */}
            <div className="flex justify-center gap-2">
              {!state.isRunning || state.isPaused ? (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      size="sm"
                      onClick={() =>
                        state.isPaused ? actions.resume() : actions.start()
                      }
                      className="gap-1"
                    >
                      <Play className="h-4 w-4" />
                      {state.isPaused ? 'Resume' : 'Start'}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Start timer (P)</p>
                  </TooltipContent>
                </Tooltip>
              ) : (
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => actions.pause()}
                      className="gap-1"
                    >
                      <Pause className="h-4 w-4" />
                      Pause
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Pause timer (P)</p>
                  </TooltipContent>
                </Tooltip>
              )}

              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => actions.reset()}
                  >
                    <RotateCcw className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Reset timer</p>
                </TooltipContent>
              </Tooltip>

              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => actions.skip()}
                  >
                    <SkipForward className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Skip to {isWork ? 'break' : 'next focus'}</p>
                </TooltipContent>
              </Tooltip>
            </div>
          </div>
        </Card>

        {/* Session progress to long break */}
        {state.mode === 'work' && (
          <div className="px-2">
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex items-center justify-between text-xs cursor-default">
                  <span className="text-muted-foreground flex items-center gap-1">
                    <Coffee className="h-3 w-3" />
                    Long break in
                  </span>
                  <span className="font-medium">
                    {4 - (state.completedSessions % 4)} sessions
                  </span>
                </div>
              </TooltipTrigger>
              <TooltipContent side="top">
                <p>Complete 4 focus sessions for a 15-min break</p>
              </TooltipContent>
            </Tooltip>
          </div>
        )}

        {/* Linked task */}
        {state.linkedTaskId && state.linkedTaskTitle && (
          <div className="px-2">
            <div className="flex items-center gap-2 text-xs">
              <Link2 className="h-3 w-3 text-muted-foreground flex-shrink-0" />
              <span className="text-muted-foreground truncate flex-1">
                {state.linkedTaskTitle}
              </span>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-5 w-5 p-0"
                    onClick={() => actions.unlinkTask()}
                  >
                    <X className="h-3 w-3" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Unlink task</p>
                </TooltipContent>
              </Tooltip>
            </div>
          </div>
        )}
      </div>
    </TooltipProvider>
  );
}
