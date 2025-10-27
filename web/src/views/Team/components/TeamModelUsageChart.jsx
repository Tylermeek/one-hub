import PropTypes from 'prop-types';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

const TeamModelUsageChart = ({ data = [], isLoading = false, className = '' }) => {
    // 图表颜色
    const colors = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)'];

    const total = data && Array.isArray(data) ? data.reduce((sum, item) => sum + item.value, 0) : 0;

    // 自定义 Tooltip
    const CustomTooltip = ({ active, payload }) => {
        if (active && payload && payload.length) {
            const data = payload[0];
            return (
                <div className="bg-background border border-border rounded-lg p-3 shadow-lg">
                    <p className="text-sm font-medium">
                        {data.name}: <span className="font-bold">{data.value.toLocaleString()}</span>
                    </p>
                </div>
            );
        }
        return null;
    };

    // 自定义 Legend
    const CustomLegend = ({ payload }) => {
        return (
            <div className="flex flex-wrap justify-center gap-4 mt-4">
                {payload?.map((entry, index) => (
                    <div key={index} className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-full" style={{ backgroundColor: entry.color }} />
                        <span className="text-sm text-muted-foreground">{entry.value}</span>
                        <span className="text-sm font-medium">{entry.payload.value.toLocaleString()}</span>
                    </div>
                ))}
            </div>
        );
    };

    return (
        <Card className={cn('@container/card h-auto', className)}>
            <CardHeader>
                <CardTitle className="text-lg font-semibold">模型使用分布</CardTitle>
            </CardHeader>
            <CardContent>
                {isLoading ? (
                    <Skeleton className="h-72 w-full" />
                ) : data.length > 0 ? (
                    <div className="space-y-4">
                        <div className="relative">
                            <ResponsiveContainer width="100%" height={200}>
                                <PieChart>
                                    <Pie data={data} cx="50%" cy="50%" innerRadius={45} outerRadius={85} paddingAngle={2} dataKey="value">
                                        {data.map((entry, index) => (
                                            <Cell
                                                key={`cell-${index}`}
                                                fill={colors[index % colors.length]}
                                                className="transition-all duration-200 hover:opacity-80"
                                            />
                                        ))}
                                    </Pie>
                                    <Tooltip content={<CustomTooltip />} />
                                </PieChart>
                            </ResponsiveContainer>
                            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                                <div className="text-center">
                                    <p className="text-sm text-muted-foreground">总计</p>
                                    <p className="text-2xl font-bold">{total.toLocaleString()}</p>
                                </div>
                            </div>
                        </div>
                        <Legend content={<CustomLegend />} />
                    </div>
                ) : (
                    <div className="h-72 flex items-center justify-center rounded-lg bg-muted/50">
                        <p className="text-lg font-medium text-muted-foreground">暂无数据</p>
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

TeamModelUsageChart.propTypes = {
    data: PropTypes.arrayOf(
        PropTypes.shape({
            name: PropTypes.string.isRequired,
            value: PropTypes.number.isRequired
        })
    ),
    isLoading: PropTypes.bool,
    className: PropTypes.string
};

export default TeamModelUsageChart;
