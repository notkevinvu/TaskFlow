'use client';

import { useState } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, startOfWeek, endOfWeek, isSameMonth, isToday } from 'date-fns';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useCalendarTasks } from '@/hooks/useTasks';
import { CalendarTaskPopover } from '@/components/CalendarTaskPopover';

interface CalendarProps {
  onTaskClick?: (taskId: string) => void;
  onCreateTask?: (dueDate: string) => void; // YYYY-MM-DD format
}

export function Calendar({ onTaskClick, onCreateTask }: CalendarProps) {
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [selectedDate, setSelectedDate] = useState<string | null>(null);

  // Calculate date range for current month
  const monthStart = startOfMonth(currentMonth);
  const monthEnd = endOfMonth(currentMonth);
  const calendarStart = startOfWeek(monthStart, { weekStartsOn: 0 }); // Sunday start
  const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 0 });

  // Get all days to display
  const days = eachDayOfInterval({ start: calendarStart, end: calendarEnd });

  // Fetch calendar data for entire visible range (includes prev/next month days)
  // Only fetch incomplete tasks to hide completed tasks from calendar view
  const { data: calendarData, isLoading, error } = useCalendarTasks({
    start_date: format(calendarStart, 'yyyy-MM-dd'),
    end_date: format(calendarEnd, 'yyyy-MM-dd'),
    status: 'todo', // Exclude completed tasks from calendar
  });

  const handlePreviousMonth = () => {
    setCurrentMonth(prev => new Date(prev.getFullYear(), prev.getMonth() - 1, 1));
  };

  const handleNextMonth = () => {
    setCurrentMonth(prev => new Date(prev.getFullYear(), prev.getMonth() + 1, 1));
  };

  const handleDayClick = (date: Date) => {
    const dateKey = format(date, 'yyyy-MM-dd');
    const dayData = calendarData?.dates[dateKey];

    if (dayData && dayData.tasks.length > 0) {
      // Day has tasks - toggle popover
      setSelectedDate(selectedDate === dateKey ? null : dateKey);
    } else {
      // Day has no tasks - trigger create dialog
      onCreateTask?.(dateKey);
    }
  };

  // Split days into weeks (7 days per week)
  const weeks: Date[][] = [];
  for (let i = 0; i < days.length; i += 7) {
    weeks.push(days.slice(i, i + 7));
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
        <p className="text-sm text-red-600 dark:text-red-400">Failed to load calendar</p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 p-4 w-full relative">
      {/* Loading Overlay */}
      {isLoading && (
        <div className="absolute inset-0 bg-white/50 dark:bg-gray-900/50 rounded-lg flex items-center justify-center z-10">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-gray-900 dark:border-white"></div>
        </div>
      )}

      {/* Month Navigation */}
      <div className="flex items-center justify-between mb-3">
        <button
          onClick={handlePreviousMonth}
          className="w-8 h-8 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-all hover:scale-105 hover:shadow-md cursor-pointer flex-shrink-0"
          aria-label="Previous month"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        <h3 className="text-sm font-semibold whitespace-nowrap flex-1 text-center">
          {format(currentMonth, 'MMM yyyy')}
        </h3>

        <button
          onClick={handleNextMonth}
          className="w-8 h-8 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-all hover:scale-105 hover:shadow-md cursor-pointer flex-shrink-0"
          aria-label="Next month"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>

      {/* Calendar Table */}
      <table className="border-collapse w-full">
          {/* Day Headers */}
          <thead>
            <tr>
              {['S', 'M', 'T', 'W', 'T', 'F', 'S'].map((day, i) => (
                <th
                  key={i}
                  className="text-sm font-semibold text-gray-900 dark:text-gray-100 pb-2 text-center w-8"
                >
                  {day}
                </th>
              ))}
            </tr>
          </thead>

          {/* Calendar Grid */}
          <tbody>
            {weeks.map((week, weekIdx) => (
              <tr key={weekIdx}>
                {week.map((day, dayIdx) => {
                  const dateKey = format(day, 'yyyy-MM-dd');
                  const dayData = calendarData?.dates[dateKey];
                  const isCurrentMonth = isSameMonth(day, currentMonth);
                  const isCurrentDay = isToday(day);
                  const hasTasks = dayData && dayData.tasks.length > 0;

                  const dayButton = (
                    <button
                      onClick={() => handleDayClick(day)}
                      className={`
                        w-8 h-8 flex items-center justify-center rounded text-sm
                        transition-all hover:scale-105 hover:shadow-md cursor-pointer relative
                        ${isCurrentMonth
                          ? 'text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800'
                          : 'text-gray-300 dark:text-gray-600 hover:text-gray-400'
                        }
                        ${isCurrentDay
                          ? 'bg-blue-100 dark:bg-blue-900/30 font-bold'
                          : ''
                        }
                        ${hasTasks ? 'font-semibold' : ''}
                      `}
                    >
                      {format(day, 'd')}

                      {/* Badge */}
                      {dayData && dayData.count > 0 && (
                        <span
                          className={`
                            absolute -top-0.5 -right-0.5 min-w-[14px] h-[14px] px-0.5
                            flex items-center justify-center
                            rounded-full text-[8px] font-bold text-white leading-none
                            ${dayData.badge_color === 'red' ? 'bg-red-500' : ''}
                            ${dayData.badge_color === 'yellow' ? 'bg-yellow-500' : ''}
                            ${dayData.badge_color === 'blue' ? 'bg-blue-500' : ''}
                          `}
                        >
                          {dayData.count}
                        </span>
                      )}
                    </button>
                  );

                  return (
                    <td
                      key={dayIdx}
                      className="p-0 relative w-8"
                    >
                      <div className="flex items-center justify-center">
                      {hasTasks ? (
                        <CalendarTaskPopover
                          date={day}
                          tasks={dayData.tasks}
                          trigger={dayButton}
                          open={selectedDate === dateKey}
                          onOpenChange={(open) => {
                            if (!open) setSelectedDate(null);
                          }}
                          onTaskClick={(taskId) => onTaskClick?.(taskId)}
                        />
                      ) : (
                        dayButton
                      )}
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
    </div>
  );
}
