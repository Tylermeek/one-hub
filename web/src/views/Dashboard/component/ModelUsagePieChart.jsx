import React from 'react';
import PropTypes from 'prop-types';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from 'react-i18next';

const ModelUsagePieChart = ({ isLoading, data }) => {
  const { t } = useTranslation();

  // 创建具有渐变色效果的调色板 - 使用 CSS 变量
  const generateColors = () => {
    const baseColors = [
      'var(--chart-1)', // 珊瑚色
      'var(--chart-2)', // 青色
      'var(--chart-3)', // 海军蓝
      'var(--chart-4)', // 黄色
      'var(--chart-5)', // 橙色
      'var(--chart-1)', // 循环使用
      'var(--chart-2)',
      'var(--chart-3)',
      'var(--chart-4)',
      'var(--chart-5)'
    ];

    return baseColors;
  };

  const colors = generateColors();
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

  // 自定义 Legend - 横向排列
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
    <>
      {isLoading ? (
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
          </CardHeader>
          <CardContent>
            <Skeleton className="h-72 w-full" />
          </CardContent>
        </Card>
      ) : (
        <Card className="@container/card overflow-hidden h-auto">
          <CardHeader>
            <CardTitle className="text-lg font-semibold">{t('dashboard_index.7days_model_usage_pie')}</CardTitle>
          </CardHeader>
          <CardContent>
            {data.length > 0 ? (
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
                      <p className="text-sm text-muted-foreground">{t('dashboard_index.total')}</p>
                      <p className="text-2xl font-bold">{total.toLocaleString()}</p>
                    </div>
                  </div>
                </div>
                <Legend content={<CustomLegend />} />
              </div>
            ) : (
              <div className="h-80 flex items-center justify-center rounded-lg bg-muted/50">
                <p className="text-lg font-medium text-muted-foreground">{t('dashboard_index.no_data_available')}</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </>
  );
};

ModelUsagePieChart.propTypes = {
  isLoading: PropTypes.bool,
  data: PropTypes.array
};

export default ModelUsagePieChart;
