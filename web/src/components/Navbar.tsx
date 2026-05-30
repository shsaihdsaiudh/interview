import { Link, useNavigate } from 'react-router-dom';
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

// 自定义事件，用于跨组件同步登录状态
const AUTH_CHANGE = 'auth-change';
export function notifyAuthChange(): void {
  window.dispatchEvent(new Event(AUTH_CHANGE));
}

function Navbar() {
  const navigate = useNavigate();
  const [user, setUserState] = useState<UserInfo | null>(getUser());

  useEffect(() => {
    const handler = () => setUserState(getUser());
    window.addEventListener(AUTH_CHANGE, handler);
    return () => window.removeEventListener(AUTH_CHANGE, handler);
  }, []);

  const handleLogout = () => {
    removeToken();
    clearUser();
    setUserState(null);
    navigate('/');
  };

  return (
    <nav style={navStyle}>
      <Link to="/" style={logoStyle}>
        🎯 面试互助平台
      </Link>

      <div style={rightStyle}>
        {user ? (
          <>
            <span style={nicknameStyle}>👤 {user.nickname}</span>
            <button onClick={handleLogout} style={btnStyle}>
              退出
            </button>
          </>
        ) : (
          <>
            <Link to="/login" style={linkStyle}>
              登录
            </Link>
            <Link to="/register" style={{ ...linkStyle, ...registerBtnStyle }}>
              注册
            </Link>
          </>
        )}
      </div>
    </nav>
  );
}

const navStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '12px 40px',
  background: '#fff',
  borderBottom: '1px solid #e8e8e8',
  fontFamily: 'system-ui',
};

const logoStyle: React.CSSProperties = {
  fontSize: 20,
  fontWeight: 700,
  textDecoration: 'none',
  color: '#1677ff',
};

const rightStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 16,
};

const nicknameStyle: React.CSSProperties = {
  fontSize: 16,
  color: '#333',
};

const linkStyle: React.CSSProperties = {
  padding: '8px 20px',
  borderRadius: 6,
  textDecoration: 'none',
  fontSize: 15,
  color: '#1677ff',
  border: '1px solid #1677ff',
  background: '#fff',
  cursor: 'pointer',
};

const registerBtnStyle: React.CSSProperties = {
  background: '#1677ff',
  color: '#fff',
};

const btnStyle: React.CSSProperties = {
  padding: '8px 20px',
  borderRadius: 6,
  border: '1px solid #e8e8e8',
  background: '#fff',
  cursor: 'pointer',
  fontSize: 15,
  color: '#666',
};

export default Navbar;
