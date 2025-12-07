'use client';

import { QueryClient, QueryClientProvider, QueryCache, MutationCache } from '@tanstack/react-query';
import { Toaster } from 'sonner';
import { useState } from 'react';
import { ThemeProvider } from '@/components/ThemeProvider';
import { KeyboardShortcutsProvider } from '@/contexts/KeyboardShortcutsContext';
import { getApiErrorMessage } from '@/lib/api';

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient({
    queryCache: new QueryCache({
      onError: (error, query) => {
        // Log all query errors for debugging
        const context = query.queryKey[0] as string || 'Query';
        console.error(`[${context}]`, getApiErrorMessage(error, 'Query failed'));
      },
    }),
    mutationCache: new MutationCache({
      onError: (error, _variables, _context, mutation) => {
        // Log mutation errors that aren't already handled by onError callbacks
        if (!mutation.options.onError) {
          console.error('[Mutation]', getApiErrorMessage(error, 'Mutation failed'));
        }
      },
    }),
    defaultOptions: {
      queries: {
        staleTime: 60 * 1000, // 1 minute
        refetchOnWindowFocus: false,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider
        attribute="class"
        defaultTheme="system"
        enableSystem
        disableTransitionOnChange
      >
        <KeyboardShortcutsProvider>
          {children}
          <Toaster position="bottom-center" />
        </KeyboardShortcutsProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}
