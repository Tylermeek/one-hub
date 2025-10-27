import PropTypes from 'prop-types';
import React, { useState, useEffect } from 'react';
import { useTheme } from '@mui/material/styles';
import { Activity, TrendingUp, TrendingDown, Minus } from 'lucide-react';

// shadcn components
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from 'react-i18next';
import { API } from 'utils/api';
import { showError, renderNumber } from 'utils/common';
import { cn } from '@/lib/utils';

// 获取趋势图标
const getTrendIcon = (trend) => {
  if (trend > 0) return TrendingUp;
  if (trend < 0) return TrendingDown;
  return Minus;
};

// 获取趋势颜色
const getTrendColor = (trend) => {
  if (trend > 0) return 'var(--chart-5)'; // 橙色 - 上升
  if (trend < 0) return 'var(--chart-2)'; // 青色 - 下降
  return 'var(--chart-1)'; // 珊瑚色 - 持平
};

// ==============================|| RPM - REQUEST RATE LIMIT CARD ||============================== //

const RPM = ({ isLoading = false }) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const [rateData, setRateData] = useState({
    rpm: 0,
    maxRPM: 600,
    tpm: 0,
    maxTPM: 0,
    usageRpmRate: 0,
    usageTpmRate: 0,
    trend: 0 // 新增趋势数据
  });
  const [localLoading, setLocalLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);

  const fetchRPMData = async () => {
    setLocalLoading(true);
    try {
      const res = await API.get('/api/user/dashboard/rate');
      const { success, message, data } = res.data;
      if (success && data) {
        // 模拟趋势计算（实际应该从历史数据计算）
        const trend = Math.random() > 0.5 ? Math.floor(Math.random() * 20) - 10 : 0;

        setRateData({
          rpm: data.rpm || 0,
          maxRPM: data.maxRPM || 600,
          tpm: data.tpm || 0,
          maxTPM: data.maxTPM || 0,
          usageRpmRate: data.usageRpmRate || 0,
          usageTpmRate: data.usageTpmRate || 0,
          trend: trend
        });
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Error fetching RPM data:', error);
    } finally {
      setLocalLoading(false);
      setInitialLoading(false);
    }
  };

  // 初始加载数据
  useEffect(() => {
    fetchRPMData();
  }, []);

  const TrendIcon = getTrendIcon(rateData.trend);
  const trendColor = getTrendColor(rateData.trend);

  return (
    <>
      {isLoading || initialLoading ? (
        <Card className={cn('h-36 border-l-4 border-l-emerald-500')}>
          <CardContent className="p-4">
            <Skeleton className="h-4 w-20 mb-2" />
            <Skeleton className="h-8 w-16 mb-4" />
            <Skeleton className="h-3 w-24 mb-4" />
            <Skeleton className="h-12 w-full" />
          </CardContent>
        </Card>
      ) : (
        <Card className={cn('h-36 flex flex-col border-l-4 border-l-emerald-500 transition-all duration-200 hover:shadow-md')}>
          <CardContent className="p-4 flex-1 flex flex-col">
            <div className="flex justify-between items-center mb-1">
              <div>
                <h3 className="text-2xl font-medium text-foreground">{renderNumber(rateData.rpm)}</h3>
              </div>
              {rateData.trend !== 0 && (
                <div className="flex items-center bg-muted rounded-xl py-1 px-2">
                  <TrendIcon className="w-4 h-4 mr-1" style={{ color: trendColor }} />
                  <span className="text-xs font-medium" style={{ color: trendColor }}>
                    {`${Math.abs(rateData.trend)}%`}
                  </span>
                </div>
              )}
            </div>

            <p className="text-xs text-muted-foreground mb-4">{t('dashboard_index.RPM')}</p>

            <div className="mt-auto h-12 w-full flex items-center justify-center">
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <Activity className="w-4 h-4 text-emerald-500" />
                  <span className="text-sm text-muted-foreground">{rateData.usageRpmRate}% 使用率</span>
                </div>
                <div className="text-xs text-muted-foreground">
                  {rateData.rpm}/{rateData.maxRPM}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </>
  );
};

RPM.propTypes = {
  isLoading: PropTypes.bool
};

export default RPM;
