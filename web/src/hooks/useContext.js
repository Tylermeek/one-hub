import { useSelector, useDispatch } from 'react-redux';
import { useCallback } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { API } from 'utils/api';
import { enqueueSnackbar } from 'notistack';
import { SWITCH_CONTEXT, LOAD_USER_TEAMS } from 'store/actions';

const useContext = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { currentContext, userTeams } = useSelector((state) => state.context);
  const { user } = useSelector((state) => state.account);

  // 切换上下文
  const switchContext = useCallback((context) => {
    dispatch({ type: SWITCH_CONTEXT, payload: context });
    enqueueSnackbar(`已切换到「${context.name}」`, { variant: 'success' });
    
    // 根据上下文类型和当前路径决定是否需要跳转
    const currentPath = location.pathname;
    
    // 如果从团队上下文切换到个人上下文
    if (context.type === 'user' && currentContext.type === 'team') {
      // 如果当前在团队相关页面，跳转到 Dashboard
      if (currentPath.includes('/team/') && !currentPath.endsWith('/team')) {
        navigate('/panel', { replace: true });
      }
    }
    
    // 如果从个人上下文切换到团队上下文
    if (context.type === 'team' && currentContext.type === 'user') {
      // 如果当前在个人相关页面，可以跳转到团队 Dashboard 或保持当前页面
      // 这里可以根据具体需求调整
    }
    
    // 触发页面数据刷新（通过 window.location.reload 或自定义事件）
    // 这里使用自定义事件，让各个页面组件监听并刷新数据
    window.dispatchEvent(new CustomEvent('contextChanged', { 
      detail: { 
        newContext: context, 
        oldContext: currentContext 
      } 
    }));
  }, [dispatch, navigate, location.pathname, currentContext]);

  // 加载用户团队列表
  const loadUserTeams = useCallback(async () => {
    try {
      const response = await API.get('/api/team/list');
      if (response.data.success) {
        const teamsData = response.data.data;
        const teams = Array.isArray(teamsData) ? teamsData : (teamsData?.data || []);
        dispatch({ type: LOAD_USER_TEAMS, payload: teams });
        return teams;
      }
      return [];
    } catch (error) {
      console.error('获取团队列表失败:', error);
      enqueueSnackbar('获取团队列表失败', { variant: 'error' });
      return [];
    }
  }, [dispatch]);

  // 获取当前上下文的额度信息
  const getContextQuota = useCallback(async () => {
    try {
      const response = await API.get('/api/user/context_quota');
      if (response.data.success) {
        return response.data.data;
      }
      return null;
    } catch (error) {
      console.error('获取上下文额度失败:', error);
      return null;
    }
  }, []);

  // 检查用户是否有团队
  const hasTeams = userTeams.length > 0;

  // 检查当前是否在团队上下文
  const isTeamContext = currentContext.type === 'team';

  // 检查当前是否在个人上下文
  const isUserContext = currentContext.type === 'user';

  return {
    currentContext,
    userTeams,
    hasTeams,
    isTeamContext,
    isUserContext,
    switchContext,
    loadUserTeams,
    getContextQuota
  };
};

export default useContext;
