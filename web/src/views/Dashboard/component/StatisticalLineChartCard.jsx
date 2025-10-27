import PropTypes from 'prop-types';
import { useTheme } from '@mui/material/styles';
import { LineChart, Line, ResponsiveContainer } from 'recharts';

// shadcn components
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Icon } from '@iconify/react';
import { renderNumber } from 'utils/common';
import { cn } from '@/lib/utils';

// 获取图表颜色 - 使用 CSS 变量
const getChartColor = (type = 'default') => {
    switch (type) {
        case 'request':
            return 'var(--chart-1)'; // 珊瑚色
        case 'quota':
            return 'var(--chart-4)'; // 黄色
        case 'token':
            return 'var(--chart-5)'; // 橙色
        default:
            return 'var(--chart-1)';
    }
};

// 获取卡片样式
const getCardStyles = (type = 'default') => {
    switch (type) {
        case 'request':
            return 'border-l-4 border-l-blue-500';
        case 'quota':
            return 'border-l-4 border-l-amber-500';
        case 'token':
            return 'border-l-4 border-l-rose-500';
        case 'rpm':
            return 'border-l-4 border-l-emerald-500';
        default:
            return '';
    }
};

// ==============================|| DASHBOARD - TOTAL ORDER LINE CHART CARD ||============================== //

const StatisticalLineChartCard = ({ isLoading, title, chartData, todayValue, lastDayValue, type = 'default' }) => {
    // 获取趋势图标
    const getTrendIcon = (percentChange) => {
        if (percentChange > 0) return 'mdi:trending-up';
        if (percentChange < 0) return 'mdi:trending-down';
        return 'mdi:trending-neutral';
    };

    // 获取趋势颜色 - 使用 CSS 变量
    const getTrendColor = (percentChange) => {
        if (percentChange > 0) return 'var(--chart-5)'; // 橙色
        if (percentChange < 0) return 'var(--chart-2)'; // 青色
        return 'var(--chart-1)'; // 珊瑚色
    };

    // 计算百分比变化
    const getPercentChange = () => {
        const todayValueNum = parseFloat((todayValue || '0').toString().replace('$', ''));
        const lastDayValueNum = parseFloat((lastDayValue || '0').toString().replace('$', ''));

        if (todayValueNum === 0 && lastDayValueNum === 0) return 0;
        if (todayValueNum === 0 && lastDayValueNum > 0) return -100;
        if (todayValueNum > 0 && lastDayValueNum === 0) return 100;
        return Math.round(((todayValueNum - lastDayValueNum) / lastDayValueNum) * 100);
    };

    const percentChange = lastDayValue !== undefined ? getPercentChange() : 0;
    const trendIcon = getTrendIcon(percentChange);
    const trendColor = getTrendColor(percentChange);

    // 转换图表数据为 Recharts 格式
    const rechartsData = chartData?.series?.[0]?.data || [];

    return (
        <>
            {isLoading ? (
                <Card className={cn('h-36', getCardStyles(type))}>
                    <CardContent className="p-4">
                        <Skeleton className="h-4 w-20 mb-2" />
                        <Skeleton className="h-8 w-16 mb-4" />
                        <Skeleton className="h-3 w-24 mb-4" />
                        <Skeleton className="h-12 w-full" />
                    </CardContent>
                </Card>
            ) : (
                <Card className={cn('h-36 flex flex-col transition-all duration-200 hover:shadow-md', getCardStyles(type))}>
                    <CardContent className="p-4 flex-1 flex flex-col">
                        <div className="flex justify-between items-center mb-1">
                            <div>
                                <h3 className="text-2xl font-medium text-foreground">{renderNumber(todayValue || 0)}</h3>
                            </div>
                            {lastDayValue !== undefined && (
                                <div className="flex items-center bg-muted rounded-xl py-1 px-2">
                                    <Icon icon={trendIcon} className="w-4 h-4 mr-1" style={{ color: trendColor }} />
                                    <span className="text-xs font-medium" style={{ color: trendColor }}>
                                        {`${Math.abs(percentChange)}%`}
                                    </span>
                                </div>
                            )}
                        </div>

                        <p className="text-xs text-muted-foreground mb-4">{title}</p>

                        <div className="mt-auto h-12 w-full">
                            {rechartsData.length > 0 && (
                                <ResponsiveContainer width="100%" height="100%">
                                    <LineChart data={rechartsData}>
                                        <Line
                                            type="monotone"
                                            dataKey="y"
                                            stroke={getChartColor(type)}
                                            strokeWidth={2}
                                            dot={false}
                                            activeDot={{ r: 3 }}
                                        />
                                    </LineChart>
                                </ResponsiveContainer>
                            )}
                        </div>
                    </CardContent>
                </Card>
            )}
        </>
    );
};

StatisticalLineChartCard.propTypes = {
    isLoading: PropTypes.bool,
    title: PropTypes.string,
    chartData: PropTypes.oneOfType([PropTypes.array, PropTypes.object]),
    todayValue: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    lastDayValue: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    type: PropTypes.string
};

export default StatisticalLineChartCard;
