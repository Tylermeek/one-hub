import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Box,
    Button,
    Card,
    CardContent,
    Grid,
    Typography,
    Chip,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
} from '@mui/material';
import { Icon } from '@iconify/react';
import { useTheme } from '@mui/material/styles';
import MainCard from 'ui-component/cards/MainCard';
import { gridSpacing } from 'store/constant';
import { API } from 'utils/api';
import { enqueueSnackbar } from 'notistack';
import { getRoleText, getRoleColor } from 'utils/teamPermissions';
import { renderQuota } from 'utils/common';

const Team = () => {
    const theme = useTheme();
    const navigate = useNavigate();
    const [teams, setTeams] = useState([]);
    const [loading, setLoading] = useState(true);
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [newTeam, setNewTeam] = useState({ name: '' });
    const [userTeamRoles, setUserTeamRoles] = useState({}); // 存储用户在团队中的角色

    // 获取用户团队列表
    const fetchTeams = async () => {
        try {
            setLoading(true);
            const response = await API.get('/api/team/list');
            if (response.data.success) {
                // 确保 data 是数组格式
                const teamsData = response.data.data;
                let teamsList = [];
                if (Array.isArray(teamsData)) {
                    teamsList = teamsData;
                } else if (teamsData && Array.isArray(teamsData.data)) {
                    teamsList = teamsData.data;
                }

                setTeams(teamsList);

                // 直接从团队数据中提取角色信息
                const rolesMap = {};
                teamsList.forEach(team => {
                    rolesMap[team.id] = {
                        role: team.current_user_role || 2,
                        isOwner: team.is_owner || false
                    };
                });
                setUserTeamRoles(rolesMap);
            } else {
                setTeams([]);
            }
        } catch (error) {
            console.error('获取团队列表失败:', error);
            enqueueSnackbar('获取团队列表失败', { variant: 'error' });
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTeams();
    }, []);

    // 创建团队
    const handleCreateTeam = async () => {
        if (!newTeam.name.trim()) {
            enqueueSnackbar('请输入团队名称', { variant: 'error' });
            return;
        }

        try {
            const response = await API.post('/api/team/', {
                name: newTeam.name.trim()
            });

            if (response.data.success) {
                enqueueSnackbar('团队创建成功', { variant: 'success' });
                setCreateDialogOpen(false);
                setNewTeam({ name: '' });
                fetchTeams();
            } else {
                enqueueSnackbar(response.data.message || '创建团队失败', { variant: 'error' });
            }
        } catch (error) {
            console.error('创建团队失败:', error);
            enqueueSnackbar('创建团队失败', { variant: 'error' });
        }
    };

    // 格式化额度显示
    const formatQuota = (quota, unlimited) => {
        if (unlimited) return '无限制';
        return renderQuota(quota || 0, 2);
    };

    // 获取状态颜色
    const getStatusColor = (status) => {
        return status === 1 ? 'success' : 'error';
    };

    // 获取状态文本
    const getStatusText = (status) => {
        return status === 1 ? '正常' : '禁用';
    };

    return (
        <MainCard>
            <Grid container spacing={gridSpacing}>
                <Grid item xs={12}>
                    <Grid container alignItems="center" justifyContent="space-between">
                        <Grid item>
                            <Typography variant="h3">我的团队</Typography>
                        </Grid>
                        <Grid item>
                            <Button
                                variant="contained"
                                startIcon={<Icon icon="solar:add-circle-bold-duotone" />}
                                onClick={() => setCreateDialogOpen(true)}
                            >
                                创建团队
                            </Button>
                        </Grid>
                    </Grid>
                </Grid>

                <Grid item xs={12}>
                    {loading ? (
                        <Box display="flex" justifyContent="center" p={3}>
                            <Typography>加载中...</Typography>
                        </Box>
                    ) : teams.length === 0 ? (
                        <Card>
                            <CardContent>
                                <Box textAlign="center" py={4}>
                                    <Icon icon="solar:users-group-rounded-bold-duotone" width={64} color={theme.palette.grey[400]} />
                                    <Typography variant="h6" color="textSecondary" mt={2}>
                                        您还没有创建任何团队
                                    </Typography>
                                    <Typography variant="body2" color="textSecondary" mb={3}>
                                        创建团队来管理成员和共享额度
                                    </Typography>
                                    <Button
                                        variant="contained"
                                        startIcon={<Icon icon="solar:add-circle-bold-duotone" />}
                                        onClick={() => setCreateDialogOpen(true)}
                                    >
                                        创建第一个团队
                                    </Button>
                                </Box>
                            </CardContent>
                        </Card>
                    ) : (
                        <Grid container spacing={gridSpacing}>
                            {Array.isArray(teams) && teams.map((team) => (
                                <Grid item xs={12} md={6} lg={4} key={team.id}>
                                    <Card
                                        sx={{
                                            cursor: 'pointer',
                                            transition: 'all 0.3s ease',
                                            '&:hover': {
                                                transform: 'translateY(-4px)',
                                                boxShadow: theme.shadows[8]
                                            }
                                        }}
                                        onClick={() => navigate(`/panel/team/${team.id}`)}
                                    >
                                        <CardContent>
                                            <Box display="flex" alignItems="center" justifyContent="space-between" mb={2}>
                                                <Typography variant="h6" noWrap>
                                                    {team.name}
                                                </Typography>
                                                <Box display="flex" gap={1}>
                                                    <Chip
                                                        label={getStatusText(team.status)}
                                                        color={getStatusColor(team.status)}
                                                        size="small"
                                                    />
                                                    {userTeamRoles[team.id] && (
                                                        <Chip
                                                            label={getRoleText(userTeamRoles[team.id].role, userTeamRoles[team.id].isOwner)}
                                                            color={getRoleColor(userTeamRoles[team.id].role, userTeamRoles[team.id].isOwner)}
                                                            size="small"
                                                        />
                                                    )}
                                                </Box>
                                            </Box>

                                            <Box mb={2}>
                                                <Typography variant="body2" color="textSecondary" gutterBottom>
                                                    团队额度
                                                </Typography>
                                                <Typography variant="h5" color="primary">
                                                    {formatQuota(team.quota, team.unlimited_quota)}
                                                </Typography>
                                            </Box>


                                            <Box display="flex" alignItems="center" justifyContent="space-between">
                                                <Typography variant="body2" color="textSecondary">
                                                    创建时间: {new Date(team.created_time * 1000).toLocaleDateString()}
                                                </Typography>
                                                <Icon icon="solar:arrow-right-bold-duotone" />
                                            </Box>
                                        </CardContent>
                                    </Card>
                                </Grid>
                            ))}
                        </Grid>
                    )}
                </Grid>
            </Grid>

            {/* 创建团队对话框 */}
            <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="sm" fullWidth>
                <DialogTitle>创建团队</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="团队名称"
                        fullWidth
                        variant="outlined"
                        value={newTeam.name}
                        onChange={(e) => setNewTeam({ ...newTeam, name: e.target.value })}
                        placeholder="请输入团队名称"
                    />
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setCreateDialogOpen(false)}>取消</Button>
                    <Button onClick={handleCreateTeam} variant="contained">
                        创建
                    </Button>
                </DialogActions>
            </Dialog>
        </MainCard>
    );
};

export default Team;
