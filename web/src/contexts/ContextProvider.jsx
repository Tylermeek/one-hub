import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { RESTORE_CONTEXT } from 'store/actions';

const ContextProvider = ({ children }) => {
    const dispatch = useDispatch();

    useEffect(() => {
        // 应用启动时恢复保存的上下文
        dispatch({ type: RESTORE_CONTEXT });
    }, [dispatch]);

    return children;
};

export default ContextProvider;
