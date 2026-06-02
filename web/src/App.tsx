import { useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import { apiGet, removeToken } from './api/client';
import { clearUser, getUser } from './components/Navbar';
import Navbar from './components/Navbar';

// App 作为全局布局组件，包含导航栏和页面内容
export default function App() {
  // 启动时验证当前用户是否仍然存在（防止 DB 删用户后前端仍显示登录态）
  useEffect(() => {
    const user = getUser();
    const token = localStorage.getItem('auth_token');
    if (!user || !token) return;

    apiGet<{ email: string }>('/auth/me')
      .catch(() => {
        // 用户已被删除或 token 失效 → 清除前端状态
        removeToken();
        clearUser();
        window.dispatchEvent(new Event('auth-change'));
      });
  }, []);

  return (
    <>
      <Navbar />
      <Outlet />
    </>
  );
}
