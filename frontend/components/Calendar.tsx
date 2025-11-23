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

  const handleDayClick = (date: Date) => {
    const dateKey = format(date, 'yyyy-MM-dd');
    const dayData = calendarData?.dates[dateKey];
    onDayClick?.(date, dayData);
  };

  // Custom day content renderer to show badges
  const renderDay = (date: Date) => {
    const dateKey = format(date, 'yyyy-MM-dd');
    const dayData = calendarData?.dates[dateKey];

    return (
      <div className="relative w-full h-full flex items-center justify-center">
        <span>{date.getDate()}</span>
        {dayData && dayData.count > 0 && (
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
        )}
      </div>
    );
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
        <DayPicker
          mode="single"
          month={currentMonth}
          onMonthChange={setCurrentMonth}
          onDayClick={handleDayClick}
          components={{
            Day: ({ date }) => renderDay(date),
          }}
          classNames={{
            months: 'w-full',
            month: 'w-full',
            caption: 'hidden', // We use custom navigation
            table: 'w-full border-collapse',
            head_row: 'flex w-full',
            head_cell: 'flex-1 text-center text-sm font-medium text-gray-500 dark:text-gray-400 pb-2',
            row: 'flex w-full',
            cell: 'flex-1 aspect-square p-0 relative',
            day: 'w-full h-full flex items-center justify-center rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer transition-colors text-sm',
            day_today: 'bg-blue-50 dark:bg-blue-900/30 font-bold',
            day_outside: 'text-gray-300 dark:text-gray-600',
          }}
        />
      )}
    </div>
  );
}
