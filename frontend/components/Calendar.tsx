'use client';

import { useState } from 'react';
import { DayPicker } from 'react-day-picker';
import { format, startOfMonth, endOfMonth } from 'date-fns';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useCalendarTasks } from '@/hooks/useTasks';
import { CalendarDayData } from '@/lib/api';
import 'react-day-picker/dist/style.css';

interface CalendarProps {
  onDayClick?: (date: Date, dayData?: CalendarDayData) => void;
}

export function Calendar({ onDayClick }: CalendarProps) {
  const [currentMonth, setCurrentMonth] = useState(new Date());

  // Calculate date range for current month
  const monthStart = startOfMonth(currentMonth);
  const monthEnd = endOfMonth(currentMonth);

  // Fetch calendar data
  const { data: calendarData, isLoading, error } = useCalendarTasks({
    start_date: format(monthStart, 'yyyy-MM-dd'),
    end_date: format(monthEnd, 'yyyy-MM-dd'),
  });

  const handlePreviousMonth = () => {
    setCurrentMonth(prev => new Date(prev.getFullYear(), prev.getMonth() - 1, 1));
  };

  const handleNextMonth = () => {
    setCurrentMonth(prev => new Date(prev.getFullYear(), prev.getMonth() + 1, 1));
  };

  const handleDayClick = (date: Date | undefined) => {
    if (!date) return;
    const dateKey = format(date, 'yyyy-MM-dd');
    const dayData = calendarData?.dates[dateKey];
    onDayClick?.(date, dayData);
  };

  // Get days with tasks for styling
  const getDayModifiers = () => {
    if (!calendarData) return {};

    const daysWithTasks: Date[] = [];
    Object.keys(calendarData.dates).forEach(dateKey => {
      const [year, month, day] = dateKey.split('-').map(Number);
      daysWithTasks.push(new Date(year, month - 1, day));
    });

    return { hasTasks: daysWithTasks };
  };

  if (error) {
    return (
      <div className="p-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
        <p className="text-red-600 dark:text-red-400">Failed to load calendar</p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 p-4">
      {/* Month Navigation */}
      <div className="flex items-center justify-between mb-4">
        <button
          onClick={handlePreviousMonth}
          className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
          aria-label="Previous month"
        >
          <ChevronLeft className="w-5 h-5" />
        </button>

        <h2 className="text-lg font-semibold">
          {format(currentMonth, 'MMMM yyyy')}
        </h2>

        <button
          onClick={handleNextMonth}
          className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-md transition-colors"
          aria-label="Next month"
        >
          <ChevronRight className="w-5 h-5" />
        </button>
      </div>

      {/* Calendar Grid */}
      {isLoading ? (
        <div className="flex items-center justify-center h-80">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-white"></div>
        </div>
      ) : (
        <div className="relative">
          <DayPicker
            mode="single"
            month={currentMonth}
            onMonthChange={setCurrentMonth}
            onDayClick={handleDayClick}
            modifiers={getDayModifiers()}
            classNames={{
              months: 'w-full',
              month: 'w-full',
              caption: 'hidden', // We use custom navigation
              table: 'w-full border-collapse',
              head_row: 'flex w-full',
              head_cell: 'flex-1 text-center text-sm font-medium text-gray-500 dark:text-gray-400 pb-2',
              row: 'flex w-full',
              cell: 'flex-1 aspect-square p-0 relative [&:has([aria-selected])]:bg-transparent',
              day: 'w-full h-full flex items-center justify-center rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer transition-colors text-sm',
              day_today: 'bg-blue-50 dark:bg-blue-900/30 font-bold',
              day_outside: 'text-gray-300 dark:text-gray-600',
              day_selected: 'bg-transparent',
            }}
          />

          {/* Task Badges Overlay */}
          {calendarData && (
            <div className="absolute inset-0 pointer-events-none">
              {Object.entries(calendarData.dates).map(([dateKey, dayData]) => {
                const [year, month, day] = dateKey.split('-').map(Number);
                const date = new Date(year, month - 1, day);

                // Skip if date is not in current month
                if (date.getMonth() !== currentMonth.getMonth()) return null;

                // Calculate position in grid
                const firstDay = new Date(currentMonth.getFullYear(), currentMonth.getMonth(), 1);
                const dayOfWeek = (date.getDay() + 6) % 7; // Adjust for Monday start
                const weekNumber = Math.floor((date.getDate() + firstDay.getDay() - 1) / 7);

                return (
                  <div
                    key={dateKey}
                    className="absolute pointer-events-none"
                    style={{
                      top: `calc(${weekNumber * 14.28}% + 2.5rem)`,
                      left: `calc(${dayOfWeek * 14.28}%)`,
                      width: '14.28%',
                      height: '14.28%',
                    }}
                  >
                    <div className="relative w-full h-full flex items-center justify-center">
                      <div
                        className={`
                          absolute -top-1 -right-1 min-w-[18px] h-[18px]
                          flex items-center justify-center
                          rounded-full text-[10px] font-semibold text-white
                          ${dayData.badge_color === 'red' ? 'bg-red-500' : ''}
                          ${dayData.badge_color === 'yellow' ? 'bg-yellow-500' : ''}
                          ${dayData.badge_color === 'blue' ? 'bg-blue-500' : ''}
                        `}
                      >
                        {dayData.count}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
