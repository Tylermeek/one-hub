import React, { useState, useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import {
    Box,
    Button,
    Menu,
    MenuItem,
    Typography,
    Chip,
    Divider,
    ListItemIcon,
    ListItemText,
    CircularProgress
} from '@mui/material';
import { Icon } from '@iconify/react';
import { useTheme } from '@mui/material/styles';
import { API } from 'utils/api';
import { LOAD_USER_TEAMS } from 'store/actions';
import useContext from 'hooks/useContext';

const ContextSwitcher = () => {
    const theme = useTheme();
    const dispatch = useDispatch();
    const { currentContext, userTeams } = useSelector((state) => state.context);
    const { user } = useSelector((state) => state.account);

    // 使用 useContext hook 获取 switchContext 函数
    const { switchContext } = useContext();

    const [anchorEl, setAnchorEl] = useState(null);
    const [loading, setLoading] = useState(false);
    const open = Boolean(anchorEl);

    // 获取用户团队列表
    const fetchUserTeams = async () => {
        try {
            setLoading(true);
            const response = await API.get('/api/team/list');
            if (response.data.success) {
                const teamsData = response.data.data;
                const teams = Array.isArray(teamsData) ? teamsData : (teamsData?.data || []);
                dispatch({ type: LOAD_USER_TEAMS, payload: teams });
            }
        } catch (error) {
            console.error('获取团队列表失败:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        // 如果用户已登录且团队列表为空，则获取团队列表
        if (user && userTeams.length === 0) {
            fetchUserTeams();
        }
    }, [user]);

    const handleClick = (event) => {
        setAnchorEl(event.currentTarget);
        // 打开菜单时刷新团队列表
        if (user) {
            fetchUserTeams();
        }
    };

    const handleClose = () => {
        setAnchorEl(null);
    };

    const handleSwitchContext = (context) => {
        handleClose();
        // 调用 useContext hook 中的 switchContext 函数，它会调用后端 API
        switchContext(context);
    };

    // 获取上下文图标
    const getContextIcon = (type) => {
        return type === 'user' ? 'solar:home-bold-duotone' : 'solar:users-group-rounded-bold-duotone';
    };

    // 获取上下文颜色
    const getContextColor = (type) => {
        return type === 'user' ? 'primary' : 'secondary';
    };

    // 如果用户没有团队，不显示切换器
    if (!user || userTeams.length === 0) {
        return null;
    }

    return (
        <Box>
            <Button
                id="context-switcher-button"
                aria-controls={open ? 'context-switcher-menu' : undefined}
                aria-haspopup="true"
                aria-expanded={open ? 'true' : undefined}
                onClick={handleClick}
                startIcon={
                    <Icon
                        icon={getContextIcon(currentContext.type)}
                        width={20}
                        color={theme.palette[getContextColor(currentContext.type)].main}
                    />
                }
                endIcon={<Icon icon="solar:alt-arrow-down-bold-duotone" width={16} />}
                sx={{
                    color: 'text.primary',
                    textTransform: 'none',
                    fontWeight: 500,
                    '&:hover': {
                        backgroundColor: theme.palette.action.hover
                    }
                }}
            >
                <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start', mr: 1 }}>
                    <Typography variant="body2" sx={{ lineHeight: 1.2 }}>
                        {currentContext.name}
                    </Typography>
                    <Chip
                        label={currentContext.type === 'user' ? '个人空间' : '团队空间'}
                        size="small"
                        color={getContextColor(currentContext.type)}
                        variant="outlined"
                        sx={{
                            height: 16,
                            fontSize: '0.65rem',
                            '& .MuiChip-label': { px: 0.5 }
                        }}
                    />
                </Box>
            </Button>

            <Menu
                id="context-switcher-menu"
                anchorEl={anchorEl}
                open={open}
                onClose={handleClose}
                MenuListProps={{
                    'aria-labelledby': 'context-switcher-button',
                }}
                PaperProps={{
                    sx: {
                        minWidth: 200,
                        maxHeight: 400,
                        '& .MuiMenuItem-root': {
                            px: 2,
                            py: 1
                        }
                    }
                }}
            >
                {/* 个人空间 */}
                <MenuItem
                    onClick={() => handleSwitchContext({
                        type: 'user',
                        id: user.id,
                        name: '个人空间'
                    })}
                    selected={currentContext.type === 'user'}
                >
                    <ListItemIcon>
                        <Icon
                            icon="solar:home-bold-duotone"
                            width={20}
                            color={theme.palette.primary.main}
                        />
                    </ListItemIcon>
                    <ListItemText
                        primary="个人空间"
                        secondary="个人资源和设置"
                    />
                </MenuItem>

                <Divider />

                {/* 团队列表 */}
                {loading ? (
                    <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                        <CircularProgress size={20} />
                    </Box>
                ) : (
                    userTeams.map((team) => (
                        <MenuItem
                            key={team.id}
                            onClick={() => handleSwitchContext({
                                type: 'team',
                                id: team.id,
                                name: team.name
                            })}
                            selected={currentContext.type === 'team' && currentContext.id === team.id}
                        >
                            <ListItemIcon>
                                <Icon
                                    icon="solar:users-group-rounded-bold-duotone"
                                    width={20}
                                    color={theme.palette.secondary.main}
                                />
                            </ListItemIcon>
                            <ListItemText
                                primary={team.name}
                                secondary="团队空间"
                            />
                        </MenuItem>
                    ))
                )}

            </Menu>
        </Box>
    );
};

export default ContextSwitcher;
