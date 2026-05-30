import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { removeToken } from '../api/client';

interface UserInfo {
  email: string;
  nickname: string;
}

const USER_KEY = 'auth_user';

export function getUser(): UserInfo | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function setUser(user: UserInfo): void {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearUser(): void {
  localStorage.removeItem(USER_KEY);
}

const AUTH_CHANGE = 'auth-change';
export function notifyAuthChange(): void {
  window.dispatchEvent(new Event(AUTH_CHANGE));
}

function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUserState] = useState<UserInfo | null>(getUser());
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handler = () => setUserState(getUser());
    window.addEventListener(AUTH_CHANGE, handler);
    return () => window.removeEventListener(AUTH_CHANGE, handler);
  }, []);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 10);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  const handleLogout = () => {
    removeToken();
    clearUser();
    setUserState(null);
    navigate('/');
  };

  const isActive = (path: string) => location.pathname.startsWith(path);

  const navLink = (path: string, label: string) => (
    <Link
      to={path}
      className={`text-sm font-medium transition-colors no-underline px-3 py-1.5 rounded-md ${
        isActive(path)
          ? 'text-brand-700 bg-brand-50'
          : 'text-text-secondary hover:text-text'
      }`}
    >
      {label}
    </Link>
  );

  return (
    <nav
      className={`sticky top-0 z-50 transition ${
        scrolled
          ? 'glass border-b border-border'
          : 'bg-white border-b border-transparent'
      }`}
    >
      <div className="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 no-underline">
          <div className="w-7 h-7 rounded-lg bg-brand-600 flex items-center justify-center">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round">
              <circle cx="12" cy="12" r="10" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </div>
          <span className="text-sm font-bold text-text">面试互助平台</span>
        </Link>

        <div className="flex items-center gap-1">
          {user ? (
            <>
              {navLink('/find', '找人')}
              {navLink('/appointments', '预约')}
              {navLink('/settings/availability', '设置')}

              <span className="w-px h-4 bg-border mx-2" />

              <span className="text-sm font-medium text-text flex items-center gap-1.5">
                <span className="w-6 h-6 rounded-full bg-brand-600 flex items-center justify-center text-white text-xs font-bold">
                  {user.nickname.charAt(0)}
                </span>
                <span className="hidden sm:inline">{user.nickname}</span>
              </span>

              <button
                onClick={handleLogout}
                className="text-sm text-text-muted hover:text-danger transition px-2 py-1 rounded-md cursor-pointer border-none bg-transparent font-medium"
              >
                退出
              </button>
            </>
          ) : (
            <div className="flex items-center gap-2">
              <Link
                to="/login"
                className="text-sm font-medium text-text-secondary hover:text-text transition no-underline px-3 py-1.5"
              >
                登录
              </Link>
              <Link
                to="/register"
                className="text-sm font-medium text-white bg-brand-600 hover:bg-brand-700 transition no-underline px-4 py-1.5 rounded-lg"
              >
                注册
              </Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}

export default Navbar;
