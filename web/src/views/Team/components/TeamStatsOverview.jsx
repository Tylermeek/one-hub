import React from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { Activity, DollarSign, Users, Zap } from 'lucide-react';
import MetricCard from '../../Dashboard/component/MetricCard';
import { Skeleton } from '@/components/ui/skeleton';
import { renderNumber, renderQuota } from 'utils/common';

/**
 * 团队统计概览组件
 * 显示 4 个核心统计指标
 */
function TeamStatsOverview({ stats, isLoading, className }) {
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className={className}>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-32 w-full" />
          ))}
        </div>
      </div>
    );
  }

  const metrics = [
    {
      title: t('team_analytics.total_requests'),
      value: renderNumber(stats?.total_requests || 0),
      description: t('team_analytics.total_requests_desc'),
      icon: Activity,
      trend: stats?.requests_trend,
      trendLabel: t('common.vs_last_period')
    },
    {
      title: t('team_analytics.total_consumption'),
      value: renderQuota(stats?.total_consumption || 0),
      description: t('team_analytics.total_consumption_desc'),
      icon: DollarSign,
      trend: stats?.consumption_trend,
      trendLabel: t('common.vs_last_period')
    },
    {
      title: t('team_analytics.active_members'),
      value: renderNumber(stats?.active_members || 0),
      suffix: ` / ${stats?.total_members || 0}`,
      description: t('team_analytics.active_members_desc'),
      icon: Users
    },
    {
      title: t('team_analytics.avg_rpm'),
      value: renderNumber(stats?.avg_rpm || 0),
      description: t('team_analytics.avg_rpm_desc'),
      icon: Zap,
      trend: stats?.rpm_trend,
      trendLabel: t('common.vs_last_period')
    }
  ];

  return (
    <div className={className}>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {metrics.map((metric, index) => (
          <MetricCard key={index} {...metric} />
        ))}
      </div>
    </div>
  );
}

TeamStatsOverview.propTypes = {
  stats: PropTypes.shape({
    total_requests: PropTypes.number,
    total_consumption: PropTypes.number,
    active_members: PropTypes.number,
    total_members: PropTypes.number,
    avg_rpm: PropTypes.number,
    requests_trend: PropTypes.number,
    consumption_trend: PropTypes.number,
    rpm_trend: PropTypes.number
  }),
  isLoading: PropTypes.bool,
  className: PropTypes.string
};

export default TeamStatsOverview;
