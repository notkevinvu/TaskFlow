'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { X, Sparkles } from 'lucide-react';
import { ConvertDialog } from './ConvertDialog';
import { tokens } from '@/lib/tokens';

export function GuestBanner() {
  const { user, isAnonymous } = useAuth();
  const [dismissed, setDismissed] = useState(false);
  const [convertDialogOpen, setConvertDialogOpen] = useState(false);

  // Don't show if not anonymous or dismissed
  if (!user || !isAnonymous() || dismissed) {
    return null;
  }

  // Calculate days remaining until expiry
  const daysRemaining = user.expires_at
    ? Math.max(0, Math.ceil((new Date(user.expires_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24)))
    : 30;

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
