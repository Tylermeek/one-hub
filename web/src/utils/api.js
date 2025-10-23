import { showError } from './common';
import axios from 'axios';
import { store } from '../store';
import { LOGIN } from 'store/actions';

export const API = axios.create({
  // ... 其他代码 ...

  baseURL: import.meta.env.VITE_APP_SERVER || '/'
});

// 请求拦截器：添加上下文信息到请求头
API.interceptors.request.use((config) => {
  const state = store.getState();
  const currentContext = state.context?.currentContext;
  
  if (currentContext) {
    config.headers['X-Context-Type'] = currentContext.type;
    config.headers['X-Context-Id'] = currentContext.id;
  }
  
  return config;
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('user');
      store.dispatch({ type: LOGIN, payload: null });
      // window.location.href = '/login';
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
