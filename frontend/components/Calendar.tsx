'use client';

import { useState } from 'react';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, startOfWeek, endOfWeek, isSameMonth, isToday } from 'date-fns';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useCalendarTasks } from '@/hooks/useTasks';
import { CalendarDayData } from '@/lib/api';

interface CalendarProps {
  onDayClick?: (date: Date, dayData?: CalendarDayData) => void;
}

export function Calendar({ onDayClick }: CalendarProps) {
  const [currentMonth, setCurrentMonth] = useState(new Date());

  // Calculate date range for current month
  const monthStart = startOfMonth(currentMonth);
  const monthEnd = endOfMonth(currentMonth);
  const calendarStart = startOfWeek(monthStart, { weekStartsOn: 1 }); // Monday start
  const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });

  // Get all days to display
  const days = eachDayOfInterval({ start: calendarStart, end: calendarEnd });

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
    <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800 p-3 w-fit">
      {/* Month Navigation */}
      <div className="flex items-center justify-between mb-3 gap-2">
        <button
          onClick={handlePreviousMonth}
          className="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
          aria-label="Previous month"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        <h3 className="text-sm font-semibold whitespace-nowrap">
          {format(currentMonth, 'MMM yyyy')}
        </h3>

        <button
          onClick={handleNextMonth}
          className="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
          aria-label="Next month"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>

      {/* Calendar Table */}
      {isLoading ? (
        <div className="flex items-center justify-center h-48">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-gray-900 dark:border-white"></div>
        </div>
      ) : (
        <table className="border-collapse">
          {/* Day Headers */}
          <thead>
            <tr>
              {['M', 'T', 'W', 'T', 'F', 'S', 'S'].map((day, i) => (
                <th
                  key={i}
                  className="text-[10px] font-medium text-gray-500 dark:text-gray-400 pb-1 text-center w-8"
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

                  return (
                    <td
                      key={dayIdx}
                      className="p-0 text-center relative"
                    >
                      <button
                        onClick={() => handleDayClick(day)}
                        className={`
                          w-8 h-8 flex items-center justify-center rounded-md text-xs
                          transition-colors relative
                          ${isCurrentMonth
                            ? 'text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800'
                            : 'text-gray-300 dark:text-gray-600'
                          }
                          ${isCurrentDay
                            ? 'bg-blue-100 dark:bg-blue-900/30 font-bold'
                            : ''
                          }
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
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
