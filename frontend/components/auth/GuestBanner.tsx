'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { X, Sparkles } from 'lucide-react';
import { ConvertDialog } from './ConvertDialog';
import { tokens } from '@/lib/tokens';

// Helper to calculate days remaining from expiry date
function calculateDaysRemaining(expiresAt: string | undefined): number {
  if (!expiresAt) return 30;
  return Math.max(0, Math.ceil((new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24)));
}

export function GuestBanner() {
  const { user, isAnonymous } = useAuth();
  const [dismissed, setDismissed] = useState(false);
  const [convertDialogOpen, setConvertDialogOpen] = useState(false);
  const [daysRemaining, setDaysRemaining] = useState(() => calculateDaysRemaining(user?.expires_at));

  // Update countdown periodically (every hour) to prevent stale values
  useEffect(() => {
    const updateDays = () => {
      setDaysRemaining(calculateDaysRemaining(user?.expires_at));
    };

    // Update immediately when user changes
    updateDays();

    // Refresh hourly to keep countdown accurate
    const interval = setInterval(updateDays, 60 * 60 * 1000);
    return () => clearInterval(interval);
  }, [user?.expires_at]);

  // Don't show if not anonymous or dismissed
  if (!user || !isAnonymous() || dismissed) {
    return null;
  }

  return (
    <>
      <div
        className="flex items-center justify-between px-4 py-2 text-sm"
        style={{
          background: tokens.gradient.primary,
          color: 'white',
        }}
      >
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4" />
          <span>
            You&apos;re using TaskFlow as a guest.{' '}
            <strong>Your data expires in {daysRemaining} days.</strong>
          </span>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setConvertDialogOpen(true)}
            className="h-7 text-xs"
          >
            Create Account
          </Button>
          <button
            onClick={() => setDismissed(true)}
            className="p-1 hover:bg-white/20 rounded transition-colors"
            aria-label="Dismiss banner"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      </div>

      <ConvertDialog
        open={convertDialogOpen}
        onOpenChange={setConvertDialogOpen}
      />
    </>
  );
}
