import React from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useLocation } from 'react-router-dom';
import { useEffect } from 'react';
import { enqueueSnackbar } from 'notistack';

/**
 * 上下文路由守卫组件
 * 检查当前路由是否适用于当前上下文
 */
const ContextGuard = ({ children, requiredContext = null, fallbackPath = '/panel' }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { currentContext } = useSelector((state) => state.context);

  useEffect(() => {
    // 如果指定了必需的上下文类型
    if (requiredContext) {
      // 检查当前上下文是否匹配
      if (currentContext.type !== requiredContext) {
        // 显示提示信息
        const contextName = requiredContext === 'team' ? '团队' : '个人';
        enqueueSnackbar(`此页面仅适用于${contextName}上下文`, { variant: 'warning' });
        
        // 跳转到回退路径
        navigate(fallbackPath, { replace: true });
        return;
      }
    }

    // 特殊路由检查
    const pathname = location.pathname;
    
    // 团队详情页面需要团队上下文
    if (pathname.match(/^\/panel\/team\/\d+/) && currentContext.type !== 'team') {
      enqueueSnackbar('团队详情页面需要切换到团队上下文', { variant: 'warning' });
      navigate('/panel/team', { replace: true });
      return;
    }

    // 团队管理相关页面需要团队上下文
    if (pathname.includes('/team/') && !pathname.endsWith('/team') && currentContext.type !== 'team') {
      enqueueSnackbar('团队管理页面需要切换到团队上下文', { variant: 'warning' });
      navigate('/panel/team', { replace: true });
      return;
    }

  }, [currentContext, location.pathname, navigate, requiredContext, fallbackPath]);

  return <>{children}</>;
};

export default ContextGuard;
