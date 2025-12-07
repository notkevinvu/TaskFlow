'use client';

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
  useMemo,
  ReactNode,
} from 'react';

// Timer mode type - must be defined before POMODORO_DURATIONS for type safety
export type PomodoroMode = 'work' | 'shortBreak' | 'longBreak';

// Pomodoro timer durations in seconds
// Record<PomodoroMode, number> ensures all modes have durations at compile time
export const POMODORO_DURATIONS: Record<PomodoroMode, number> = {
  work: 25 * 60,      // 25 minutes
  shortBreak: 5 * 60,  // 5 minutes
  longBreak: 15 * 60,  // 15 minutes
};

// Linked task - both id and title are always set together or both null
export interface LinkedTask {
  id: string;
  title: string;
}

export interface PomodoroState {
  // Current timer state
  isRunning: boolean;
  isPaused: boolean;
  mode: PomodoroMode;
  timeRemaining: number; // seconds

  // Session tracking
  completedSessions: number;
  sessionsUntilLongBreak: number; // Usually 4

  // Task linking - coupled fields prevent mismatched state
  linkedTask: LinkedTask | null;
}

export interface PomodoroActions {
  // Timer controls
  start: (linkedTask?: LinkedTask) => void;
  pause: () => void;
  resume: () => void;
  reset: () => void;
  skip: () => void;

  // Task linking
  linkTask: (task: LinkedTask) => void;
  unlinkTask: () => void;

  // Settings
  setMode: (mode: PomodoroMode) => void;
}

const STORAGE_KEY = 'taskflow_pomodoro_state';
const SESSIONS_UNTIL_LONG_BREAK = 4;

// Helper to create audio notification using Web Audio API
function playNotificationSound(type: 'work' | 'break') {
  if (typeof window === 'undefined') return;

  try {
    const audioContext = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);

    // Different tones for work vs break completion
    if (type === 'work') {
      // Cheerful ascending tone for work completion
      oscillator.frequency.setValueAtTime(523.25, audioContext.currentTime); // C5
      oscillator.frequency.setValueAtTime(659.25, audioContext.currentTime + 0.15); // E5
      oscillator.frequency.setValueAtTime(783.99, audioContext.currentTime + 0.3); // G5
    } else {
      // Gentle descending tone for break ending
      oscillator.frequency.setValueAtTime(783.99, audioContext.currentTime); // G5
      oscillator.frequency.setValueAtTime(659.25, audioContext.currentTime + 0.15); // E5
      oscillator.frequency.setValueAtTime(523.25, audioContext.currentTime + 0.3); // C5
    }

    gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5);

    oscillator.start(audioContext.currentTime);
    oscillator.stop(audioContext.currentTime + 0.5);

    // Close AudioContext after sound finishes to prevent memory leak
    oscillator.onended = () => {
      audioContext.close().catch((closeError) => {
        console.warn('[Pomodoro] AudioContext cleanup failed:', {
          error: closeError instanceof Error ? closeError.message : String(closeError),
        });
      });
    };
  } catch (error) {
    console.warn('[Pomodoro] Audio notification failed:', {
      type,
      error: error instanceof Error ? error.message : String(error),
    });
  }
}

function getInitialState(): PomodoroState {
  if (typeof window === 'undefined') {
    return {
      isRunning: false,
      isPaused: false,
      mode: 'work',
      timeRemaining: POMODORO_DURATIONS.work,
      completedSessions: 0,
      sessionsUntilLongBreak: SESSIONS_UNTIL_LONG_BREAK,
      linkedTask: null,
    };
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored) as Partial<PomodoroState>;
      return {
        isRunning: false, // Always start paused on reload
        isPaused: parsed.isPaused ?? false,
        mode: parsed.mode ?? 'work',
        timeRemaining: parsed.timeRemaining ?? POMODORO_DURATIONS[parsed.mode ?? 'work'],
        completedSessions: parsed.completedSessions ?? 0,
        sessionsUntilLongBreak: parsed.sessionsUntilLongBreak ?? SESSIONS_UNTIL_LONG_BREAK,
        linkedTask: parsed.linkedTask ?? null,
      };
    }
  } catch (error) {
    console.warn('[Pomodoro] Failed to restore saved state, using defaults:', {
      error: error instanceof Error ? error.message : String(error),
    });
    // Clear corrupted state to prevent repeated failures
    try {
      localStorage.removeItem(STORAGE_KEY);
    } catch (cleanupError) {
      console.warn('[Pomodoro] localStorage completely blocked, state will not persist:', {
        error: cleanupError instanceof Error ? cleanupError.message : String(cleanupError),
      });
    }
  }

  return {
    isRunning: false,
    isPaused: false,
    mode: 'work',
    timeRemaining: POMODORO_DURATIONS.work,
    completedSessions: 0,
    sessionsUntilLongBreak: SESSIONS_UNTIL_LONG_BREAK,
    linkedTask: null,
  };
}

const PomodoroStateContext = createContext<PomodoroState | null>(null);
const PomodoroActionsContext = createContext<PomodoroActions | null>(null);

export function PomodoroProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<PomodoroState>(getInitialState);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  // Persist state to localStorage (excluding isRunning)
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const toStore: Partial<PomodoroState> = {
        mode: state.mode,
        timeRemaining: state.timeRemaining,
        completedSessions: state.completedSessions,
        sessionsUntilLongBreak: state.sessionsUntilLongBreak,
        linkedTask: state.linkedTask,
        isPaused: state.isPaused,
      };
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(toStore));
      } catch (error) {
        console.warn('[Pomodoro] Failed to persist state:', {
          error: error instanceof Error ? error.message : String(error),
        });
      }
    }
  }, [state]);

  // Timer tick effect
  useEffect(() => {
    if (state.isRunning && !state.isPaused) {
      intervalRef.current = setInterval(() => {
        setState((prev) => {
          if (prev.timeRemaining <= 1) {
            // Timer completed
            const wasWork = prev.mode === 'work';
            playNotificationSound(wasWork ? 'work' : 'break');

            if (wasWork) {
              // Work session completed
              const newCompleted = prev.completedSessions + 1;
              const isLongBreakDue = newCompleted % SESSIONS_UNTIL_LONG_BREAK === 0;
              const nextMode = isLongBreakDue ? 'longBreak' : 'shortBreak';

              return {
                ...prev,
                isRunning: false,
                isPaused: false,
                mode: nextMode,
                timeRemaining: POMODORO_DURATIONS[nextMode],
                completedSessions: newCompleted,
              };
            } else {
              // Break completed - ready for next work session
              return {
                ...prev,
                isRunning: false,
                isPaused: false,
                mode: 'work',
                timeRemaining: POMODORO_DURATIONS.work,
              };
            }
          }

          return {
            ...prev,
            timeRemaining: prev.timeRemaining - 1,
          };
        });
      }, 1000);

      return () => {
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
        }
      };
    }
  }, [state.isRunning, state.isPaused]);

  const start = useCallback((linkedTask?: LinkedTask) => {
    setState((prev) => ({
      ...prev,
      isRunning: true,
      isPaused: false,
      linkedTask: linkedTask ?? prev.linkedTask,
    }));
  }, []);

  const pause = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isPaused: true,
    }));
  }, []);

  const resume = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isRunning: true,
      isPaused: false,
    }));
  }, []);

  const reset = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isRunning: false,
      isPaused: false,
      timeRemaining: POMODORO_DURATIONS[prev.mode],
    }));
  }, []);

  const skip = useCallback(() => {
    setState((prev) => {
      const wasWork = prev.mode === 'work';

      if (wasWork) {
        // Skip to break (don't count as completed)
        const isLongBreakDue = (prev.completedSessions + 1) % SESSIONS_UNTIL_LONG_BREAK === 0;
        const nextMode = isLongBreakDue ? 'longBreak' : 'shortBreak';
        return {
          ...prev,
          isRunning: false,
          isPaused: false,
          mode: nextMode,
          timeRemaining: POMODORO_DURATIONS[nextMode],
        };
      } else {
        // Skip break - go to next work session
        return {
          ...prev,
          isRunning: false,
          isPaused: false,
          mode: 'work',
          timeRemaining: POMODORO_DURATIONS.work,
        };
      }
    });
  }, []);

  const linkTask = useCallback((task: LinkedTask) => {
    setState((prev) => ({
      ...prev,
      linkedTask: task,
    }));
  }, []);

  const unlinkTask = useCallback(() => {
    setState((prev) => ({
      ...prev,
      linkedTask: null,
    }));
  }, []);

  const setMode = useCallback((mode: PomodoroMode) => {
    setState((prev) => ({
      ...prev,
      isRunning: false,
      isPaused: false,
      mode,
      timeRemaining: POMODORO_DURATIONS[mode],
    }));
  }, []);

  const actions: PomodoroActions = useMemo(
    () => ({
      start,
      pause,
      resume,
      reset,
      skip,
      linkTask,
      unlinkTask,
      setMode,
    }),
    [start, pause, resume, reset, skip, linkTask, unlinkTask, setMode]
  );

  return (
    <PomodoroStateContext.Provider value={state}>
      <PomodoroActionsContext.Provider value={actions}>
        {children}
      </PomodoroActionsContext.Provider>
    </PomodoroStateContext.Provider>
  );
}

export function usePomodoroState(): PomodoroState {
  const context = useContext(PomodoroStateContext);
  if (!context) {
    throw new Error('usePomodoroState must be used within PomodoroProvider');
  }
  return context;
}

export function usePomodoroActions(): PomodoroActions {
  const context = useContext(PomodoroActionsContext);
  if (!context) {
    throw new Error('usePomodoroActions must be used within PomodoroProvider');
  }
  return context;
}

export function usePomodoro() {
  return {
    state: usePomodoroState(),
    actions: usePomodoroActions(),
  };
}

// Utility function to format time
export function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

// Utility function to get mode display name
export function getModeDisplayName(mode: PomodoroMode): string {
  switch (mode) {
    case 'work':
      return 'Focus';
    case 'shortBreak':
      return 'Short Break';
    case 'longBreak':
      return 'Long Break';
  }
}
