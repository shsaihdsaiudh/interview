import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { apiPost, setToken } from '../api/client';
import { setUser, notifyAuthChange } from '../components/Navbar';

interface LoginResponse {
  token: string;
  user: {
    email: string;
    nickname: string;
    student_id: string;
    email_verified: boolean;
    created_at: string;
  };
}

function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !password) {
      setError('请填写邮箱和密码');
      return;
    }

    setLoading(true);
    try {
      const data = await apiPost<LoginResponse>('/auth/login', { email, password });
      setToken(data.token);
      setUser({ email: data.user.email, nickname: data.user.nickname });
      notifyAuthChange();
      navigate('/');
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        '登录失败，请稍后重试';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={containerStyle}>
      <h1 style={titleStyle}>登录</h1>

      {error && <div style={errorStyle}>{error}</div>}

      <form onSubmit={handleSubmit} style={formStyle}>
        <label style={labelStyle}>
          邮箱
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="student@university.edu"
            style={inputStyle}
            disabled={loading}
          />
        </label>

        <label style={labelStyle}>
          密码
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="请输入密码"
            style={inputStyle}
            disabled={loading}
          />
        </label>

        <button type="submit" disabled={loading} style={submitStyle}>
          {loading ? '登录中...' : '登录'}
        </button>
      </form>

      <p style={hintStyle}>
        还没有账号？<Link to="/register">立即注册</Link>
      </p>
    </div>
  );
}

// ── 样式 ──

const containerStyle: React.CSSProperties = {
  maxWidth: 400,
  margin: '60px auto',
  padding: 40,
  fontFamily: 'system-ui',
  background: '#fff',
  borderRadius: 12,
  boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
};

const titleStyle: React.CSSProperties = {
  fontSize: 28,
  marginBottom: 24,
  textAlign: 'center',
};

const formStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 16,
};

const labelStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 6,
  fontSize: 14,
  color: '#555',
};

const inputStyle: React.CSSProperties = {
  padding: '10px 14px',
  borderRadius: 6,
  border: '1px solid #d9d9d9',
  fontSize: 15,
  outline: 'none',
  transition: 'border-color .2s',
};

const submitStyle: React.CSSProperties = {
  padding: '12px 0',
  borderRadius: 6,
  border: 'none',
  background: '#1677ff',
  color: '#fff',
  fontSize: 16,
  fontWeight: 600,
  cursor: 'pointer',
  marginTop: 8,
};

const errorStyle: React.CSSProperties = {
  background: '#fff2f0',
  border: '1px solid #ffccc7',
  color: '#ff4d4f',
  padding: '10px 16px',
  borderRadius: 6,
  marginBottom: 16,
  fontSize: 14,
};

const hintStyle: React.CSSProperties = {
  textAlign: 'center',
  marginTop: 20,
  fontSize: 14,
  color: '#888',
};

export default Login;
