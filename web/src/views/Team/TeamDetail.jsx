import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Box,
    Button,
    Card,
    CardContent,
    Grid,
    Typography,
    Chip,
    IconButton,
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    TextField,
    Alert,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Paper,
    Avatar,
    Tooltip,
    Tabs,
    Tab,
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    Switch,
    FormControlLabel
} from '@mui/material';
import { Icon } from '@iconify/react';
import { useTheme } from '@mui/material/styles';
import MainCard from 'ui-component/cards/MainCard';
import { gridSpacing } from 'store/constant';
import { API } from 'utils/api';
import { enqueueSnackbar } from 'notistack';
import { hasManagePermission, getRoleText, getRoleColor } from 'utils/teamPermissions';
import { renderQuota } from 'utils/common';

const TeamDetail = () => {
    const theme = useTheme();
    const navigate = useNavigate();
    const { id } = useParams();
    const [team, setTeam] = useState(null);
    const [members, setMembers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [currentUserRole, setCurrentUserRole] = useState(2);
    const [isOwner, setIsOwner] = useState(false);
    const [tabValue, setTabValue] = useState(0);
    const [inviteDialogOpen, setInviteDialogOpen] = useState(false);
    const [allocateDialogOpen, setAllocateDialogOpen] = useState(false);
    const [searchUsers, setSearchUsers] = useState([]);
    const [searchKeyword, setSearchKeyword] = useState('');
    const [selectedUser, setSelectedUser] = useState('');
    const [allocateAmount, setAllocateAmount] = useState('');
    const [unlimitedQuota, setUnlimitedQuota] = useState(false);

    // 获取团队详情
    const fetchTeamDetail = async () => {
        try {
            setLoading(true);
            const response = await API.get(`/api/team/${id}`);
            if (response.data.success) {
                setTeam(response.data.data);
                // 设置当前用户角色信息
                setCurrentUserRole(response.data.data.current_user_role || 2);
                setIsOwner(response.data.data.is_owner || false);
            }
        } catch (error) {
            console.error('获取团队详情失败:', error);
            enqueueSnackbar('获取团队详情失败', { variant: 'error' });
        } finally {
            setLoading(false);
        }
    };

    // 获取团队成员列表
    const fetchMembers = async () => {
        try {
            const response = await API.get(`/api/team/${id}/members`);
            if (response.data.success) {
                // 确保 data 是数组格式
                const membersData = response.data.data;
                if (Array.isArray(membersData)) {
                    setMembers(membersData);
                } else if (membersData && Array.isArray(membersData.data)) {
                    setMembers(membersData.data);
                } else {
                    setMembers([]);
                }
                // 更新角色信息（如果接口返回了的话）
                if (response.data.current_user_role !== undefined) {
                    setCurrentUserRole(response.data.current_user_role);
                }
                if (response.data.is_owner !== undefined) {
                    setIsOwner(response.data.is_owner);
                }
            } else {
                setMembers([]);
            }
        } catch (error) {
            console.error('获取成员列表失败:', error);
            enqueueSnackbar('获取成员列表失败', { variant: 'error' });
        }
    };

    useEffect(() => {
        if (id) {
            fetchTeamDetail();
            fetchMembers();
        }
    }, [id]);

    // 搜索用户
    const handleSearchUsers = async () => {
        if (!searchKeyword.trim()) {
            setSearchUsers([]);
            return;
        }

        try {
            const response = await API.get(`/api/team/search_users?keyword=${encodeURIComponent(searchKeyword)}`);
            if (response.data.success) {
                // 确保 data 是数组格式
                const usersData = response.data.data;
                if (Array.isArray(usersData)) {
                    setSearchUsers(usersData);
                } else if (usersData && Array.isArray(usersData.data)) {
                    setSearchUsers(usersData.data);
                } else {
                    setSearchUsers([]);
                }
            } else {
                setSearchUsers([]);
            }
        } catch (error) {
            console.error('搜索用户失败:', error);
            enqueueSnackbar('搜索用户失败', { variant: 'error' });
        }
    };

    // 邀请成员
    const handleInviteMember = async () => {
        if (!selectedUser) {
            enqueueSnackbar('请选择要邀请的用户', { variant: 'error' });
            return;
        }

        try {
            const response = await API.post(`/api/team/${id}/invite`, {
                user_id: selectedUser
            });

            if (response.data.success) {
                enqueueSnackbar('邀请发送成功', { variant: 'success' });
                setInviteDialogOpen(false);
                setSelectedUser('');
                setSearchKeyword('');
                setSearchUsers([]);
                fetchMembers();
            } else {
                enqueueSnackbar(response.data.message || '邀请失败', { variant: 'error' });
            }
        } catch (error) {
            console.error('邀请成员失败:', error);
            enqueueSnackbar('邀请成员失败', { variant: 'error' });
        }
    };

    // 分配额度
    const handleAllocateQuota = async () => {
        if (!unlimitedQuota && (!allocateAmount || allocateAmount <= 0)) {
            enqueueSnackbar('请输入有效的金额', { variant: 'error' });
            return;
        }

        // 检查是否超过个人余额
        if (!unlimitedQuota && team.owner_quota && parseFloat(allocateAmount) > team.owner_quota / 1000000) {
            enqueueSnackbar('团队消费上限不能超过您的个人余额', { variant: 'error' });
            return;
        }

        try {
            const response = await API.post(`/api/team/${id}/allocate`, {
                amount: unlimitedQuota ? 0 : parseFloat(allocateAmount),
                unlimited: unlimitedQuota
            });

            if (response.data.success) {
                enqueueSnackbar(response.data.message || '消费上限设置成功', { variant: 'success' });
                setAllocateDialogOpen(false);
                setAllocateAmount('');
                setUnlimitedQuota(false);
                fetchTeamDetail();
            } else {
                enqueueSnackbar(response.data.message || '设置失败', { variant: 'error' });
            }
        } catch (error) {
            console.error('设置失败:', error);
            enqueueSnackbar('设置失败', { variant: 'error' });
        }
    };

    // 移除成员
    const handleRemoveMember = async (userId) => {
        if (!window.confirm('确定要移除此成员吗？')) {
            return;
        }

        try {
            const response = await API.delete(`/api/team/${id}/member/${userId}`);
            if (response.data.success) {
                enqueueSnackbar('成员移除成功', { variant: 'success' });
                fetchMembers();
            } else {
                enqueueSnackbar(response.data.message || '移除成员失败', { variant: 'error' });
            }
        } catch (error) {
            console.error('移除成员失败:', error);
            enqueueSnackbar('移除成员失败', { variant: 'error' });
        }
    };

    // 设置成员额度限制
    const handleSetMemberQuota = async (userId, maxQuota) => {
        try {
            const response = await API.put(`/api/team/${id}/member/${userId}/quota`, {
                max_quota: maxQuota
            });
            if (response.data.success) {
                enqueueSnackbar('成员额度设置成功', { variant: 'success' });
                fetchMembers();
            } else {
                enqueueSnackbar(response.data.message || '设置失败', { variant: 'error' });
            }
        } catch (error) {
            console.error('设置成员额度失败:', error);
            enqueueSnackbar('设置成员额度失败', { variant: 'error' });
        }
    };

    // 格式化额度显示
    const formatQuota = (quota, unlimited) => {
        if (unlimited) return '无限制';
        if (quota === undefined || quota === null) return '0';
        return renderQuota(quota, 2);
    };

    // 获取角色文本（使用工具函数）
    const getRoleDisplayText = (role, isOwner) => {
        return getRoleText(role, isOwner);
    };

    // 获取角色颜色（使用工具函数）
    const getRoleDisplayColor = (role, isOwner) => {
        return getRoleColor(role, isOwner);
    };

    if (loading) {
        return (
            <MainCard>
                <Box display="flex" justifyContent="center" p={3}>
                    <Typography>加载中...</Typography>
                </Box>
            </MainCard>
        );
    }

    if (!team) {
        return (
            <MainCard>
                <Alert severity="error">团队不存在或您没有权限访问</Alert>
            </MainCard>
        );
    }

    return (
        <MainCard>
            <Grid container spacing={gridSpacing}>
                {/* 团队基本信息 */}
                <Grid item xs={12}>
                    <Card>
                        <CardContent>
                            <Grid container alignItems="center" justifyContent="space-between" mb={2}>
                                <Grid item>
                                    <Box display="flex" alignItems="center" gap={2} mb={1}>
                                        <Typography variant="h4">{team.name}</Typography>
                                        <Chip
                                            label={getRoleText(currentUserRole, isOwner)}
                                            color={getRoleColor(currentUserRole, isOwner)}
                                            size="small"
                                        />
                                    </Box>
                                    <Typography variant="body2" color="textSecondary">
                                        创建时间: {new Date(team.created_time * 1000).toLocaleString()}
                                    </Typography>
                                </Grid>
                                <Grid item>
                                    <Button
                                        variant="outlined"
                                        startIcon={<Icon icon="solar:arrow-left-bold-duotone" />}
                                        onClick={() => navigate('/panel/team')}
                                    >
                                        返回列表
                                    </Button>
                                </Grid>
                            </Grid>

                            <Grid container spacing={3}>
                                <Grid item xs={12} md={4}>
                                    <Box textAlign="center">
                                        <Typography variant="body2" color="textSecondary" gutterBottom>
                                            团队额度
                                        </Typography>
                                        <Typography variant="h4" color="primary">
                                            {formatQuota(team.quota, team.unlimited_quota)}
                                        </Typography>
                                        {team.unlimited_quota && team.owner_balance !== undefined && (
                                            <Typography variant="body2" color="textSecondary">
                                                （实际可用最大额度: {renderQuota(team.owner_balance, 2)}）
                                            </Typography>
                                        )}
                                    </Box>
                                </Grid>
                                <Grid item xs={12} md={4}>
                                    <Box textAlign="center">
                                        <Typography variant="body2" color="textSecondary" gutterBottom>
                                            已使用额度
                                        </Typography>
                                        <Typography variant="h4">
                                            {renderQuota(team.used_quota || 0, 2)}
                                        </Typography>
                                    </Box>
                                </Grid>
                                <Grid item xs={12} md={4}>
                                    <Box textAlign="center">
                                        <Typography variant="body2" color="textSecondary" gutterBottom>
                                            成员数量
                                        </Typography>
                                        <Typography variant="h4">
                                            {members.length}
                                        </Typography>
                                    </Box>
                                </Grid>
                            </Grid>
                        </CardContent>
                    </Card>
                </Grid>

                {/* 操作按钮 */}
                <Grid item xs={12}>
                    <Box display="flex" gap={2}>
                        {hasManagePermission(team, null, currentUserRole, isOwner) && (
                            <>
                                <Button
                                    variant="contained"
                                    startIcon={<Icon icon="solar:user-plus-bold-duotone" />}
                                    onClick={() => setInviteDialogOpen(true)}
                                >
                                    邀请成员
                                </Button>
                                <Button
                                    variant="contained"
                                    startIcon={<Icon icon="solar:wallet-money-bold-duotone" />}
                                    onClick={() => setAllocateDialogOpen(true)}
                                >
                                    分配额度
                                </Button>
                            </>
                        )}
                    </Box>
                </Grid>

                {/* 成员列表 */}
                <Grid item xs={12}>
                    <Card>
                        <CardContent>
                            <Typography variant="h6" gutterBottom>
                                团队成员
                            </Typography>
                            <TableContainer>
                                <Table>
                                    <TableHead>
                                        <TableRow>
                                            <TableCell>用户</TableCell>
                                            <TableCell>角色</TableCell>
                                            <TableCell>最大额度</TableCell>
                                            <TableCell>额度使用情况</TableCell>
                                            <TableCell>加入时间</TableCell>
                                            {hasManagePermission(team, null, currentUserRole, isOwner) && (
                                                <TableCell>操作</TableCell>
                                            )}
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {Array.isArray(members) && members.map((member) => (
                                            <TableRow key={member.id}>
                                                <TableCell>
                                                    <Box display="flex" alignItems="center">
                                                        <Avatar sx={{ mr: 2 }}>
                                                            {member.user?.username?.charAt(0).toUpperCase() || 'U'}
                                                        </Avatar>
                                                        <Box>
                                                            <Typography variant="subtitle2">{member.user?.username || '未知用户'}</Typography>
                                                            <Typography variant="body2" color="textSecondary">
                                                                {member.user?.email || '无邮箱'}
                                                            </Typography>
                                                        </Box>
                                                    </Box>
                                                </TableCell>
                                                <TableCell>
                                                    <Chip
                                                        label={getRoleDisplayText(member.role, false)}
                                                        color={getRoleDisplayColor(member.role, false)}
                                                        size="small"
                                                    />
                                                </TableCell>
                                                <TableCell>
                                                    {member.max_quota === 0 ? '无限制' : renderQuota(member.max_quota || 0, 2)}
                                                </TableCell>
                                                <TableCell>
                                                    <Box>
                                                        <Typography variant="body2">
                                                            已用: {renderQuota(member.used_quota || 0, 2)}
                                                        </Typography>
                                                        <Typography variant="body2" color="primary">
                                                            可用: {member.unlimited_quota ? '无限制' : renderQuota(member.available_quota || 0, 2)}
                                                        </Typography>
                                                    </Box>
                                                </TableCell>
                                                <TableCell>
                                                    {new Date(member.joined_time * 1000).toLocaleDateString()}
                                                </TableCell>
                                                {hasManagePermission(team, null, currentUserRole, isOwner) && (
                                                    <TableCell>
                                                        <Box display="flex" gap={1}>
                                                            <Tooltip title="设置额度限制">
                                                                <IconButton
                                                                    size="small"
                                                                    onClick={() => {
                                                                        const newQuota = prompt('请输入最大额度（0表示无限制）:', member.max_quota);
                                                                        if (newQuota !== null) {
                                                                            handleSetMemberQuota(member.user_id, parseInt(newQuota) || 0);
                                                                        }
                                                                    }}
                                                                >
                                                                    <Icon icon="solar:settings-bold-duotone" />
                                                                </IconButton>
                                                            </Tooltip>
                                                            {member.role !== 1 && (
                                                                <Tooltip title="移除成员">
                                                                    <IconButton
                                                                        size="small"
                                                                        color="error"
                                                                        onClick={() => handleRemoveMember(member.user_id)}
                                                                    >
                                                                        <Icon icon="solar:user-minus-bold-duotone" />
                                                                    </IconButton>
                                                                </Tooltip>
                                                            )}
                                                        </Box>
                                                    </TableCell>
                                                )}
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        </CardContent>
                    </Card>
                </Grid>
            </Grid>

            {/* 邀请成员对话框 */}
            <Dialog open={inviteDialogOpen} onClose={() => setInviteDialogOpen(false)} maxWidth="sm" fullWidth>
                <DialogTitle>邀请成员</DialogTitle>
                <DialogContent>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="搜索用户"
                        fullWidth
                        variant="outlined"
                        value={searchKeyword}
                        onChange={(e) => setSearchKeyword(e.target.value)}
                        onKeyPress={(e) => e.key === 'Enter' && handleSearchUsers()}
                        placeholder="输入用户名或邮箱"
                    />
                    <Button
                        variant="outlined"
                        onClick={handleSearchUsers}
                        sx={{ mt: 1 }}
                        startIcon={<Icon icon="solar:magnifer-bold-duotone" />}
                    >
                        搜索
                    </Button>

                    {searchUsers.length > 0 && (
                        <Box mt={2}>
                            <Typography variant="subtitle2" gutterBottom>
                                搜索结果:
                            </Typography>
                            {Array.isArray(searchUsers) && searchUsers.map((user) => (
                                <Card
                                    key={user.id}
                                    sx={{
                                        mb: 1,
                                        cursor: 'pointer',
                                        border: selectedUser === user.id ? `2px solid ${theme.palette.primary.main}` : '1px solid transparent'
                                    }}
                                    onClick={() => setSelectedUser(user.id)}
                                >
                                    <CardContent sx={{ py: 1 }}>
                                        <Box display="flex" alignItems="center">
                                            <Avatar sx={{ mr: 2 }}>
                                                {user.username?.charAt(0).toUpperCase()}
                                            </Avatar>
                                            <Box>
                                                <Typography variant="subtitle2">{user.username}</Typography>
                                                <Typography variant="body2" color="textSecondary">
                                                    {user.email}
                                                </Typography>
                                            </Box>
                                        </Box>
                                    </CardContent>
                                </Card>
                            ))}
                        </Box>
                    )}
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setInviteDialogOpen(false)}>取消</Button>
                    <Button onClick={handleInviteMember} variant="contained" disabled={!selectedUser}>
                        邀请
                    </Button>
                </DialogActions>
            </Dialog>

            {/* 分配额度对话框 */}
            <Dialog open={allocateDialogOpen} onClose={() => setAllocateDialogOpen(false)} maxWidth="sm" fullWidth>
                <DialogTitle>分配额度</DialogTitle>
                <DialogContent>
                    <FormControlLabel
                        control={
                            <Switch
                                checked={unlimitedQuota}
                                onChange={(e) => setUnlimitedQuota(e.target.checked)}
                            />
                        }
                        label="无限制额度"
                    />

                    {!unlimitedQuota && (
                        <TextField
                            autoFocus
                            margin="dense"
                            label="消费上限（美元）"
                            fullWidth
                            variant="outlined"
                            type="number"
                            step="0.01"
                            value={allocateAmount}
                            onChange={(e) => setAllocateAmount(e.target.value)}
                            placeholder="请输入团队消费上限"
                            helperText={team.owner_quota ? `您的个人余额：${renderQuota(team.owner_quota, 2)}` : ''}
                        />
                    )}

                    <Alert severity="info" sx={{ mt: 2 }}>
                        设置团队消费上限，不会从您的个人余额扣除。无限额度时，上限为您的个人余额。
                    </Alert>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setAllocateDialogOpen(false)}>取消</Button>
                    <Button onClick={handleAllocateQuota} variant="contained">
                        分配
                    </Button>
                </DialogActions>
            </Dialog>
        </MainCard>
    );
};

export default TeamDetail;
