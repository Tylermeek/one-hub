import { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useSelector } from 'react-redux';

// 导入新组件
import MetricCard from './component/MetricCard';
import WeeklyTrendChart from './component/WeeklyTrendChart';
import QuotaAlertCard from './component/QuotaAlertCard';
import TopModelsChart from './component/TopModelsChart';
import QuickActionsCard from './component/QuickActionsCard';
import ModelUsagePieChart from './component/ModelUsagePieChart';
import SupportModels from './component/SupportModels';
import QuickStartCard from './component/QuickStartCard';
import InviteCard from './component/InviteCard';
import StatusPanel from './component/StatusPanel';
import RecentLogsTable from './component/RecentLogsTable';

// 导入工具函数
import { API } from 'utils/api';
import { showError, calculateQuota } from 'utils/common';

const Dashboard = () => {
    const [isLoading, setLoading] = useState(true);
    const [summaryData, setSummaryData] = useState(null);
    const [dashboardData, setDashboardData] = useState(null);
    const [quotaData, setQuotaData] = useState(null);
    const [rateData, setRateData] = useState(null);

    const { t } = useTranslation();
    const siteInfo = useSelector((state) => state.siteInfo);

    // 获取 Dashboard 汇总数据
    const fetchSummaryData = useCallback(async () => {
        try {
            const res = await API.get('/api/user/dashboard/summary');
            const { success, message, data } = res.data;
            if (success && data) {
                setSummaryData(data);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Error fetching summary data:', error);
        }
    }, []);

    // 获取详细 Dashboard 数据
    const fetchDashboardData = useCallback(async () => {
        try {
            const res = await API.get('/api/user/dashboard');
            const { success, message, data } = res.data;
            if (success && data) {
                setDashboardData(data);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Error fetching dashboard data:', error);
        }
    }, []);

    // 获取额度数据
    const fetchQuotaData = useCallback(async () => {
        try {
            const res = await API.get('/api/user/context_quota');
            const { success, message, data } = res.data;
            if (success && data) {
                setQuotaData(data);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Error fetching quota data:', error);
        }
    }, []);

    // 获取速率数据
    const fetchRateData = useCallback(async () => {
        try {
            const res = await API.get('/api/user/dashboard/rate');
            const { success, message, data } = res.data;
            if (success && data) {
                setRateData(data);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Error fetching rate data:', error);
        }
    }, []);

    // 获取所有数据
    const fetchAllData = useCallback(async () => {
        setLoading(true);
        try {
            await Promise.all([fetchSummaryData(), fetchDashboardData(), fetchQuotaData(), fetchRateData()]);
        } catch (error) {
            console.error('Error fetching all data:', error);
        } finally {
            setLoading(false);
        }
    }, [fetchSummaryData, fetchDashboardData, fetchQuotaData, fetchRateData]);

    // 初始加载
    useEffect(() => {
        fetchAllData();
    }, [fetchAllData]);

    // 监听上下文切换事件，刷新数据
    useEffect(() => {
        const handleContextChange = () => {
            fetchAllData();
        };

        window.addEventListener('contextChanged', handleContextChange);
        return () => {
            window.removeEventListener('contextChanged', handleContextChange);
        };
    }, [fetchAllData]);
    // 处理模型使用数据
    const getModelUsageData = (data) => {
        if (!data || !Array.isArray(data)) {
            return [];
        }

        const modelUsage = {};
        data.forEach((item) => {
            if (!modelUsage[item.ModelName]) {
                modelUsage[item.ModelName] = 0;
            }
            modelUsage[item.ModelName] += item.RequestCount;
        });

        return Object.entries(modelUsage).map(([name, count]) => ({
            name,
            value: count
        }));
    };

    // 计算趋势数据（简化版，实际应该从历史数据计算）
    const calculateTrend = (current, previous) => {
        if (!previous || previous === 0) return 0;
        return Math.round(((current - previous) / previous) * 100);
    };

    // 生成图表数据
    const getChartData = (data, field) => {
        if (!data || !Array.isArray(data)) {
            return [];
        }

        const lastSevenDays = [];
        for (let i = 6; i >= 0; i--) {
            const date = new Date();
            date.setDate(date.getDate() - i);
            lastSevenDays.push(date.toISOString().split('T')[0]);
        }

        return lastSevenDays.map((date) => {
            const dayData = data.find((item) => item.Date === date);
            let value = 0;

            if (dayData) {
                switch (field) {
                    case 'requests':
                        value = dayData.RequestCount;
                        break;
                    case 'quota':
                        value = calculateQuota(dayData.Quota, 3);
                        break;
                    case 'tokens':
                        value = dayData.PromptTokens + dayData.CompletionTokens;
                        break;
                }
            }

            return { date: date, [field]: value };
        });
    };

    return (
        <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                {/* Section 1: 核心指标卡片 - 4列响应式网格 */}
                <div className="*:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-linear-to-t *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
                    <MetricCard
                        type="request"
                        title={t('dashboard_index.today_requests')}
                        value={summaryData?.today?.requests || 0}
                        trend={calculateTrend(summaryData?.today?.requests, summaryData?.week?.requests / 7)}
                        chartData={getChartData(dashboardData, 'requests')}
                        isLoading={isLoading}
                    />
                    <MetricCard
                        type="quota"
                        title={t('dashboard_index.today_consumption')}
                        value={summaryData?.today?.quota || 0}
                        trend={calculateTrend(summaryData?.today?.quota, summaryData?.week?.quota / 7)}
                        chartData={getChartData(dashboardData, 'quota')}
                        isLoading={isLoading}
                    />
                    <MetricCard
                        type="token"
                        title={t('dashboard_index.today_tokens')}
                        value={summaryData?.today?.tokens || 0}
                        trend={calculateTrend(summaryData?.today?.tokens, summaryData?.week?.tokens / 7)}
                        chartData={getChartData(dashboardData, 'tokens')}
                        isLoading={isLoading}
                    />
                    <MetricCard
                        type="rpm"
                        title={t('dashboard_index.RPM')}
                        value={rateData?.rpm || 0}
                        trend={0} // RPM 趋势需要从历史数据计算
                        maxValue={rateData?.maxRPM || 100} // 使用 maxRPM 作为进度条最大值
                        isLoading={isLoading}
                    />
                </div>

                {/* Section 2: 趋势图表 - 全宽 */}
                <div className="px-4 lg:px-6">
                    <WeeklyTrendChart data={summaryData?.daily_trend || []} isLoading={isLoading} />
                </div>

                {/* Section 3: API 日志表格 - 全宽 */}
                <div className="px-4 lg:px-6">
                    <RecentLogsTable />
                </div>

                {/* Section 4: 其他辅助组件 - 2列网格 */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 px-4 lg:px-6">
                    <QuotaAlertCard quotaData={quotaData} isLoading={isLoading} />
                    <QuickActionsCard />
                </div>

                {/* Section 5: 模型分析组件 - 2列网格 */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 px-4 lg:px-6">
                    <ModelUsagePieChart data={getModelUsageData(dashboardData)} isLoading={isLoading} />
                    <TopModelsChart data={summaryData?.top_models || []} isLoading={isLoading} />
                </div>

                {/* Section 6: 模型支持列表 - 全宽 */}
                <div className="px-4 lg:px-6">
                    <SupportModels />
                </div>

                {/* Section 7: 底部信息卡片 - 2列网格 */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 px-4 lg:px-6">
                    <QuickStartCard />
                    <InviteCard />
                </div>

                {/* Section 8: 状态监控面板 - 条件渲染 */}
                {siteInfo.UptimeEnabled && (
                    <div className="px-4 lg:px-6">
                        <StatusPanel />
                    </div>
                )}
            </div>
        </div>
    );
};

export default Dashboard;
