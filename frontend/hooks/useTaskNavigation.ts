'use client';

import { useState, useEffect, useCallback } from 'react';
import { useKeyboardShortcuts } from '@/contexts/KeyboardShortcutsContext';
import { Task } from '@/lib/api';

export interface UseTaskNavigationOptions {
  tasks: Task[];
  onTaskSelect: (taskId: string) => void;
  onTaskEdit: (task: Task) => void;
  onTaskComplete: (taskId: string) => void;
  onTaskDelete: (taskId: string) => void;
  isDialogOpen?: boolean;
}

export function useTaskNavigation({
  tasks,
  onTaskSelect,
  onTaskEdit,
  onTaskComplete,
  onTaskDelete,
  isDialogOpen = false,
}: UseTaskNavigationOptions) {
  const { state } = useKeyboardShortcuts();
  const [selectedIndex, setSelectedIndex] = useState<number>(-1);

  // Reset selection when tasks change significantly
  useEffect(() => {
    if (selectedIndex >= tasks.length) {
      setSelectedIndex(tasks.length > 0 ? 0 : -1);
    }
  }, [tasks.length, selectedIndex]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Check if an input is currently focused (more reliable than tracked state)
      const activeElement = document.activeElement as HTMLElement | null;
      const isInputCurrentlyFocused = activeElement && (
        activeElement.tagName === 'INPUT' ||
        activeElement.tagName === 'TEXTAREA' ||
        activeElement.contentEditable === 'true'
      );

      // Don't handle shortcuts if:
      // - Shortcuts are disabled globally
      // - Any dialog is open
      // - Input field is focused
      if (!state.enabled || isDialogOpen || state.dialogCount > 0 || isInputCurrentlyFocused) {
        return;
      }

      const key = e.key.toLowerCase();

      // j: Move down
      if (key === 'j') {
        e.preventDefault();
        setSelectedIndex((prev) => {
          if (tasks.length === 0) return -1;
          if (prev === -1) return 0;
          // Wrap around to top
          return prev >= tasks.length - 1 ? 0 : prev + 1;
        });
        return;
      }

      // k: Move up
      if (key === 'k') {
        e.preventDefault();
        setSelectedIndex((prev) => {
          if (tasks.length === 0) return -1;
          if (prev === -1) return tasks.length - 1;
          // Wrap around to bottom
          return prev <= 0 ? tasks.length - 1 : prev - 1;
        });
        return;
      }

      // Shortcuts that require a task to be selected
      if (selectedIndex !== -1 && selectedIndex < tasks.length) {
        const selectedTask = tasks[selectedIndex];

        // Enter: Open task details sidebar
        if (key === 'enter') {
          e.preventDefault();
          onTaskSelect(selectedTask.id);
          return;
        }

        // e: Edit task
        if (key === 'e') {
          e.preventDefault();
          onTaskEdit(selectedTask);
          return;
        }

        // c: Complete task
        if (key === 'c') {
          e.preventDefault();
          onTaskComplete(selectedTask.id);
          return;
        }

        // d: Delete task (will trigger confirmation dialog)
        if (key === 'd') {
          e.preventDefault();
          onTaskDelete(selectedTask.id);
          return;
        }
      }
    },
    [
      state.enabled,
      state.dialogCount,
      isDialogOpen,
      tasks,
      selectedIndex,
      onTaskSelect,
      onTaskEdit,
      onTaskComplete,
      onTaskDelete,
    ]
  );

  // Attach keyboard event listener
  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  return {
    selectedIndex,
    selectedTaskId: selectedIndex >= 0 && selectedIndex < tasks.length ? tasks[selectedIndex].id : null,
    setSelectedIndex,
  };
}
