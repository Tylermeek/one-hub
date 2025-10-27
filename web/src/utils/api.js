import { showError } from './common';
import axios from 'axios';
import { store } from '../store';
import { LOGIN } from 'store/actions';

export const API = axios.create({
    // ... 其他代码 ...

    baseURL: import.meta.env.VITE_APP_SERVER || '/'
});

// ❌ 删除这段代码（不再需要前端设置 Header）
// 后端现在从 Session 读取上下文，前端无法篡改

// ✅ 保留响应拦截器，增加权限错误处理
API.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('user');
            store.dispatch({ type: LOGIN, payload: null });
            // window.location.href = '/login';
        } else if (error.response?.status === 403) {
            // 处理权限不足错误
            if (error.response?.data?.message) {
                error.message = error.response.data.message;
            }
        }

        if (error.response?.data?.message) {
            error.message = error.response.data.message;
        }

        showError(error);
    }
);

export const LoginCheckAPI = axios.create({
    // ... 其他代码 ...

    baseURL: import.meta.env.VITE_APP_SERVER || '/'
});
