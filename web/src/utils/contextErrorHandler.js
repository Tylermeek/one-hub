import { enqueueSnackbar } from 'notistack';
import { useNavigate } from 'react-router-dom';

/**
 * 上下文错误处理工具
 * 统一处理上下文相关的错误和提示
 */
export class ContextErrorHandler {
    /**
     * 处理权限不足错误
     * @param {string} message - 错误消息
     * @param {string} fallbackPath - 回退路径
     */
    static handlePermissionError(message = '权限不足', fallbackPath = '/panel') {
        enqueueSnackbar(message, { variant: 'error' });

        // 延迟跳转，让用户看到错误提示
        setTimeout(() => {
            window.location.href = fallbackPath;
        }, 2000);
    }

    /**
     * 处理上下文不存在错误
     * @param {string} contextType - 上下文类型
     * @param {string} fallbackPath - 回退路径
     */
    static handleContextNotFoundError(contextType = 'team', fallbackPath = '/panel') {
        const contextName = contextType === 'team' ? '团队' : '个人';
        enqueueSnackbar(`${contextName}上下文不存在或已被删除`, { variant: 'error' });

        setTimeout(() => {
            window.location.href = fallbackPath;
        }, 2000);
    }

    /**
     * 处理团队不存在错误
     * @param {number} teamId - 团队ID
     * @param {string} fallbackPath - 回退路径
     */
    static handleTeamNotFoundError(teamId, fallbackPath = '/panel/team') {
        enqueueSnackbar(`团队 ${teamId} 不存在或已被删除`, { variant: 'error' });

        setTimeout(() => {
            window.location.href = fallbackPath;
        }, 2000);
    }

    /**
     * 处理上下文切换错误
     * @param {string} message - 错误消息
     * @param {Object} fallbackContext - 回退上下文
     */
    static handleContextSwitchError(message, fallbackContext = { type: 'user', id: 0, name: '个人空间' }) {
        enqueueSnackbar(`上下文切换失败: ${message}`, { variant: 'error' });

        // 触发回退上下文切换
        window.dispatchEvent(
            new CustomEvent('contextSwitchError', {
                detail: { fallbackContext }
            })
        );
    }

    /**
     * 处理API错误
     * @param {Error} error - 错误对象
     * @param {string} operation - 操作名称
     */
    static handleApiError(error, operation = '操作') {
        console.error(`${operation}失败:`, error);

        let message = `${operation}失败`;

        if (error.response) {
            const status = error.response.status;
            const data = error.response.data;

            switch (status) {
                case 403:
                    message = data?.message || '权限不足';
                    this.handlePermissionError(message);
                    return;
                case 404:
                    message = data?.message || '资源不存在';
                    break;
                case 500:
                    message = '服务器内部错误';
                    break;
                default:
                    message = data?.message || `${operation}失败`;
            }
        } else if (error.request) {
            message = '网络连接失败';
        }

        enqueueSnackbar(message, { variant: 'error' });
    }

    /**
     * 处理上下文验证错误
     * @param {string} expectedContext - 期望的上下文类型
     * @param {string} currentContext - 当前上下文类型
     */
    static handleContextValidationError(expectedContext, currentContext) {
        const expectedName = expectedContext === 'team' ? '团队' : '个人';
        const currentName = currentContext === 'team' ? '团队' : '个人';

        enqueueSnackbar(`此操作需要${expectedName}上下文，当前为${currentName}上下文`, {
            variant: 'warning'
        });
    }

    /**
     * 处理团队成员权限错误
     * @param {string} requiredRole - 需要的角色
     * @param {string} currentRole - 当前角色
     */
    static handleTeamPermissionError(requiredRole, currentRole) {
        const roleNames = {
            owner: '所有者',
            admin: '管理员',
            member: '成员'
        };

        enqueueSnackbar(`此操作需要${roleNames[requiredRole] || requiredRole}权限，当前为${roleNames[currentRole] || currentRole}`, {
            variant: 'error'
        });
    }
}

/**
 * React Hook 版本的错误处理器
 */
export const useContextErrorHandler = () => {
    const navigate = useNavigate();

    return {
        handlePermissionError: (message, fallbackPath) => {
            ContextErrorHandler.handlePermissionError(message, fallbackPath);
        },

        handleContextNotFoundError: (contextType, fallbackPath) => {
            ContextErrorHandler.handleContextNotFoundError(contextType, fallbackPath);
        },

        handleTeamNotFoundError: (teamId, fallbackPath) => {
            ContextErrorHandler.handleTeamNotFoundError(teamId, fallbackPath);
        },

        handleContextSwitchError: (message, fallbackContext) => {
            ContextErrorHandler.handleContextSwitchError(message, fallbackContext);
        },

        handleApiError: (error, operation) => {
            ContextErrorHandler.handleApiError(error, operation);
        },

        handleContextValidationError: (expectedContext, currentContext) => {
            ContextErrorHandler.handleContextValidationError(expectedContext, currentContext);
        },

        handleTeamPermissionError: (requiredRole, currentRole) => {
            ContextErrorHandler.handleTeamPermissionError(requiredRole, currentRole);
        }
    };
};

export default ContextErrorHandler;
