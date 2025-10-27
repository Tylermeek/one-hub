import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, BarChart3, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { canViewDetailedStats } from 'utils/teamPermissions';
import TeamStatsOverview from './components/TeamStatsOverview';
import TimeRangeSelector from './components/TimeRangeSelector';
import TeamConsumptionChart from './components/TeamConsumptionChart';
import TeamAnalyticsCharts from './components/TeamAnalyticsCharts';
import DataExportTable from './components/DataExportTable';

/**
 * 团队统计分析页面
 * 根据团队配置和用户角色决定访问权限
 */
function TeamAnalytics() {
  const { t } = useTranslation();
  const { id } = useParams();
  const navigate = useNavigate();

  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [team, setTeam] = useState(null);
  const [userRole, setUserRole] = useState(null);
  const [timeRange, setTimeRange] = useState('7d');
  const [customRange, setCustomRange] = useState(null);

  // 统计数据
  const [statsOverview, setStatsOverview] = useState(null);
  const [consumptionData, setConsumptionData] = useState([]);
  const [analyticsData, setAnalyticsData] = useState(null);
  const [detailedData, setDetailedData] = useState([]);

  /**
   * 加载团队信息
   */
  const loadTeamInfo = async () => {
    try {
      const response = await API.get(`/api/team/${id}`);
      const { success, message, data } = response.data;

      if (success) {
        setTeam(data);
        setUserRole(data.user_role);

        // 权限检查
        if (!canViewDetailedStats(data.user_role, data.permissions)) {
          showError(t('team_analytics.no_permission'));
          navigate(`/panel/team/${id}`);
        }
      } else {
        showError(message);
        navigate('/panel/team');
      }
    } catch (error) {
      console.error('Failed to load team:', error);
      showError(t('team_analytics.load_failed'));
      navigate('/panel/team');
    }
  };

  /**
   * 加载统计数据
   */
  const loadAnalyticsData = async (showRefresh = false) => {
    try {
      if (showRefresh) {
        setIsRefreshing(true);
      } else {
        setIsLoading(true);
      }

      const params = {
        time_range: timeRange
      };

      if (customRange) {
        params.start_time = Math.floor(customRange.from.getTime() / 1000);
        params.end_time = Math.floor(customRange.to.getTime() / 1000);
      }

      const response = await API.get(`/api/team/${id}/analytics`, { params });
      const { success, message, data } = response.data;

      if (success) {
        setStatsOverview(data.overview);
        setConsumptionData(data.consumption_trend || []);
        setAnalyticsData({
          model_distribution: data.model_distribution || [],
          hourly_distribution: data.hourly_distribution || [],
          source_distribution: data.source_distribution || []
        });
        setDetailedData(data.detailed_logs || []);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('Failed to load analytics:', error);
      showError(t('team_analytics.load_data_failed'));
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  /**
   * 初始化加载
   */
  useEffect(() => {
    loadTeamInfo();
  }, [id]);

  /**
   * 时间范围变化时重新加载
   */
  useEffect(() => {
    if (team) {
      loadAnalyticsData();
    }
  }, [team, timeRange, customRange]);

  /**
   * 处理时间范围变化
   */
  const handleTimeRangeChange = (range, custom) => {
    setTimeRange(range);
    setCustomRange(custom);
  };

  /**
   * 刷新数据
   */
  const handleRefresh = () => {
    loadAnalyticsData(true);
  };

  return (
    <div className="container mx-auto px-4 py-8 max-w-7xl space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={() => navigate(`/panel/team/${id}`)}>
              <ChevronLeft className="h-4 w-4 mr-1" />
              {t('common.back')}
            </Button>
            <span className="text-muted-foreground">/</span>
            <BarChart3 className="h-5 w-5 text-muted-foreground" />
            <h1 className="text-2xl font-bold tracking-tight">{t('team_analytics.page_title')}</h1>
          </div>
          {team && (
            <p className="text-sm text-muted-foreground">
              {t('team_analytics.analyzing_team')}: <span className="font-medium">{team.name}</span>
            </p>
          )}
        </div>

        <div className="flex items-center gap-4">
          {/* 时间范围选择器 */}
          <TimeRangeSelector value={timeRange} onChange={handleTimeRangeChange} />

          {/* 刷新按钮 */}
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isRefreshing}>
            <RefreshCw className={cn('h-4 w-4 mr-2', isRefreshing && 'animate-spin')} />
            {t('common.refresh')}
          </Button>
        </div>
      </div>

      {/* 统计概览 */}
      <TeamStatsOverview stats={statsOverview} isLoading={isLoading} />

      {/* 消费趋势图 */}
      <TeamConsumptionChart data={consumptionData} isLoading={isLoading} timeRange={timeRange} />

      {/* 分析图表 */}
      <TeamAnalyticsCharts analyticsData={analyticsData} isLoading={isLoading} />

      {/* 详细数据表格 */}
      <DataExportTable data={detailedData} userRole={userRole} isLoading={isLoading} />
    </div>
  );
}

export default TeamAnalytics;
