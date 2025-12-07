'use client';

import { useKeyboardShortcuts } from '@/contexts/KeyboardShortcutsContext';
import { useDialogKeyboardShortcuts } from '@/hooks/useDialogKeyboardShortcuts';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Keyboard } from 'lucide-react';

interface ShortcutItem {
  keys: string[];
  description: string;
  category: 'global' | 'navigation' | 'actions' | 'pomodoro';
}

const shortcuts: ShortcutItem[] = [
  // Global shortcuts
  { keys: ['Cmd', 'K'], description: 'Quick Add Task', category: 'global' },
  { keys: ['?'], description: 'Show Keyboard Shortcuts', category: 'global' },
  { keys: ['Esc'], description: 'Close Dialog', category: 'global' },

  // Navigation shortcuts
  { keys: ['j'], description: 'Select Next Task', category: 'navigation' },
  { keys: ['k'], description: 'Select Previous Task', category: 'navigation' },
  { keys: ['Enter'], description: 'Open Task Details', category: 'navigation' },

  // Action shortcuts
  { keys: ['e'], description: 'Edit Selected Task', category: 'actions' },
  { keys: ['c'], description: 'Complete Selected Task', category: 'actions' },
  { keys: ['d'], description: 'Delete Selected Task', category: 'actions' },

  // Pomodoro shortcuts
  { keys: ['p'], description: 'Start/Pause Pomodoro Timer', category: 'pomodoro' },
];

const categoryLabels = {
  global: 'Global',
  navigation: 'Task Navigation',
  actions: 'Task Actions',
  pomodoro: 'Pomodoro Timer',
};

export function KeyboardShortcutsHelp() {
  const { state, actions } = useKeyboardShortcuts();

  // Track dialog state for keyboard shortcuts
  useDialogKeyboardShortcuts(state.helpDialogOpen);

  const isMac = typeof window !== 'undefined' && navigator.platform.toUpperCase().indexOf('MAC') >= 0;

  const renderKeys = (keys: string[]) => {
    return keys.map((key, index) => {
      const displayKey = key === 'Cmd' ? (isMac ? '⌘' : 'Ctrl') : key;
      return (
        <span key={index} className="inline-flex items-center gap-1">
          {index > 0 && <span className="text-muted-foreground">+</span>}
          <Badge variant="outline" className="font-mono text-xs px-2 py-0.5">
            {displayKey}
          </Badge>
        </span>
      );
    });
  };

  const groupedShortcuts = shortcuts.reduce((acc, shortcut) => {
    if (!acc[shortcut.category]) {
      acc[shortcut.category] = [];
    }
    acc[shortcut.category].push(shortcut);
    return acc;
  }, {} as Record<string, ShortcutItem[]>);

  return (
    <Dialog open={state.helpDialogOpen} onOpenChange={actions.setHelpDialogOpen}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <Keyboard className="h-5 w-5" />
            <DialogTitle>Keyboard Shortcuts</DialogTitle>
          </div>
          <DialogDescription>
            Navigate faster with these keyboard shortcuts
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {Object.entries(groupedShortcuts).map(([category, items]) => (
            <div key={category}>
              <h3 className="font-semibold text-sm text-muted-foreground mb-3">
                {categoryLabels[category as keyof typeof categoryLabels]}
              </h3>
              <div className="space-y-2">
                {items.map((shortcut, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between py-2 px-3 rounded-md hover:bg-muted/50 transition-colors"
                  >
                    <span className="text-sm">{shortcut.description}</span>
                    <div className="flex items-center gap-1">
                      {renderKeys(shortcut.keys)}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="border-t pt-4">
          <p className="text-xs text-muted-foreground">
            <strong>Note:</strong> Task navigation shortcuts (j, k, e, c, x, d) only work when no
            input field is focused and no dialog is open.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}
