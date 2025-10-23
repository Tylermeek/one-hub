import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { RESTORE_CONTEXT, SWITCH_CONTEXT } from 'store/actions';
import { API } from 'utils/api';

const ContextProvider = ({ children }) => {
    const dispatch = useDispatch();
    const { user } = useSelector((state) => state.account);

    useEffect(() => {
        // 应用启动时恢复保存的上下文
        dispatch({ type: RESTORE_CONTEXT });
    }, [dispatch]);

    useEffect(() => {
        const initContext = async () => {
            if (user) {
                try {
                    const response = await API.get('/api/context/current');
                    if (response.data.success) {
                        const { type, id } = response.data.data;

                        // 获取空间详细信息
                        if (type === 'user') {
                            dispatch({
                                type: SWITCH_CONTEXT,
                                payload: { type: 'user', id, name: '个人空间' }
                            });
                        } else if (type === 'team') {
                            const teamRes = await API.get(`/api/team/${id}`);
                            if (teamRes.data.success) {
                                dispatch({
                                    type: SWITCH_CONTEXT,
                                    payload: {
                                        type: 'team',
                                        id,
                                        name: teamRes.data.data.name
                                    }
                                });
                            }
                        }
                    }
                } catch (error) {
                    console.error('获取当前空间失败:', error);
                }
            }
        };

        initContext();
    }, [user, dispatch]);

    return children;
};

export default ContextProvider;
