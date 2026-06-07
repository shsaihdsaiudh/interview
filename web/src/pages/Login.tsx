import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { apiPost, getApiErrorMessage, setToken } from '../api/client';
import { setUser, notifyAuthChange, getUser } from '../components/Navbar';

interface LoginResponse {
  token: string;
  user: { email: string; nickname: string; student_id: string; email_verified: boolean; account_status: string; created_at: string; };
}

const EMAIL_SUFFIX = '@std.uestc.edu.cn';

function Login() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const currentUser = getUser();
  const redirect = searchParams.get('redirect');

  const [studentId, setStudentId] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (currentUser) {
    navigate(redirect || '/', { replace: true });
    return null;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!studentId || !password) { setError('请填写学号和密码'); return; }
    setLoading(true);
    try {
      const data = await apiPost<LoginResponse>('/auth/login', { email: studentId + EMAIL_SUFFIX, password });
      setToken(data.token);
      setUser({ email: data.user.email, nickname: data.user.nickname });
      notifyAuthChange();
      navigate(redirect || '/');
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, '登录失败'));
    } finally { setLoading(false); }
  };

  return (
    <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
      <div className="w-full max-w-sm animate-fade-up">
        <div className="text-center mb-8">
          <h1 className="text-text" style={{ fontSize: 26, fontWeight: 700 }}>登录</h1>
          <p className="text-text-muted mt-2" style={{ fontSize: 18 }}>登录你的账号</p>
        </div>

        <div className="bg-card border border-border pixel-corners p-6">
          {error && (
            <div className="mb-5 px-3 py-2 text-danger pixel-corners-sm"
                 style={{ fontSize: 17, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <label className="flex flex-col gap-1">
              <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>学号</span>
              <div className="flex">
                <input type="text" value={studentId} onChange={(e) => setStudentId(e.target.value)}
                  placeholder="请输入学号" disabled={loading}
                  className="flex-1 min-w-0 px-3 py-2 bg-surface border border-border text-text pixel-corners-sm"
                  style={{ fontSize: 18, borderRight: 0, clipPath: 'polygon(0 2px, 2px 2px, 2px 0, 100% 0, 100% 100%, 0 100%)' }} />
                <span className="px-2 py-2 bg-surface border border-border text-text-muted select-none whitespace-nowrap"
                      style={{ fontSize: 17, borderLeft: 0 }}>
                  @std.uestc.edu.cn
                </span>
              </div>
            </label>

            <label className="flex flex-col gap-1">
              <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>密码</span>
              <input type="password" value={password} onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码" disabled={loading}
                className="w-full px-3 py-2 bg-surface border border-border text-text pixel-corners-sm"
                style={{ fontSize: 18 }} />
            </label>

            <div className="flex justify-end">
              <Link to="/forgot-password" className="no-underline text-text-muted hover:text-text-secondary transition-colors" style={{ fontSize: 18 }}>
                忘记密码?
              </Link>
            </div>

            <button type="submit" disabled={loading}
              className="pixel-btn primary w-full justify-center"
              style={{ fontSize: 18, padding: '10px' }}>
              {loading ? '...' : '登录'}
            </button>
          </form>

          <p className="text-center mt-5 text-text-muted" style={{ fontSize: 17 }}>
            没有账号?
            <Link to="/register" className="no-underline text-brand-600 hover:text-brand-700 ml-1">注册</Link>
          </p>
        </div>
      </div>
    </div>
  );
}

export default Login;
