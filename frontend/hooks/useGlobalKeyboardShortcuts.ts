'use client';

import { useEffect } from 'react';
import { useKeyboardShortcuts } from '@/contexts/KeyboardShortcutsContext';
import { usePomodoro } from '@/contexts/PomodoroContext';

/**
 * Global keyboard shortcuts hook
 * Handles app-wide keyboard shortcuts that work everywhere
 * - Cmd/Ctrl+K: Quick Add task
 * - ?: Show keyboard shortcuts help
 * - ESC: Close dialogs
 * - P: Toggle Pomodoro timer (when not in input)
 */
export function useGlobalKeyboardShortcuts() {
  const { state, actions } = useKeyboardShortcuts();
  const { state: pomodoroState, actions: pomodoroActions } = usePomodoro();

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Don't handle shortcuts if disabled
      if (!state.enabled) return;

      const isMac = typeof window !== 'undefined' && navigator.platform.toUpperCase().indexOf('MAC') >= 0;
      const modKey = isMac ? e.metaKey : e.ctrlKey;

      // Cmd/Ctrl+K: Quick Add (works globally, even with input focused)
      if (modKey && e.key === 'k') {
        e.preventDefault();
        actions.setQuickAddOpen(!state.quickAddOpen);
        return;
      }

      // ?: Show keyboard shortcuts help (works globally, even with input focused)
      // Note: '?' requires Shift on most keyboards, so we just check for the key
      if (e.key === '?' && !modKey) {
        e.preventDefault();
        actions.setHelpDialogOpen(true);
        return;
      }

      // ESC: Close dialogs
      if (e.key === 'Escape') {
        if (state.quickAddOpen) {
          actions.setQuickAddOpen(false);
          return;
        }
        if (state.helpDialogOpen) {
          actions.setHelpDialogOpen(false);
          return;
        }
      }

      // P: Toggle Pomodoro timer (only when not in input)
      if (e.key === 'p' && !modKey && !state.inputFocused && state.dialogCount === 0) {
        e.preventDefault();
        if (pomodoroState.isRunning && !pomodoroState.isPaused) {
          pomodoroActions.pause();
        } else if (pomodoroState.isPaused) {
          pomodoroActions.resume();
        } else {
          pomodoroActions.start();
        }
        return;
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [state, actions, pomodoroState, pomodoroActions]);

  // Track input focus state for conditional shortcut handling
  useEffect(() => {
    function handleFocusIn(e: FocusEvent) {
      const target = e.target as HTMLElement;
      const isInput =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.contentEditable === 'true';

      if (isInput) {
        actions.setInputFocused(true);
      }
    }

    function handleFocusOut(e: FocusEvent) {
      const target = e.target as HTMLElement;
      const isInput =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.contentEditable === 'true';

      if (isInput) {
        actions.setInputFocused(false);
      }
    }

    window.addEventListener('focusin', handleFocusIn);
    window.addEventListener('focusout', handleFocusOut);
    return () => {
      window.removeEventListener('focusin', handleFocusIn);
      window.removeEventListener('focusout', handleFocusOut);
    };
  }, [actions]);
}
