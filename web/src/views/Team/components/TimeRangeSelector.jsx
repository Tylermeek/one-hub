import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { Calendar } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Calendar as CalendarComponent } from '@/components/ui/calendar';
import { cn } from '@/lib/utils';

/**
 * 预设时间范围选项
 */
const PRESET_RANGES = [
  { value: 'today', label: 'time_range.today' },
  { value: '7d', label: 'time_range.last_7_days' },
  { value: '30d', label: 'time_range.last_30_days' },
  { value: 'this_month', label: 'time_range.this_month' },
  { value: 'last_month', label: 'time_range.last_month' },
  { value: 'custom', label: 'time_range.custom' }
];

/**
 * 时间范围选择器组件
 */
function TimeRangeSelector({ value, onChange, className }) {
  const { t } = useTranslation();
  const [isCustom, setIsCustom] = useState(value === 'custom');
  const [customRange, setCustomRange] = useState({
    from: null,
    to: null
  });

  /**
   * 处理预设范围选择
   */
  const handlePresetChange = (newValue) => {
    if (newValue === 'custom') {
      setIsCustom(true);
    } else {
      setIsCustom(false);
      onChange && onChange(newValue, null);
    }
  };

  /**
   * 处理自定义日期选择
   */
  const handleCustomRangeChange = (range) => {
    setCustomRange(range);
    if (range?.from && range?.to) {
      onChange && onChange('custom', range);
    }
  };

  /**
   * 格式化日期显示
   */
  const formatDateRange = () => {
    if (!customRange.from) return t('time_range.select_range');
    if (!customRange.to) return t('time_range.select_end_date');

    const from = new Date(customRange.from).toLocaleDateString();
    const to = new Date(customRange.to).toLocaleDateString();
    return `${from} - ${to}`;
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <Calendar className="h-4 w-4 text-muted-foreground" />

      {!isCustom ? (
        <Select value={value} onValueChange={handlePresetChange}>
          <SelectTrigger className="w-48">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {PRESET_RANGES.map((range) => (
              <SelectItem key={range.value} value={range.value}>
                {t(range.label)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ) : (
        <div className="flex items-center gap-2">
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" className="w-64 justify-start text-left">
                {formatDateRange()}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <CalendarComponent mode="range" selected={customRange} onSelect={handleCustomRangeChange} numberOfMonths={2} initialFocus />
            </PopoverContent>
          </Popover>

          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setIsCustom(false);
              onChange && onChange('7d', null);
            }}
          >
            {t('common.cancel')}
          </Button>
        </div>
      )}
    </div>
  );
}

TimeRangeSelector.propTypes = {
  value: PropTypes.string.isRequired,
  onChange: PropTypes.func,
  className: PropTypes.string
};

export default TimeRangeSelector;
