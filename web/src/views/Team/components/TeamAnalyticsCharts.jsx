import React from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { renderQuota } from 'utils/common';

/**
 * 自定义 Tooltip 组件
 */
const CustomTooltip = ({ active, payload, label }) => {
    if (!active || !payload || !payload.length) return null;

    return (
        <div className="bg-background border rounded-lg shadow-lg p-3">
            <p className="text-sm font-medium mb-2">{label}</p>
            {payload.map((entry, index) => (
                <div key={index} className="flex items-center gap-2 text-xs">
                    <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: entry.color }} />
                    <span className="text-muted-foreground">{entry.name}:</span>
                    <span className="font-medium">{entry.value}</span>
                </div>
            ))}
        </div>
    );
};

/**
 * 图表色彩数组
 */
const CHART_COLORS = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)'];

/**
 * 模型使用分布饼图
 */
export function ModelDistributionChart({ data, isLoading }) {
    const { t } = useTranslation();

    if (isLoading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>{t('team_analytics.model_distribution')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <Skeleton className="h-80 w-full" />
                </CardContent>
            </Card>
        );
    }

    if (!data || data.length === 0) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>{t('team_analytics.model_distribution')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="h-80 flex items-center justify-center text-muted-foreground">{t('common.no_data')}</div>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('team_analytics.model_distribution')}</CardTitle>
            </CardHeader>
            <CardContent>
                <ResponsiveContainer width="100%" height={320}>
                    <PieChart>
                        <Pie data={data} cx="50%" cy="50%" innerRadius={60} outerRadius={120} paddingAngle={2} dataKey="value">
                            {data.map((entry, index) => (
                                <Cell
                                    key={`cell-${index}`}
                                    fill={CHART_COLORS[index % CHART_COLORS.length]}
                                    className="transition-all duration-200 hover:opacity-80"
                                />
                            ))}
                        </Pie>
                        <Tooltip content={<CustomTooltip />} />
                        <Legend
                            verticalAlign="bottom"
                            height={36}
                            iconType="circle"
                            formatter={(value, entry) => (
                                <span className="text-sm">
                                    {value} ({entry.payload.percentage}%)
                                </span>
                            )}
                        />
                    </PieChart>
                </ResponsiveContainer>
            </CardContent>
        </Card>
    );
}

ModelDistributionChart.propTypes = {
    data: PropTypes.arrayOf(
        PropTypes.shape({
            name: PropTypes.string,
            value: PropTypes.number,
            percentage: PropTypes.number
        })
    ),
    isLoading: PropTypes.bool
};

/**
 * 时段分布柱状图
 */
export function HourlyDistributionChart({ data, isLoading }) {
    const { t } = useTranslation();

    if (isLoading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>{t('team_analytics.hourly_distribution')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <Skeleton className="h-80 w-full" />
                </CardContent>
            </Card>
        );
    }

    if (!data || data.length === 0) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>{t('team_analytics.hourly_distribution')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="h-80 flex items-center justify-center text-muted-foreground">{t('common.no_data')}</div>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('team_analytics.hourly_distribution')}</CardTitle>
            </CardHeader>
            <CardContent>
                <ResponsiveContainer width="100%" height={320}>
                    <BarChart data={data}>
                        <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                        <XAxis dataKey="hour" className="text-xs" tick={{ fill: 'var(--muted-foreground)' }} />
                        <YAxis className="text-xs" tick={{ fill: 'var(--muted-foreground)' }} />
                        <Tooltip content={<CustomTooltip />} />
                        <Bar
                            dataKey="requests"
                            fill="var(--chart-1)"
                            radius={[4, 4, 0, 0]}
                            className="transition-all duration-200 hover:opacity-80"
                        />
                    </BarChart>
                </ResponsiveContainer>
            </CardContent>
        </Card>
    );
}

HourlyDistributionChart.propTypes = {
    data: PropTypes.arrayOf(
        PropTypes.shape({
            hour: PropTypes.string,
            requests: PropTypes.number
        })
    ),
    isLoading: PropTypes.bool
};

/**
 * 终端来源分布饼图
 */
export function SourceDistributionChart({ data, isLoading }) {
    const { t } = useTranslation();

    if (isLoading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>{t('team_analytics.source_distribution')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <Skeleton className="h-80 w-full" />
                </CardContent>
            </Card>
        );
    }

    if (!data || data.length === 0) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle>{t('team_analytics.source_distribution')}</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="h-80 flex items-center justify-center text-muted-foreground">{t('common.no_data')}</div>
                </CardContent>
            </Card>
        );
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>{t('team_analytics.source_distribution')}</CardTitle>
            </CardHeader>
            <CardContent>
                <ResponsiveContainer width="100%" height={320}>
                    <PieChart>
                        <Pie data={data} cx="50%" cy="50%" innerRadius={70} outerRadius={110} paddingAngle={3} dataKey="value">
                            {data.map((entry, index) => (
                                <Cell
                                    key={`cell-${index}`}
                                    fill={CHART_COLORS[index % CHART_COLORS.length]}
                                    className="transition-all duration-200 hover:opacity-80"
                                />
                            ))}
                        </Pie>
                        <Tooltip content={<CustomTooltip />} />
                        <Legend
                            verticalAlign="bottom"
                            height={36}
                            iconType="circle"
                            formatter={(value, entry) => (
                                <span className="text-sm">
                                    {value} ({entry.payload.percentage}%)
                                </span>
                            )}
                        />
                    </PieChart>
                </ResponsiveContainer>
            </CardContent>
        </Card>
    );
}

SourceDistributionChart.propTypes = {
    data: PropTypes.arrayOf(
        PropTypes.shape({
            name: PropTypes.string,
            value: PropTypes.number,
            percentage: PropTypes.number
        })
    ),
    isLoading: PropTypes.bool
};

/**
 * 综合分析图表组件
 */
function TeamAnalyticsCharts({ analyticsData, isLoading, className }) {
    return (
        <div className={className}>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <ModelDistributionChart data={analyticsData?.model_distribution} isLoading={isLoading} />
                <HourlyDistributionChart data={analyticsData?.hourly_distribution} isLoading={isLoading} />
                <SourceDistributionChart data={analyticsData?.source_distribution} isLoading={isLoading} />
            </div>
        </div>
    );
}

TeamAnalyticsCharts.propTypes = {
    analyticsData: PropTypes.shape({
        model_distribution: PropTypes.array,
        hourly_distribution: PropTypes.array,
        source_distribution: PropTypes.array
    }),
    isLoading: PropTypes.bool,
    className: PropTypes.string
};

export default TeamAnalyticsCharts;
