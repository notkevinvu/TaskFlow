'use client';

import React, { createContext, useContext, useState, useCallback, useMemo, ReactNode } from 'react';

export interface KeyboardShortcutsState {
  // Whether keyboard shortcuts are currently enabled
  enabled: boolean;
  // Whether the Quick Add dialog is open
  quickAddOpen: boolean;
  // Whether the shortcuts help dialog is open
  helpDialogOpen: boolean;
  // Number of dialogs currently open (shortcuts disabled when > 0)
  dialogCount: number;
  // Whether an input field is currently focused (some shortcuts disabled)
  inputFocused: boolean;
}

export interface KeyboardShortcutsActions {
  // Enable/disable all shortcuts
  setEnabled: (enabled: boolean) => void;
  // Open/close Quick Add dialog
  setQuickAddOpen: (open: boolean) => void;
  // Open/close help dialog
  setHelpDialogOpen: (open: boolean) => void;
  // Track dialog open/close for disabling shortcuts
  incrementDialogCount: () => void;
  decrementDialogCount: () => void;
  // Track input focus state
  setInputFocused: (focused: boolean) => void;
}

const KeyboardShortcutsStateContext = createContext<KeyboardShortcutsState | null>(null);
const KeyboardShortcutsActionsContext = createContext<KeyboardShortcutsActions | null>(null);

export function KeyboardShortcutsProvider({ children }: { children: ReactNode }) {
  const [enabled, setEnabled] = useState(true);
  const [quickAddOpen, setQuickAddOpen] = useState(false);
  const [helpDialogOpen, setHelpDialogOpen] = useState(false);
  const [dialogCount, setDialogCount] = useState(0);
  const [inputFocused, setInputFocused] = useState(false);

  const incrementDialogCount = useCallback(() => {
    setDialogCount((count) => count + 1);
  }, []);

  const decrementDialogCount = useCallback(() => {
    setDialogCount((count) => Math.max(0, count - 1));
  }, []);

  const state: KeyboardShortcutsState = {
    enabled,
    quickAddOpen,
    helpDialogOpen,
    dialogCount,
    inputFocused,
  };

  const actions: KeyboardShortcutsActions = useMemo(
    () => ({
      setEnabled,
      setQuickAddOpen,
      setHelpDialogOpen,
      incrementDialogCount,
      decrementDialogCount,
      setInputFocused,
    }),
    [incrementDialogCount, decrementDialogCount]
  );

  return (
    <KeyboardShortcutsStateContext.Provider value={state}>
      <KeyboardShortcutsActionsContext.Provider value={actions}>
        {children}
      </KeyboardShortcutsActionsContext.Provider>
    </KeyboardShortcutsStateContext.Provider>
  );
}

export function useKeyboardShortcutsState(): KeyboardShortcutsState {
  const context = useContext(KeyboardShortcutsStateContext);
  if (!context) {
    throw new Error('useKeyboardShortcutsState must be used within KeyboardShortcutsProvider');
  }
  return context;
}

export function useKeyboardShortcutsActions(): KeyboardShortcutsActions {
  const context = useContext(KeyboardShortcutsActionsContext);
  if (!context) {
    throw new Error('useKeyboardShortcutsActions must be used within KeyboardShortcutsProvider');
  }
  return context;
}

export function useKeyboardShortcuts() {
  return {
    state: useKeyboardShortcutsState(),
    actions: useKeyboardShortcutsActions(),
  };
}
