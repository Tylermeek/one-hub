import PropTypes from 'prop-types';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';

// 图表颜色配置
const CHART_COLORS = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)'];

// 自定义 Tooltip
const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
        const data = payload[0].payload;
        return (
            <div className="bg-background border border-border rounded-lg p-3 shadow-lg">
                <p className="text-sm font-medium text-foreground mb-2">{label}</p>
                <div className="space-y-1 text-sm">
                    <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: payload[0].color }} />
                        <span className="text-muted-foreground">消费:</span>
                        <span className="font-medium text-foreground">${data.quota.toFixed(2)}</span>
                    </div>
                </div>
            </div>
        );
    }
    return null;
};

// 自定义 X 轴标签
const CustomXAxisTick = ({ x, y, payload }) => {
    return (
        <g transform={`translate(${x},${y})`}>
            <text x={0} y={0} dy={16} textAnchor="middle" fill="var(--muted-foreground)" fontSize="12">
                {payload.value}
            </text>
        </g>
    );
};

const MemberRankingChart = ({ data = [], isLoading = false, anonymize = false, className = '' }) => {
    // 转换数据格式并排序
    const chartData = data
        .map((item, index) => ({
            name: anonymize ? `成员${String.fromCharCode(65 + index)}` : item.name,
            quota: item.quota || 0,
            color: CHART_COLORS[index % CHART_COLORS.length]
        }))
        .sort((a, b) => b.quota - a.quota)
        .slice(0, 5); // 只取前5个

    return (
        <Card className={cn('@container/card h-auto', className)}>
            <CardHeader>
                <CardTitle className="flex items-center justify-between text-lg">
                    <span>成员消费排行</span>
                    <Badge variant="secondary" className="text-xs">
                        Top 5
                    </Badge>
                </CardTitle>
            </CardHeader>
            <CardContent className="px-2">
                {isLoading ? (
                    <div className="h-[250px] w-full">
                        <Skeleton className="h-full w-full" />
                    </div>
                ) : chartData.length === 0 ? (
                    <div className="h-[250px] w-full flex items-center justify-center">
                        <div className="text-center text-muted-foreground">
                            <p className="text-sm">暂无数据</p>
                        </div>
                    </div>
                ) : (
                    <div className="h-[250px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart
                                data={chartData}
                                layout="horizontal"
                                margin={{
                                    top: 0,
                                    right: 0,
                                    left: 0,
                                    bottom: 0
                                }}
                            >
                                <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
                                <XAxis type="number" className="text-xs text-muted-foreground" />
                                <YAxis type="category" dataKey="name" tick={<CustomXAxisTick />} width={60} />
                                <Tooltip content={<CustomTooltip />} />

                                <Bar dataKey="quota" radius={[0, 4, 4, 0]} className="transition-all duration-200 hover:opacity-80">
                                    {chartData.map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={entry.color} />
                                    ))}
                                </Bar>
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                )}

                {/* 移动端列表显示 */}
                <div className="mt-4 space-y-2 md:hidden">
                    {chartData.map((item, index) => (
                        <div key={item.name} className="flex items-center justify-between p-2 bg-muted/50 rounded-lg">
                            <div className="flex items-center gap-2">
                                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                                <span className="text-sm font-medium">{item.name}</span>
                            </div>
                            <div className="text-right">
                                <div className="text-sm font-medium">${item.quota.toFixed(2)}</div>
                            </div>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
};

MemberRankingChart.propTypes = {
    data: PropTypes.arrayOf(
        PropTypes.shape({
            name: PropTypes.string.isRequired,
            quota: PropTypes.number.isRequired
        })
    ),
    isLoading: PropTypes.bool,
    anonymize: PropTypes.bool, // 是否匿名显示（用于 Member 角色）
    className: PropTypes.string
};

export default MemberRankingChart;
