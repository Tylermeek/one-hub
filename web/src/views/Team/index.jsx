import { useState, useEffect, useMemo } from 'react';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { API } from 'utils/api';
import { showError } from 'utils/common';
import { useTranslation } from 'react-i18next';

// 导入组件
import TeamCard from './components/TeamCard';
import EmptyTeamState from './components/EmptyTeamState';
import CreateTeamDialog from './components/CreateTeamDialog';
import TeamSearchFilter from './components/TeamSearchFilter';

const Team = () => {
    const { t } = useTranslation();
    const [teams, setTeams] = useState([]);
    const [loading, setLoading] = useState(true);
    const [createDialogOpen, setCreateDialogOpen] = useState(false);

    // 搜索和筛选状态
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [roleFilter, setRoleFilter] = useState('all');

    // 获取团队列表
    const fetchTeams = async () => {
        try {
            setLoading(true);
            const response = await API.get('/api/team/list');
            if (response.data.success) {
                const teamsData = response.data.data;
                let teamsList = [];
                if (Array.isArray(teamsData)) {
                    teamsList = teamsData;
                } else if (teamsData && Array.isArray(teamsData.data)) {
                    teamsList = teamsData.data;
                }
                setTeams(teamsList);
            } else {
                setTeams([]);
                showError(response.data.message || '获取团队列表失败');
            }
        } catch (error) {
            console.error('获取团队列表失败:', error);
            showError('获取团队列表失败');
            setTeams([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTeams();
    }, []);

    // 过滤团队列表
    const filteredTeams = useMemo(() => {
        return teams.filter((team) => {
            // 搜索过滤
            if (searchTerm && !team.name.toLowerCase().includes(searchTerm.toLowerCase())) {
                return false;
            }

            // 状态过滤
            if (statusFilter !== 'all') {
                if (statusFilter === 'active' && team.status !== 1) return false;
                if (statusFilter === 'disabled' && team.status === 1) return false;
            }

            // 角色过滤
            if (roleFilter !== 'all') {
                if (roleFilter === 'owner' && team.current_user_role !== 0) return false;
                if (roleFilter === 'admin' && team.current_user_role !== 1) return false;
                if (roleFilter === 'member' && team.current_user_role !== 2) return false;
            }

            return true;
        });
    }, [teams, searchTerm, statusFilter, roleFilter]);

    const handleCreateSuccess = () => {
        fetchTeams();
    };

    return (
        <div className="@container/main flex flex-1 flex-col gap-4 py-4 md:gap-6 md:py-6">
            {/* 头部 */}
            <div className="flex flex-col gap-4 px-4 lg:px-6">
                <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">我的团队</h1>
                        <p className="text-muted-foreground mt-1">管理团队成员和共享额度</p>
                    </div>
                    <Button onClick={() => setCreateDialogOpen(true)} size="lg">
                        <Plus className="w-4 h-4 mr-2" />
                        创建团队
                    </Button>
                </div>

                {/* 搜索和筛选 */}
                {!loading && teams.length > 0 && (
                    <TeamSearchFilter
                        searchTerm={searchTerm}
                        onSearchChange={setSearchTerm}
                        statusFilter={statusFilter}
                        onStatusFilterChange={setStatusFilter}
                        roleFilter={roleFilter}
                        onRoleFilterChange={setRoleFilter}
                    />
                )}
            </div>

            {/* 团队列表 */}
            <div className="px-4 lg:px-6">
                {loading ? (
                    // 骨架屏加载状态
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {[...Array(6)].map((_, index) => (
                            <Card key={index}>
                                <CardContent className="p-6">
                                    <Skeleton className="h-6 w-3/4 mb-4" />
                                    <Skeleton className="h-4 w-full mb-2" />
                                    <Skeleton className="h-4 w-2/3 mb-4" />
                                    <Skeleton className="h-2 w-full mb-2" />
                                    <Skeleton className="h-4 w-1/2" />
                                </CardContent>
                            </Card>
                        ))}
                    </div>
                ) : filteredTeams.length === 0 ? (
                    // 空状态
                    teams.length === 0 ? (
                        <EmptyTeamState onCreateTeam={() => setCreateDialogOpen(true)} />
                    ) : (
                        <Card className="border-dashed">
                            <CardContent className="flex flex-col items-center justify-center py-12">
                                <p className="text-muted-foreground">没有找到符合条件的团队</p>
                                <Button
                                    variant="outline"
                                    className="mt-4"
                                    onClick={() => {
                                        setSearchTerm('');
                                        setStatusFilter('all');
                                        setRoleFilter('all');
                                    }}
                                >
                                    清除筛选
                                </Button>
                            </CardContent>
                        </Card>
                    )
                ) : (
                    // 团队卡片网格
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {filteredTeams.map((team) => (
                            <TeamCard key={team.id} team={team} />
                        ))}
                    </div>
                )}
            </div>

            {/* 创建团队对话框 */}
            <CreateTeamDialog open={createDialogOpen} onOpenChange={setCreateDialogOpen} onSuccess={handleCreateSuccess} />
        </div>
    );
};

export default Team;
