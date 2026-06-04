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
      className={`no-underline px-3 py-1 transition-colors ${
        isActive(path)
          ? 'text-brand-600'
          : 'text-text-muted hover:text-text-secondary'
      }`}
      style={{ fontSize: 18 }}
    >
      {label}
    </Link>
  );

  return (
    <nav
      className={`sticky top-0 z-50 transition ${
        scrolled
          ? 'glass border-b border-border'
          : 'bg-transparent border-b border-transparent'
      }`}
    >
      <div className="max-w-6xl mx-auto px-6 h-14 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2.5 no-underline">
          <div
            className="grid gap-px flex-shrink-0"
            style={{ gridTemplateColumns: '5px 5px', width: 12, height: 12 }}
          >
            <span style={{ display: 'block', background: 'var(--color-brand-600)' }} />
            <span style={{ display: 'block', background: 'var(--color-brand-600)', opacity: 0.6 }} />
            <span style={{ display: 'block', background: 'var(--color-brand-600)', opacity: 0.6 }} />
            <span style={{ display: 'block', background: 'var(--color-brand-600)', opacity: 0.3 }} />
          </div>
          <span className="text-text-secondary tracking-wide" style={{ fontSize: 18 }}>
            mock·io
          </span>
        </Link>

        <div className="flex items-center gap-1">
          {user ? (
            <>
              {navLink('/find', '发现')}
              {navLink('/appointments', '预约')}
              {navLink('/my-card', '名片')}

              <span className="w-px h-4 bg-border mx-2" />

              <Link
                to={`/user/${user.email}`}
                className="no-underline text-text flex items-center gap-1.5 hover:text-brand-600 transition-colors"
                style={{ fontSize: 18 }}
              >
                <span
                  className="w-6 h-6 flex items-center justify-center text-white font-bold"
                  style={{ background: 'var(--color-brand-600)', fontSize: 14 }}
                >
                  {user.nickname.charAt(0)}
                </span>
                <span className="hidden sm:inline">{user.nickname}</span>
              </Link>

              <button
                onClick={handleLogout}
                className="cursor-pointer border-none bg-transparent text-text-muted hover:text-danger transition-colors px-2 py-1"
                style={{ fontSize: 18 }}
              >
                退出
              </button>
            </>
          ) : (
            <div className="flex items-center gap-2">
              <Link
                to="/login"
                className="no-underline text-text-muted hover:text-text-secondary transition-colors px-3 py-1.5"
                style={{ fontSize: 18 }}
              >
                sign in
              </Link>
              <Link
                to="/register"
                className="no-underline text-white px-4 py-1.5 transition pixel-corners-sm"
                style={{ fontSize: 16, background: 'var(--color-brand-600)' }}
              >
                sign up
              </Link>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}

export default Navbar;
