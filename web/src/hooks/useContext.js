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

  // 切换上下文
  const switchContext = useCallback(
    async (context) => {
      try {
        // 🔐 调用后端 API 切换空间
        const response = await API.post('/api/context/switch', {
          type: context.type,
          id: context.id
        });

        if (response.data.success) {
          // 更新 Redux 状态
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

          // 触发页面数据刷新 - 添加延迟确保状态更新完成
          setTimeout(() => {
            window.dispatchEvent(
              new CustomEvent('contextChanged', {
                detail: {
                  newContext: context,
                  oldContext: currentContext
                }
              })
            );
          }, 100);
        } else {
          enqueueSnackbar(response.data.message, { variant: 'error' });
        }
      } catch (error) {
        console.error('切换空间失败:', error);
        enqueueSnackbar('切换空间失败', { variant: 'error' });
      }
    },
    [dispatch, navigate, location.pathname, currentContext]
  );

  // 加载用户团队列表
  const loadUserTeams = useCallback(async () => {
    try {
      console.log('loadUserTeams: Starting to load teams...');
      const response = await API.get('/api/team/list');
      console.log('loadUserTeams: API response:', response.data);
      if (response.data.success) {
        const teamsData = response.data.data;
        console.log('loadUserTeams: Raw teamsData:', teamsData);

        let teams = [];
        if (Array.isArray(teamsData)) {
          teams = teamsData;
        } else if (teamsData && Array.isArray(teamsData.data)) {
          teams = teamsData.data;
        } else if (teamsData && typeof teamsData === 'object') {
          // 如果 teamsData 是对象，尝试提取数组
          teams = Object.values(teamsData).find((value) => Array.isArray(value)) || [];
        }

        // 参考旧版代码的处理方式：确保 teams 是数组
        if (!Array.isArray(teams)) {
          console.log('loadUserTeams: teams is not array, converting...', teams);
          teams = [];
        }

        console.log('loadUserTeams: Processed teams:', teams);
        dispatch({ type: LOAD_USER_TEAMS, payload: teams });
        return teams;
      }
      console.log('loadUserTeams: API returned success=false');
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
