import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { apiPost, getApiErrorMessage, setToken } from '../api/client';
import { setUser, notifyAuthChange } from '../components/Navbar';

interface LoginResponse {
  token: string;
  user: {
    email: string;
    nickname: string;
    student_id: string;
    email_verified: boolean;
    account_status: string;
    created_at: string;
  };
}

const EMAIL_SUFFIX = '@std.uestc.edu.cn';

function Login() {
  const navigate = useNavigate();
  const [studentId, setStudentId] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!studentId || !password) {
      setError('请填写学号和密码');
      return;
    }

    setLoading(true);
    try {
      const data = await apiPost<LoginResponse>('/auth/login', { email: studentId + EMAIL_SUFFIX, password });
      setToken(data.token);
      setUser({ email: data.user.email, nickname: data.user.nickname });
      notifyAuthChange();
      navigate('/');
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, '登录失败，请稍后重试'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
      <div className="w-full max-w-lg">
        <h1 className="text-2xl font-bold text-text text-center mb-8">登录</h1>

        <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-xl text-sm mb-6">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">学号</span>
              <div className="flex items-center">
                <input
                  type="text"
                  value={studentId}
                  onChange={(e) => setStudentId(e.target.value)}
                  placeholder="2024010914026"
                  className="flex-1 min-w-0 px-4 py-2.5 border border-border rounded-l-xl border-r-0 outline-none bg-surface-alt text-sm disabled:opacity-50"
                  disabled={loading}
                />
                <span className="px-3 py-2.5 border border-border rounded-r-xl bg-surface-alt text-sm text-text-secondary select-none whitespace-nowrap flex-shrink-0">@std.uestc.edu.cn</span>
              </div>
            </label>

            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">密码</span>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="请输入密码"
                className="px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt disabled:opacity-50"
                disabled={loading}
              />
            </label>

            <button
              type="submit"
              disabled={loading}
              className="mt-2 w-full py-2.5 rounded-xl bg-brand-600 hover:bg-brand-700 text-white font-medium text-sm
                         transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer border-none"
            >
              {loading ? '登录中...' : '登录'}
            </button>
          </form>

          <p className="text-center mt-6 text-sm text-text-muted">
            还没有账号？
            <Link to="/register" className="text-brand-600 hover:text-brand-700 font-medium ml-1">
              立即注册
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

export default Login;
