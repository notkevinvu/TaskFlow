'use client';

import { useEffect, useRef } from 'react';
import { useKeyboardShortcutsActions } from '@/contexts/KeyboardShortcutsContext';

/**
 * Hook to automatically track dialog open/close state for keyboard shortcuts.
 * When a dialog is open, it increments the dialog count (disabling shortcuts).
 * When the dialog closes, it decrements the count.
 *
 * @param open - Whether the dialog is currently open
 */
export function useDialogKeyboardShortcuts(open: boolean): void {
  const actions = useKeyboardShortcutsActions();
  const wasOpenRef = useRef(false);

  useEffect(() => {
    // Only act on actual changes
    if (open && !wasOpenRef.current) {
      // Dialog just opened
      actions.incrementDialogCount();
      wasOpenRef.current = true;
    } else if (!open && wasOpenRef.current) {
      // Dialog just closed
      actions.decrementDialogCount();
      wasOpenRef.current = false;
    }
  }, [open, actions]);

  // Cleanup on unmount - if dialog was open when unmounted, decrement
  useEffect(() => {
    return () => {
      if (wasOpenRef.current) {
        actions.decrementDialogCount();
      }
    };
  }, [actions]);
}
