import PropTypes from 'prop-types';
import React from 'react';
import { IconTrendingUp, IconTrendingDown, IconMinus } from '@tabler/icons-react';

// shadcn components
import { Card, CardHeader, CardTitle, CardDescription, CardAction, CardFooter } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { renderNumber, renderQuota } from 'utils/common';

// 获取趋势图标
const getTrendIcon = (trend) => {
    if (trend > 0) return IconTrendingUp;
    if (trend < 0) return IconTrendingDown;
    return IconMinus;
};

// 获取趋势颜色
const getTrendColor = (trend) => {
    if (trend > 0) return 'var(--chart-5)'; // 橙色 - 上升
    if (trend < 0) return 'var(--chart-2)'; // 青色 - 下降
    return 'var(--chart-1)'; // 珊瑚色 - 持平
};

// 根据指标类型获取渲染函数
const getValueRenderer = (type) => {
    switch (type) {
        case 'quota':
            return renderQuota; // 配额使用 renderQuota
        case 'request':
        case 'token':
        case 'rpm':
        default:
            return renderNumber; // 其他类型使用 renderNumber
    }
};

const MetricCard = ({ type = 'default', title, value, trend = 0, isLoading = false }) => {
    const TrendIcon = getTrendIcon(trend);
    const trendColor = getTrendColor(trend);
    const valueRenderer = getValueRenderer(type);

    // 处理特殊字符串值（如无限额度的 '∞'）
    const displayValue = typeof value === 'string' && value === '∞' ? '∞' : valueRenderer(value);

    // 获取趋势描述文案
    const getTrendDescription = (trend, type) => {
        if (trend > 0) {
            switch (type) {
                case 'request':
                    return '请求量持续增长';
                case 'quota':
                    return '配额使用量上升';
                case 'token':
                    return 'Token 消耗增加';
                case 'rpm':
                    return '请求频率提升';
                default:
                    return '数据呈上升趋势';
            }
        } else if (trend < 0) {
            switch (type) {
                case 'request':
                    return '请求量有所下降';
                case 'quota':
                    return '配额使用量减少';
                case 'token':
                    return 'Token 消耗降低';
                case 'rpm':
                    return '请求频率下降';
                default:
                    return '数据呈下降趋势';
            }
        } else {
            switch (type) {
                case 'request':
                    return '请求量保持稳定';
                case 'quota':
                    return '配额使用平稳';
                case 'token':
                    return 'Token 消耗稳定';
                case 'rpm':
                    return '请求频率稳定';
                default:
                    return '数据保持稳定';
            }
        }
    };

    // 获取底部描述文案
    const getFooterDescription = (type) => {
        switch (type) {
            case 'request':
                return '近7日请求趋势';
            case 'quota':
                return '配额使用情况';
            case 'token':
                return 'Token 消耗统计';
            case 'rpm':
                return '每分钟请求数';
            default:
                return '数据统计信息';
        }
    };

    return (
        <>
            {isLoading ? (
                <Card className="@container/card">
                    <CardHeader>
                        <Skeleton className="h-4 w-20 mb-2" />
                        <Skeleton className="h-8 w-16 mb-4" />
                        <Skeleton className="h-6 w-16 mb-4" />
                    </CardHeader>
                    <CardFooter className="flex-col items-start gap-1.5 text-sm">
                        <Skeleton className="h-4 w-32 mb-1" />
                        <Skeleton className="h-3 w-24" />
                    </CardFooter>
                </Card>
            ) : (
                <Card className="@container/card">
                    <CardHeader>
                        <CardDescription>{title}</CardDescription>
                        <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">{displayValue}</CardTitle>
                        {trend !== 0 && (
                            <CardAction>
                                <Badge variant="outline">
                                    <TrendIcon className="size-4" style={{ color: trendColor }} />
                                    {`${Math.abs(trend)}%`}
                                </Badge>
                            </CardAction>
                        )}
                    </CardHeader>
                    <CardFooter className="flex-col items-start gap-1.5 text-sm">
                        <div className="line-clamp-1 flex gap-2 font-medium">
                            {getTrendDescription(trend, type)} <TrendIcon className="size-4" style={{ color: trendColor }} />
                        </div>
                        <div className="text-muted-foreground">{getFooterDescription(type)}</div>
                    </CardFooter>
                </Card>
            )}
        </>
    );
};

MetricCard.propTypes = {
    type: PropTypes.oneOf(['request', 'quota', 'token', 'rpm', 'default']),
    title: PropTypes.string.isRequired,
    value: PropTypes.oneOfType([PropTypes.number, PropTypes.string]).isRequired,
    trend: PropTypes.number,
    isLoading: PropTypes.bool
};

export default MetricCard;
