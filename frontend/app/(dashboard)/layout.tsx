'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/hooks/useAuth';
import { ThemeToggle } from '@/components/ThemeToggle';
import { Calendar } from '@/components/Calendar';
import { GamificationWidget } from '@/components/GamificationWidget';
import { PomodoroWidget } from '@/components/PomodoroWidget';
import { CreateTaskDialog } from '@/components/CreateTaskDialog';
import { TemplatePickerDialog } from '@/components/TemplatePickerDialog';
import { ManageTemplatesDialog } from '@/components/ManageTemplatesDialog';
import { CreateTemplateDialog } from '@/components/CreateTemplateDialog';
import { EditTemplateDialog } from '@/components/EditTemplateDialog';
import { useEffect, useState } from 'react';
import { CreateTaskDTO, TaskTemplate } from '@/lib/api';
import { FileText, Settings } from 'lucide-react';

const navigation = [
  { name: 'Dashboard', href: '/dashboard' },
  { name: 'Analytics', href: '/analytics' },
  { name: 'Archive', href: '/archive' },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout, checkAuth, setMockUser } = useAuth();
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [initialDueDate, setInitialDueDate] = useState<string | undefined>(undefined);

  // Template dialog states
  const [templatePickerOpen, setTemplatePickerOpen] = useState(false);
  const [manageTemplatesOpen, setManageTemplatesOpen] = useState(false);
  const [createTemplateOpen, setCreateTemplateOpen] = useState(false);
  const [editTemplateOpen, setEditTemplateOpen] = useState(false);
  const [templateToEdit, setTemplateToEdit] = useState<TaskTemplate | null>(null);
  const [initialValues, setInitialValues] = useState<Partial<CreateTaskDTO> | undefined>(undefined);
  const [templateName, setTemplateName] = useState<string | undefined>(undefined);

  useEffect(() => {
    // Development mode: Auto-login with mock user
    if (process.env.NODE_ENV === 'development' && typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (!token) {
        // Create mock user for development
        const mockUser = {
          id: 'dev-user-123',
          email: 'admin@taskflow.dev',
          name: 'Admin User',
        };
        localStorage.setItem('token', 'dev-mock-token');
        setMockUser(mockUser);
        return;
      }
    }
    checkAuth();
  }, [checkAuth, setMockUser]);

  useEffect(() => {
    if (!user && typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (!token && process.env.NODE_ENV !== 'development') {
        router.push('/login');
      }
    }
  }, [user, router]);

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-gray-50 dark:bg-gray-950 overflow-hidden">
      {/* Sidebar */}
      <div className="w-80 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col flex-shrink-0">
        <div className="p-6 border-b border-gray-200 dark:border-gray-800 flex-shrink-0">
          <h1 className="text-2xl font-bold text-primary">TaskFlow</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Intelligent Prioritization
          </p>
        </div>

        {/* Scrollable middle section */}
        <div className="flex-1 overflow-y-auto min-h-0">
          <nav className="p-4 space-y-2">
            {navigation.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={`
                    block px-4 py-2 rounded-lg transition-colors
                    ${
                      isActive
                        ? 'bg-primary text-primary-foreground'
                        : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                    }
                  `}
                >
                  {item.name}
                </Link>
              );
            })}
          </nav>

          {/* Templates Section */}
          <div className="px-4 py-2 border-t border-gray-200 dark:border-gray-800">
            <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-2 px-2">
              Templates
            </p>
            <div className="space-y-1">
              <Button
                variant="ghost"
                className="w-full justify-start gap-2 h-9 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => setTemplatePickerOpen(true)}
              >
                <FileText className="h-4 w-4" />
                Create from Template
              </Button>
              <Button
                variant="ghost"
                className="w-full justify-start gap-2 h-9 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
                onClick={() => setManageTemplatesOpen(true)}
              >
                <Settings className="h-4 w-4" />
                Manage Templates
              </Button>
            </div>
          </div>

          {/* Pomodoro Timer */}
          <div className="px-4 py-2 border-t border-gray-200 dark:border-gray-800">
            <PomodoroWidget />
          </div>

          {/* Gamification Progress */}
          <div className="px-4 py-2 border-t border-gray-200 dark:border-gray-800">
            <GamificationWidget />
          </div>

          {/* Calendar */}
          <div className="px-4 py-4 border-t border-gray-200 dark:border-gray-800">
            <Calendar
              onTaskClick={(taskId) => {
                // Navigate to dashboard with task selected
                router.push(`/dashboard?taskId=${taskId}`);
              }}
              onCreateTask={(dueDate) => {
                setInitialDueDate(dueDate);
                setCreateDialogOpen(true);
              }}
            />
          </div>
        </div>

        <div className="p-4 border-t border-gray-200 dark:border-gray-800 flex-shrink-0">
          <div className="mb-3">
            <p className="text-sm font-medium">{user.name}</p>
            <p className="text-xs text-muted-foreground">{user.email}</p>
          </div>
          <div className="flex gap-2 mb-2">
            <ThemeToggle />
            <Button
              variant="outline"
              onClick={handleLogout}
              className="flex-1 h-10 transition-all hover:scale-105 hover:shadow-md cursor-pointer"
            >
              Sign Out
            </Button>
          </div>
        </div>
      </div>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        <div className="p-8">
          {children}
        </div>
      </main>

      {/* Create Task Dialog */}
      <CreateTaskDialog
        open={createDialogOpen}
        onOpenChange={(open) => {
          setCreateDialogOpen(open);
          if (!open) {
            setInitialDueDate(undefined);
            setInitialValues(undefined);
            setTemplateName(undefined);
          }
        }}
        initialDueDate={initialDueDate}
        initialValues={initialValues}
        templateName={templateName}
      />

      {/* Template Picker Dialog */}
      <TemplatePickerDialog
        open={templatePickerOpen}
        onOpenChange={setTemplatePickerOpen}
        onSelectTemplate={(formValues, template) => {
          setInitialValues(formValues);
          setTemplateName(template.name);
          setCreateDialogOpen(true);
        }}
      />

      {/* Manage Templates Dialog */}
      <ManageTemplatesDialog
        open={manageTemplatesOpen}
        onOpenChange={setManageTemplatesOpen}
        onEditTemplate={(template) => {
          setTemplateToEdit(template);
          setEditTemplateOpen(true);
        }}
        onCreateTemplate={() => {
          setCreateTemplateOpen(true);
        }}
      />

      {/* Create Template Dialog */}
      <CreateTemplateDialog
        open={createTemplateOpen}
        onOpenChange={setCreateTemplateOpen}
      />

      {/* Edit Template Dialog */}
      <EditTemplateDialog
        open={editTemplateOpen}
        onOpenChange={(open) => {
          setEditTemplateOpen(open);
          if (!open) setTemplateToEdit(null);
        }}
        template={templateToEdit}
      />
    </div>
  );
}
