import PropTypes from 'prop-types';
import React, { useState } from 'react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { useTranslation } from 'react-i18next';

// shadcn components
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardAction } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { cn } from '@/lib/utils';

// 自定义 Tooltip 组件
const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
        return (
            <div className="bg-background border border-border rounded-lg p-3 shadow-lg">
                <p className="text-sm font-medium text-foreground mb-2">{label}</p>
                {payload.map((entry, index) => (
                    <div key={index} className="flex items-center gap-2 text-sm">
                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: entry.color }} />
                        <span className="text-muted-foreground">{entry.name}:</span>
                        <span className="font-medium text-foreground">{entry.name === '消费额度' ? `$${entry.value}` : entry.value}</span>
                    </div>
                ))}
            </div>
        );
    }
    return null;
};

const WeeklyTrendChart = ({ data = [], isLoading = false, className = '' }) => {
    const { t } = useTranslation();
    const [timeRange, setTimeRange] = useState('7d');

    // 根据时间范围过滤数据
    const getFilteredData = (originalData, range) => {
        if (!originalData || originalData.length === 0) return [];

        const now = new Date();
        let daysToSubtract = 7;

        switch (range) {
            case '30d':
                daysToSubtract = 30;
                break;
            case '90d':
                daysToSubtract = 90;
                break;
            default:
                daysToSubtract = 7;
        }

        const startDate = new Date(now);
        startDate.setDate(startDate.getDate() - daysToSubtract);

        return originalData.filter((item) => {
            const itemDate = new Date(item.date);
            return itemDate >= startDate;
        });
    };

    // 转换数据格式
    const filteredData = getFilteredData(data, timeRange);
    const chartData = filteredData.map((item) => ({
        date: item.date,
        requests: item.requests || 0,
        quota: item.quota || 0
    }));

    // 格式化日期显示
    const formatDate = (dateStr) => {
        const date = new Date(dateStr);
        return date.toLocaleDateString('zh-CN', {
            month: 'short',
            day: 'numeric'
        });
    };

    return (
        <Card className={cn('@container/card h-auto', className)}>
            <CardHeader>
                <CardTitle>{t('dashboard_index.week_consumption_trend')}</CardTitle>
                <CardDescription>
                    <span className="hidden @[540px]/card:block">
                        最近 {timeRange === '7d' ? '7' : timeRange === '30d' ? '30' : '90'} 天的消费趋势
                    </span>
                    <span className="@[540px]/card:hidden">最近 {timeRange === '7d' ? '7' : timeRange === '30d' ? '30' : '90'} 天</span>
                </CardDescription>
                <CardAction>
                    <Select value={timeRange} onValueChange={setTimeRange}>
                        <SelectTrigger
                            className="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate"
                            size="sm"
                            aria-label="选择时间范围"
                        >
                            <SelectValue placeholder="最近 7 天" />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                            <SelectItem value="7d" className="rounded-lg">
                                最近 7 天
                            </SelectItem>
                            <SelectItem value="30d" className="rounded-lg">
                                最近 30 天
                            </SelectItem>
                            <SelectItem value="90d" className="rounded-lg">
                                最近 3 个月
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </CardAction>
            </CardHeader>
            <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
                {isLoading ? (
                    <div className="h-[222px] w-full">
                        <Skeleton className="h-full w-full" />
                    </div>
                ) : (
                    <div className="h-[220px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart
                                data={chartData}
                                margin={{
                                    top: 0,
                                    right: 0,
                                    left: 0,
                                    bottom: 0
                                }}
                            >
                                <defs>
                                    <linearGradient id="fillRequests" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="var(--chart-1)" stopOpacity={1.0} />
                                        <stop offset="95%" stopColor="var(--chart-1)" stopOpacity={0.1} />
                                    </linearGradient>
                                    <linearGradient id="fillQuota" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="var(--chart-4)" stopOpacity={0.8} />
                                        <stop offset="95%" stopColor="var(--chart-4)" stopOpacity={0.1} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                                <XAxis
                                    dataKey="date"
                                    tickFormatter={formatDate}
                                    className="text-xs text-muted-foreground"
                                    tickLine={false}
                                    axisLine={false}
                                    tickMargin={8}
                                    minTickGap={32}
                                />
                                <YAxis yAxisId="left" className="text-xs text-muted-foreground" />
                                <YAxis yAxisId="right" orientation="right" className="text-xs text-muted-foreground" />
                                <Tooltip content={<CustomTooltip />} />
                                <Legend />

                                {/* 请求数区域图 */}
                                <Area
                                    yAxisId="left"
                                    type="monotone"
                                    dataKey="requests"
                                    stackId="1"
                                    stroke="var(--chart-1)"
                                    fill="url(#fillRequests)"
                                    name="请求数"
                                />

                                {/* 消费额度区域图 */}
                                <Area
                                    yAxisId="right"
                                    type="monotone"
                                    dataKey="quota"
                                    stackId="2"
                                    stroke="var(--chart-4)"
                                    fill="url(#fillQuota)"
                                    name="消费额度"
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

WeeklyTrendChart.propTypes = {
    data: PropTypes.arrayOf(
        PropTypes.shape({
            date: PropTypes.string.isRequired,
            requests: PropTypes.number,
            quota: PropTypes.number
        })
    ),
    isLoading: PropTypes.bool,
    className: PropTypes.string
};

export default WeeklyTrendChart;
