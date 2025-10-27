import PropTypes from 'prop-types';
import React, { useState, useEffect } from 'react';
import { AlertTriangle, CreditCard, Calendar, TrendingUp } from 'lucide-react';
import { useTranslation } from 'react-i18next';

// shadcn components
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { API } from 'utils/api';
import { showError, renderNumber, renderQuota } from 'utils/common';
import { cn } from '@/lib/utils';

const QuotaAlertCard = ({ quotaData = null, isLoading = false, className = '' }) => {
    const { t } = useTranslation();
    const [alertData, setAlertData] = useState(null);
    const [alertLoading, setAlertLoading] = useState(false);

    // 获取额度预警数据
    const fetchAlertData = async () => {
        setAlertLoading(true);
        try {
            const res = await API.get('/api/user/dashboard/quota_alert');
            const { success, message, data } = res.data;
            if (success && data) {
                setAlertData(data);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Error fetching quota alert data:', error);
        } finally {
            setAlertLoading(false);
        }
    };

    useEffect(() => {
        fetchAlertData();
    }, []);

    // 获取预警级别
    const getAlertLevel = (usageRate) => {
        if (usageRate >= 90) return { level: 'critical', color: 'destructive' };
        if (usageRate >= 80) return { level: 'warning', color: 'secondary' };
        return { level: 'normal', color: 'default' };
    };

    const handleTopUp = () => {
        // 跳转到充值页面
        window.location.href = '/payment';
    };

    const isDataLoading = isLoading || alertLoading;
    const usageRate = alertData?.usage_rate || 0;
    const remainingDays = alertData?.remaining_days || 0;
    const dailyAvg = alertData?.daily_avg || 0;
    const alertLevel = getAlertLevel(usageRate);

    return (
        <Card className={cn('@container/card h-auto', className)}>
            <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                    <CreditCard className="w-4 h-4" />
                    {t('dashboard_index.quota_status')}
                </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
                {isDataLoading ? (
                    <div className="space-y-4">
                        <Skeleton className="h-4 w-full" />
                        <Skeleton className="h-20 w-full" />
                        <Skeleton className="h-8 w-full" />
                    </div>
                ) : (
                    <>
                        {/* 额度使用率 */}
                        <div className="space-y-2">
                            <div className="flex justify-between items-center">
                                <span className="text-sm text-muted-foreground">使用率</span>
                                <span className="text-sm font-medium">{usageRate.toFixed(1)}%</span>
                            </div>
                            <Progress value={usageRate} className="h-2" />
                            {quotaData && (
                                <div className="flex justify-between text-xs text-muted-foreground">
                                    <span>已用: {renderNumber(quotaData.used_quota)}</span>
                                    <span>总额: {quotaData.unlimited ? '∞' : renderNumber(quotaData.quota)}</span>
                                </div>
                            )}
                        </div>

                        {/* 预警信息 */}
                        {alertData?.alert && (
                            <div
                                className={cn(
                                    'p-3 rounded-lg border',
                                    alertLevel.level === 'critical' && 'bg-red-50 border-red-200 dark:bg-red-950 dark:border-red-800',
                                    alertLevel.level === 'warning' &&
                                        'bg-orange-50 border-orange-200 dark:bg-orange-950 dark:border-orange-800'
                                )}
                            >
                                <div className="flex items-center gap-2 mb-2">
                                    <AlertTriangle
                                        className={cn(
                                            'w-4 h-4',
                                            alertLevel.level === 'critical' && 'text-red-600 dark:text-red-400',
                                            alertLevel.level === 'warning' && 'text-orange-600 dark:text-orange-400'
                                        )}
                                    />
                                    <span
                                        className={cn(
                                            'text-sm font-medium',
                                            alertLevel.level === 'critical' && 'text-red-800 dark:text-red-200',
                                            alertLevel.level === 'warning' && 'text-orange-800 dark:text-orange-200'
                                        )}
                                    >
                                        {alertLevel.level === 'critical' ? '额度严重不足' : '额度使用率较高'}
                                    </span>
                                </div>
                                <p
                                    className={cn(
                                        'text-xs',
                                        alertLevel.level === 'critical' && 'text-red-700 dark:text-red-300',
                                        alertLevel.level === 'warning' && 'text-orange-700 dark:text-orange-300'
                                    )}
                                >
                                    建议及时充值以避免服务中断
                                </p>
                            </div>
                        )}

                        {/* 预测信息 */}
                        {remainingDays > 0 && (
                            <div className="space-y-2">
                                <div className="flex items-center gap-2 text-sm">
                                    <Calendar className="w-4 h-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">预计剩余天数:</span>
                                    <Badge variant="outline">{remainingDays.toFixed(1)} 天</Badge>
                                </div>
                                <div className="flex items-center gap-2 text-sm">
                                    <TrendingUp className="w-4 h-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">日均消费:</span>
                                    <span className="font-medium">{renderQuota(dailyAvg)}</span>
                                </div>
                            </div>
                        )}

                        {/* 快速充值按钮 */}
                        <Button onClick={handleTopUp} className="w-full" variant={alertData?.alert ? 'default' : 'outline'}>
                            <CreditCard className="w-4 h-4 mr-2" />
                            立即充值
                        </Button>
                    </>
                )}
            </CardContent>
        </Card>
    );
};

QuotaAlertCard.propTypes = {
    quotaData: PropTypes.shape({
        quota: PropTypes.number,
        used_quota: PropTypes.number,
        available: PropTypes.number,
        unlimited: PropTypes.bool
    }),
    isLoading: PropTypes.bool,
    className: PropTypes.string
};

export default QuotaAlertCard;
