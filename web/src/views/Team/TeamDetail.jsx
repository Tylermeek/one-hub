import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, UserPlus, Wallet, Settings as SettingsIcon, BarChart3 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { API } from 'utils/api';
import { showError, showSuccess, renderQuota } from 'utils/common';
import { canManageMembers, canManageQuota, getRoleText, getRoleBadgeVariant, ROLE } from 'utils/teamPermissions';

// 导入组件
import MetricCard from '../Dashboard/component/MetricCard';
import MemberTable from './components/MemberTable';
import InviteMemberDialog from './components/InviteMemberDialog';
import AllocateQuotaDialog from './components/AllocateQuotaDialog';
import TeamConsumptionChart from './components/TeamConsumptionChart';
import TeamModelUsageChart from './components/TeamModelUsageChart';
import MemberRankingChart from './components/MemberRankingChart';

const TeamDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  // 状态管理
  const [team, setTeam] = useState(null);
  const [members, setMembers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState(ROLE.MEMBER);

  // 对话框状态
  const [inviteDialogOpen, setInviteDialogOpen] = useState(false);
  const [allocateDialogOpen, setAllocateDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] = useState(null);

  // 统计数据
  const [statsData, setStatsData] = useState({
    trendData: [],
    modelData: [],
    memberRanking: []
  });

  // 获取团队详情
  const fetchTeamDetail = useCallback(async () => {
    try {
      setLoading(true);
      const response = await API.get(`/api/team/${id}`);
      if (response.data.success) {
        const teamData = response.data.data;
        setTeam(teamData);
        setUserRole(teamData.current_user_role ?? ROLE.MEMBER);
      } else {
        showError(response.data.message || '获取团队详情失败');
      }
    } catch (error) {
      console.error('获取团队详情失败:', error);
      showError('获取团队详情失败');
    } finally {
      setLoading(false);
    }
  }, [id]);

  // 获取团队成员
  const fetchMembers = useCallback(async () => {
    try {
      const response = await API.get(`/api/team/${id}/members`);
      if (response.data.success) {
        const membersData = response.data.data;
        let membersList = [];
        if (Array.isArray(membersData)) {
          membersList = membersData;
        } else if (membersData && Array.isArray(membersData.data)) {
          membersList = membersData.data;
        }
        setMembers(membersList);
      } else {
        showError(response.data.message || '获取成员列表失败');
      }
    } catch (error) {
      console.error('获取成员列表失败:', error);
      showError('获取成员列表失败');
    }
  }, [id]);

  useEffect(() => {
    if (id) {
      fetchTeamDetail();
      fetchMembers();
    }
  }, [id, fetchTeamDetail, fetchMembers]);

  // 设置成员额度
  const handleSetMemberQuota = async (member) => {
    const newQuota = prompt(`请输入 ${member.user?.username} 的最大额度（0表示无限制）:`, member.max_quota || 0);
    if (newQuota === null) return;

    try {
      const response = await API.put(`/api/team/${id}/member/${member.user_id}/quota`, {
        max_quota: parseInt(newQuota) || 0
      });
      if (response.data.success) {
        showSuccess('成员额度设置成功');
        fetchMembers();
      } else {
        showError(response.data.message || '设置失败');
      }
    } catch (error) {
      console.error('设置成员额度失败:', error);
      showError('设置成员额度失败');
    }
  };

  // 修改成员角色
  const handleChangeRole = async (userId, newRole) => {
    try {
      const response = await API.put(`/api/team/${id}/member/${userId}/role`, {
        role: newRole
      });
      if (response.data.success) {
        showSuccess('角色修改成功');
        fetchMembers();
      } else {
        showError(response.data.message || '修改失败');
      }
    } catch (error) {
      console.error('修改角色失败:', error);
      showError('修改角色失败');
    }
  };

  // 移除成员
  const handleRemoveMember = async (userId) => {
    try {
      const response = await API.delete(`/api/team/${id}/member/${userId}`);
      if (response.data.success) {
        showSuccess('成员移除成功');
        fetchMembers();
      } else {
        showError(response.data.message || '移除失败');
      }
    } catch (error) {
      console.error('移除成员失败:', error);
      showError('移除成员失败');
    }
  };

  // 权限检查
  const canManage = canManageMembers(userRole);
  const canAllocateQuota = canManageQuota(userRole);

  if (loading) {
    return (
      <div className="@container/main flex flex-1 flex-col gap-4 py-4 md:gap-6 md:py-6 px-4 lg:px-6">
        <Skeleton className="h-10 w-64" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardContent className="p-6">
                <Skeleton className="h-20 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (!team) {
    return (
      <div className="@container/main flex flex-1 flex-col items-center justify-center py-12">
        <p className="text-muted-foreground">团队不存在或您没有权限访问</p>
        <Button variant="outline" className="mt-4" onClick={() => navigate('/panel/team')}>
          返回列表
        </Button>
      </div>
    );
  }

  return (
    <div className="@container/main flex flex-1 flex-col gap-4 py-4 md:gap-6 md:py-6">
      {/* 顶部导航 */}
      <div className="flex flex-col gap-4 px-4 lg:px-6">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={() => navigate('/panel/team')}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              返回
            </Button>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-3xl font-bold tracking-tight">{team.name}</h1>
                <Badge variant={getRoleBadgeVariant(userRole)}>{getRoleText(userRole)}</Badge>
              </div>
              {team.description && <p className="text-muted-foreground mt-1">{team.description}</p>}
            </div>
          </div>
          <div className="flex gap-2">
            {canManage && (
              <>
                <Button variant="outline" size="sm" onClick={() => setInviteDialogOpen(true)}>
                  <UserPlus className="w-4 h-4 mr-2" />
                  邀请成员
                </Button>
                {canAllocateQuota && (
                  <Button variant="outline" size="sm" onClick={() => setAllocateDialogOpen(true)}>
                    <Wallet className="w-4 h-4 mr-2" />
                    分配额度
                  </Button>
                )}
              </>
            )}
          </div>
        </div>

        {/* 核心指标卡片 */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <MetricCard type="quota" title="团队额度" value={team.unlimited_quota ? '∞' : team.quota || 0} isLoading={false} />
          <MetricCard type="quota" title="已用额度" value={team.used_quota || 0} isLoading={false} />
          <MetricCard type="default" title="成员数量" value={members.length} isLoading={false} />
          <MetricCard type="request" title="7天请求" value={team.request_count_7d || 0} isLoading={false} />
        </div>
      </div>

      {/* Tab 标签页 */}
      <div className="px-4 lg:px-6">
        {userRole === ROLE.MEMBER ? (
          // 成员角色：只显示成员管理，无需 Tab 切换
          <div className="space-y-4">
            <MemberTable
              members={members}
              currentUserRole={userRole}
              onSetQuota={handleSetMemberQuota}
              onChangeRole={handleChangeRole}
              onRemove={handleRemoveMember}
            />
          </div>
        ) : (
          // Owner/Admin 角色：显示完整的 Tab 切换
          <Tabs defaultValue="members" className="space-y-4">
            <TabsList>
              <TabsTrigger value="members">成员管理</TabsTrigger>
              <TabsTrigger value="stats">使用统计</TabsTrigger>
            </TabsList>

            {/* Tab 1: 成员管理 */}
            <TabsContent value="members" className="space-y-4">
              <MemberTable
                members={members}
                currentUserRole={userRole}
                onSetQuota={handleSetMemberQuota}
                onChangeRole={handleChangeRole}
                onRemove={handleRemoveMember}
              />
            </TabsContent>

            {/* Tab 2: 使用统计 */}
            <TabsContent value="stats" className="space-y-4">
              <TeamConsumptionChart data={statsData.trendData} isLoading={false} />
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <TeamModelUsageChart data={statsData.modelData} isLoading={false} />
                <MemberRankingChart data={statsData.memberRanking} anonymize={userRole === ROLE.MEMBER} isLoading={false} />
              </div>
            </TabsContent>
          </Tabs>
        )}
      </div>

      {/* 对话框 */}
      <InviteMemberDialog
        open={inviteDialogOpen}
        onOpenChange={setInviteDialogOpen}
        teamId={parseInt(id)}
        onSuccess={() => {
          fetchMembers();
          fetchTeamDetail();
        }}
      />

      <AllocateQuotaDialog
        open={allocateDialogOpen}
        onOpenChange={setAllocateDialogOpen}
        teamId={parseInt(id)}
        teamData={team}
        onSuccess={fetchTeamDetail}
      />
    </div>
  );
};

export default TeamDetail;
