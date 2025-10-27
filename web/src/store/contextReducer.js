// Context Reducer for managing team context switching
export const initialState = {
    currentContext: {
        type: 'user', // 'user' | 'team'
        id: null,
        name: '个人空间'
    },
    userTeams: [] // 用户所属的所有团队
};

const contextReducer = (state = initialState, action) => {
    switch (action.type) {
        case '@context/SWITCH_CONTEXT':
            // 保存到 localStorage 实现记忆功能
            localStorage.setItem('current_context', JSON.stringify(action.payload));
            return {
                ...state,
                currentContext: action.payload
            };
        case '@context/LOAD_USER_TEAMS':
            return {
                ...state,
                userTeams: action.payload
            };
        case '@context/RESTORE_CONTEXT':
            // 从 localStorage 恢复上下文
            const saved = localStorage.getItem('current_context');
            if (saved) {
                try {
                    const parsedContext = JSON.parse(saved);
                    return {
                        ...state,
                        currentContext: parsedContext
                    };
                } catch (error) {
                    console.error('Failed to parse saved context:', error);
                    localStorage.removeItem('current_context');
                }
            }
            return state;
        case '@context/CLEAR_CONTEXT':
            localStorage.removeItem('current_context');
            return {
                ...state,
                currentContext: initialState.currentContext
            };
        default:
            return state;
    }
};

export default contextReducer;
