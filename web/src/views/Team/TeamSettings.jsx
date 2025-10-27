import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, Settings } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { canAccessSettings, isTeamOwner, ROLE } from 'utils/teamPermissions';
import TeamSettingsSidebar from './components/TeamSettingsSidebar';
import { BasicInfoForm, QuotaManagementForm, PermissionsForm, DangerZoneForm } from './components/TeamSettingsForm';

/**
 * 团队设置页面
 * 仅 Owner 和 Admin 可访问，不同角色显示不同内容
 */
function TeamSettings() {
    const { t } = useTranslation();
    const { id } = useParams();
    const navigate = useNavigate();

    const [isLoading, setIsLoading] = useState(true);
    const [team, setTeam] = useState(null);
    const [userRole, setUserRole] = useState(null);
    const [activeSection, setActiveSection] = useState('basic');

    /**
     * 加载团队数据
     */
    const loadTeamData = async () => {
        try {
            setIsLoading(true);
            const response = await API.get(`/api/team/${id}`);
            const { success, message, data } = response.data;

            if (success) {
                setTeam(data);
                const role = data.current_user_role ?? ROLE.MEMBER;
                setUserRole(role);

                // 权限检查
                if (!canAccessSettings(role)) {
                    showError(t('team_settings.no_permission'));
                    navigate(`/panel/team/${id}`);
                }
            } else {
                showError(message);
                navigate('/panel/team');
            }
        } catch (error) {
            console.error('Failed to load team:', error);
            showError(t('team_settings.load_failed'));
            navigate('/panel/team');
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => {
        loadTeamData();
    }, [id]);

    /**
     * 处理团队更新
     */
    const handleTeamUpdate = (updates) => {
        setTeam((prev) => ({ ...prev, ...updates }));
    };

    /**
     * 处理危险操作
     */
    const handleDangerAction = (action) => {
        if (action === 'transfer') {
            // 转让后重定向到详情页
            navigate(`/panel/team/${id}`);
        } else if (action === 'dissolve') {
            // 解散后重定向到列表页
            navigate('/panel/team');
        }
    };

    /**
     * 渲染内容区域
     */
    const renderContent = () => {
        if (!team) return null;

        const isOwnerRole = isTeamOwner(userRole);

        switch (activeSection) {
            case 'basic':
                return <BasicInfoForm team={team} onUpdate={handleTeamUpdate} isOwner={isOwnerRole} />;

            case 'quota':
                return isOwnerRole ? <QuotaManagementForm team={team} onUpdate={handleTeamUpdate} /> : null;

            case 'permissions':
                return isOwnerRole ? <PermissionsForm team={team} onUpdate={handleTeamUpdate} /> : null;

            case 'customization':
                return (
                    <Card className="p-6">
                        <div className="text-center text-muted-foreground py-12">{t('team_settings.customization_coming_soon')}</div>
                    </Card>
                );

            case 'danger':
                return isOwnerRole ? <DangerZoneForm team={team} onAction={handleDangerAction} /> : null;

            default:
                return null;
        }
    };

    /**
     * 渲染加载骨架
     */
    const renderSkeleton = () => (
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <div className="lg:col-span-1 space-y-2">
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
            </div>
            <div className="lg:col-span-3">
                <Skeleton className="h-96 w-full" />
            </div>
        </div>
    );

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
                        <Settings className="h-5 w-5 text-muted-foreground" />
                        <h1 className="text-2xl font-bold tracking-tight">{t('team_settings.page_title')}</h1>
                    </div>
                    {team && (
                        <p className="text-sm text-muted-foreground">
                            {t('team_settings.managing_team')}: <span className="font-medium">{team.name}</span>
                        </p>
                    )}
                </div>
            </div>

            {/* 内容区域 */}
            {isLoading ? (
                renderSkeleton()
            ) : team ? (
                <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                    {/* 侧边栏导航 */}
                    <div className="lg:col-span-1">
                        <div className="lg:sticky lg:top-6">
                            <TeamSettingsSidebar
                                activeSection={activeSection}
                                onSectionChange={setActiveSection}
                                isOwner={isTeamOwner(userRole)}
                            />
                        </div>
                    </div>

                    {/* 内容区域 */}
                    <div className="lg:col-span-3">{renderContent()}</div>
                </div>
            ) : (
                <Card className="p-12 text-center">
                    <p className="text-muted-foreground">{t('team_settings.team_not_found')}</p>
                </Card>
            )}
        </div>
    );
}

export default TeamSettings;
