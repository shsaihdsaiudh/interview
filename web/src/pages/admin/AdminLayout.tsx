import { useState, useEffect } from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { apiGet } from '../../api/client';
import { getUser } from '../../components/Navbar';

const NAV_ITEMS = [
  { path: '/admin', label: '仪表盘', icon: '📊' },
  { path: '/admin/users', label: '用户管理', icon: '👥' },
  { path: '/admin/cards', label: '名片管理', icon: '🃏' },
  { path: '/admin/appointments', label: '预约管理', icon: '📅' },
];

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const [authorized, setAuthorized] = useState(false);
  const [checking, setChecking] = useState(true);
  const user = getUser();

  useEffect(() => {
    if (!user) {
      navigate('/login?redirect=/admin');
      return;
    }

    apiGet('/admin/stats')
      .then(() => {
        setAuthorized(true);
        setChecking(false);
      })
      .catch(() => {
        setChecking(false);
        navigate('/');
      });
  }, []);

  if (checking) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-surface">
        <span className="text-text-muted" style={{ fontSize: 18 }}>验证权限中...</span>
      </div>
    );
  }

  if (!authorized) return null;

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen flex bg-surface">
      {/* ── 侧边栏 ── */}
      <aside
        className="flex-shrink-0 bg-card border-r border-border flex flex-col"
        style={{ width: 220 }}
      >
        {/* Logo */}
        <div className="px-5 py-5 border-b border-border">
          <Link
            to="/admin"
            className="no-underline text-text font-bold tracking-wider"
            style={{ fontSize: 18, fontFamily: 'var(--font-sans)' }}
          >
            <span className="text-brand-600">mock·io</span> 后台
          </Link>
        </div>

        {/* 导航 */}
        <nav className="flex-1 py-3">
          {NAV_ITEMS.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className="no-underline block px-5 py-3 mx-2 transition-colors"
              style={{
                fontSize: 16,
                color: isActive(item.path) ? 'var(--color-brand-600)' : 'var(--color-text-secondary)',
                background: isActive(item.path) ? 'rgba(224,184,104,0.08)' : 'transparent',
                clipPath: isActive(item.path) ? 'polygon(0 2px, 2px 2px, 2px 0, 100% 0, 100% calc(100% - 2px), calc(100% - 2px) calc(100% - 2px), calc(100% - 2px) 100%, 0 100%)' : undefined,
              }}
            >
              <span className="mr-2">{item.icon}</span>
              {item.label}
            </Link>
          ))}
        </nav>

        {/* 返回前台 */}
        <div className="px-5 py-4 border-t border-border">
          <Link
            to="/"
            className="no-underline text-text-muted hover:text-text-secondary transition-colors"
            style={{ fontSize: 15 }}
          >
            ← 返回前台
          </Link>
        </div>
      </aside>

      {/* ── 内容区 ── */}
      <main className="flex-1 min-w-0 overflow-auto">
        <div className="max-w-5xl mx-auto px-6 py-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
