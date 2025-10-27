import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { useTranslation } from 'react-i18next';
import { UserPlus, UserMinus, DollarSign, Shield, Activity, Settings, Clock, Filter, ChevronDown, ChevronUp } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { canViewAllActivities } from 'utils/teamPermissions';

/**
 * 活动类型图标映射
 */
const ACTIVITY_ICONS = {
    member_join: UserPlus,
    member_leave: UserMinus,
    quota_change: DollarSign,
    role_change: Shield,
    api_usage: Activity,
    settings_change: Settings
};

/**
 * 活动类型颜色映射
 */
const ACTIVITY_COLORS = {
    member_join: 'text-green-500',
    member_leave: 'text-red-500',
    quota_change: 'text-blue-500',
    role_change: 'text-purple-500',
    api_usage: 'text-yellow-500',
    settings_change: 'text-gray-500'
};

/**
 * 活动类型徽章样式
 */
const ACTIVITY_BADGE_VARIANTS = {
    member_join: 'default',
    member_leave: 'destructive',
    quota_change: 'secondary',
    role_change: 'outline',
    api_usage: 'default',
    settings_change: 'secondary'
};

/**
 * 团队活动时间线组件
 * 显示团队的活动历史记录，支持权限控制
 */
function TeamActivityTimeline({ teamId, userRole, currentUserId, className }) {
    const { t } = useTranslation();
    const [isLoading, setIsLoading] = useState(true);
    const [activities, setActivities] = useState([]);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(20);
    const [hasMore, setHasMore] = useState(true);
    const [totalCount, setTotalCount] = useState(0);

    // 筛选状态
    const [filterType, setFilterType] = useState('all');
    const [expandedItems, setExpandedItems] = useState(new Set());

    // 检查是否可以查看所有活动
    const canViewAll = canViewAllActivities(userRole);

    /**
     * 加载活动日志
     */
    const loadActivities = async (pageNum = 1, append = false) => {
        try {
            setIsLoading(true);
            const params = {
                page: pageNum,
                page_size: pageSize
            };

            // 如果不是 Owner/Admin，只加载自己的活动
            if (!canViewAll) {
                params.user_id = currentUserId;
            }

            // 添加类型筛选
            if (filterType !== 'all') {
                params.type = filterType;
            }

            const response = await API.get(`/api/team/${teamId}/activities`, { params });
            const { success, message, data } = response.data;

            if (success) {
                const newActivities = data.activities || [];
                setActivities(append ? [...activities, ...newActivities] : newActivities);
                setTotalCount(data.total || 0);
                setHasMore(newActivities.length === pageSize);
            } else {
                showError(message);
            }
        } catch (error) {
            console.error('Failed to load activities:', error);
            showError(t('team_detail.load_activities_failed'));
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        loadActivities(1, false);
    }, [teamId, filterType]);

    /**
     * 加载更多
     */
    const handleLoadMore = () => {
        const nextPage = page + 1;
        setPage(nextPage);
        loadActivities(nextPage, true);
    };

    /**
     * 切换展开/收起
     */
    const toggleExpand = (activityId) => {
        const newExpanded = new Set(expandedItems);
        if (newExpanded.has(activityId)) {
            newExpanded.delete(activityId);
        } else {
            newExpanded.add(activityId);
        }
        setExpandedItems(newExpanded);
    };

    /**
     * 格式化时间
     */
    const formatTime = (timestamp) => {
        const date = new Date(timestamp * 1000);
        const now = new Date();
        const diff = now - date;

        // 1小时内
        if (diff < 3600000) {
            const minutes = Math.floor(diff / 60000);
            return minutes <= 1 ? t('common.just_now') : `${minutes} ${t('common.minutes_ago')}`;
        }

        // 24小时内
        if (diff < 86400000) {
            const hours = Math.floor(diff / 3600000);
            return `${hours} ${t('common.hours_ago')}`;
        }

        // 7天内
        if (diff < 604800000) {
            const days = Math.floor(diff / 86400000);
            return `${days} ${t('common.days_ago')}`;
        }

        // 其他显示完整日期
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
    };

    /**
     * 渲染活动项
     */
    const renderActivity = (activity) => {
        const Icon = ACTIVITY_ICONS[activity.type] || Activity;
        const colorClass = ACTIVITY_COLORS[activity.type] || 'text-gray-500';
        const badgeVariant = ACTIVITY_BADGE_VARIANTS[activity.type] || 'default';
        const isExpanded = expandedItems.has(activity.id);

        return (
            <div key={activity.id} className="relative pl-8 pb-8 last:pb-0 group">
                {/* 时间线 */}
                <div className="absolute left-0 top-0 bottom-0 w-px bg-border group-last:hidden" />

                {/* 图标 */}
                <div
                    className={cn(
                        'absolute left-0 top-0 -translate-x-1/2 p-2 rounded-full bg-background border-2',
                        'transition-all duration-200 group-hover:scale-110',
                        colorClass
                    )}
                >
                    <Icon className="h-4 w-4" />
                </div>

                {/* 内容 */}
                <div className="space-y-2">
                    <div className="flex items-start justify-between gap-4">
                        <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                                <Badge variant={badgeVariant} className="text-xs">
                                    {t(`team_activity.type.${activity.type}`)}
                                </Badge>
                                <span className="text-sm font-medium text-foreground truncate">
                                    {activity.user_name || t('common.unknown_user')}
                                </span>
                            </div>
                            <p className="text-sm text-muted-foreground mt-1">
                                {activity.description || t(`team_activity.description.${activity.type}`)}
                            </p>
                        </div>

                        <div className="flex items-center gap-2 flex-shrink-0">
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                                <Clock className="h-3 w-3" />
                                <span>{formatTime(activity.created_at)}</span>
                            </div>
                            {activity.details && (
                                <Button variant="ghost" size="sm" onClick={() => toggleExpand(activity.id)} className="h-6 w-6 p-0">
                                    {isExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                                </Button>
                            )}
                        </div>
                    </div>

                    {/* 展开的详细信息 */}
                    {isExpanded && activity.details && (
                        <div className="mt-2 p-3 rounded-md bg-muted/50 text-sm">
                            <pre className="text-xs text-muted-foreground whitespace-pre-wrap break-all">
                                {JSON.stringify(activity.details, null, 2)}
                            </pre>
                        </div>
                    )}
                </div>
            </div>
        );
    };

    /**
     * 渲染加载骨架
     */
    const renderSkeleton = () => (
        <div className="space-y-6">
            {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="relative pl-8">
                    <Skeleton className="absolute left-0 top-0 h-8 w-8 rounded-full" />
                    <div className="space-y-2">
                        <Skeleton className="h-5 w-32" />
                        <Skeleton className="h-4 w-full" />
                        <Skeleton className="h-3 w-24" />
                    </div>
                </div>
            ))}
        </div>
    );

    return (
        <Card className={cn('overflow-hidden', className)}>
            <CardHeader>
                <div className="flex items-center justify-between">
                    <CardTitle className="text-xl font-semibold flex items-center gap-2">
                        <Activity className="h-5 w-5" />
                        {t('team_detail.activity_log')}
                        {totalCount > 0 && (
                            <Badge variant="secondary" className="ml-2">
                                {totalCount}
                            </Badge>
                        )}
                    </CardTitle>

                    {/* 筛选器 */}
                    <div className="flex items-center gap-2">
                        <Filter className="h-4 w-4 text-muted-foreground" />
                        <Select value={filterType} onValueChange={setFilterType}>
                            <SelectTrigger className="w-36">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">{t('common.all')}</SelectItem>
                                <SelectItem value="member_join">{t('team_activity.type.member_join')}</SelectItem>
                                <SelectItem value="member_leave">{t('team_activity.type.member_leave')}</SelectItem>
                                <SelectItem value="quota_change">{t('team_activity.type.quota_change')}</SelectItem>
                                <SelectItem value="role_change">{t('team_activity.type.role_change')}</SelectItem>
                                <SelectItem value="api_usage">{t('team_activity.type.api_usage')}</SelectItem>
                                <SelectItem value="settings_change">{t('team_activity.type.settings_change')}</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                {/* 权限提示 */}
                {!canViewAll && <div className="mt-2 text-xs text-muted-foreground">{t('team_detail.activity_limited_view')}</div>}
            </CardHeader>

            <CardContent className="px-6 py-4">
                {isLoading && page === 1 ? (
                    renderSkeleton()
                ) : activities.length === 0 ? (
                    <div className="text-center py-12">
                        <Activity className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                        <p className="text-muted-foreground">{t('team_detail.no_activities')}</p>
                    </div>
                ) : (
                    <>
                        <div className="space-y-0">{activities.map(renderActivity)}</div>

                        {/* 加载更多按钮 */}
                        {hasMore && (
                            <div className="mt-6 text-center">
                                <Button variant="outline" onClick={handleLoadMore} disabled={isLoading}>
                                    {isLoading ? t('common.loading') : t('common.load_more')}
                                </Button>
                            </div>
                        )}
                    </>
                )}
            </CardContent>
        </Card>
    );
}

TeamActivityTimeline.propTypes = {
    teamId: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
    userRole: PropTypes.number.isRequired,
    currentUserId: PropTypes.number.isRequired,
    className: PropTypes.string
};

export default TeamActivityTimeline;
