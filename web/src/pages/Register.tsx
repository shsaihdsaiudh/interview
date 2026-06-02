import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { apiPost, getApiErrorMessage, setToken } from '../api/client';
import { setUser, notifyAuthChange } from '../components/Navbar';

interface RegisterResponse {
  token: string;
  user: {
    email: string;
    nickname: string;
    email_verified: boolean;
    account_status: string;
  };
}

function Register() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [sending, setSending] = useState(false);
  const [codeSent, setCodeSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);

  const handleSendCode = async () => {
    setError('');
    if (!email) {
      setError('请填写邮箱');
      return;
    }
    if (!email.endsWith('@std.uestc.edu.cn')) {
      setError('仅支持 @std.uestc.edu.cn 邮箱注册');
      return;
    }

    setSending(true);
    try {
      await apiPost('/auth/send-code', { email });
      setCodeSent(true);
      setCountdown(60);
      // 倒计时
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, '发送验证码失败'));
    } finally {
      setSending(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !code || !password) {
      setError('请填写所有字段');
      return;
    }
    if (password.length < 6) {
      setError('密码至少需要 6 个字符');
      return;
    }

    setLoading(true);
    try {
      const data = await apiPost<RegisterResponse>('/auth/register', {
        email,
        code,
        password,
      });
      setToken(data.token);
      setUser({ email: data.user.email, nickname: data.user.nickname });
      notifyAuthChange();
      navigate('/');
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, '注册失败，请稍后重试'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
      <div className="w-full max-w-sm">
        <h1 className="text-2xl font-bold text-text text-center mb-8">创建账号</h1>

        <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-xl text-sm mb-6">
              {error}
            </div>
          )}

          <form onSubmit={handleRegister} className="flex flex-col gap-4">
            {/* 邮箱 + 发送验证码 */}
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">
                邮箱 <span className="text-text-muted font-normal">（@std.uestc.edu.cn）</span>
              </span>
              <div className="flex gap-2">
                <input
                  type="email"
                  value={email}
                  onChange={(e) => { setEmail(e.target.value); setCodeSent(false); }}
                  placeholder="2024010914026@std.uestc.edu.cn"
                  className="flex-1 px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt disabled:opacity-50"
                  disabled={loading}
                />
                <button
                  type="button"
                  onClick={handleSendCode}
                  disabled={sending || countdown > 0 || loading}
                  className="px-4 py-2.5 rounded-xl bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium
                             transition disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer border-none whitespace-nowrap"
                >
                  {sending ? '发送中...' : countdown > 0 ? `${countdown}s` : '发送验证码'}
                </button>
              </div>
            </label>

            {/* 验证码 */}
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">验证码</span>
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="请输入 6 位验证码"
                maxLength={6}
                className="px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt disabled:opacity-50"
                disabled={loading}
              />
              {codeSent && (
                <span className="text-xs text-text-muted">
                  开发阶段：验证码已打印到后端控制台日志
                </span>
              )}
            </label>

            {/* 密码 */}
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">密码</span>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="至少 6 位密码"
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
              {loading ? '注册中...' : '注册'}
            </button>
          </form>

          <p className="text-center mt-6 text-sm text-text-muted">
            已有账号？
            <Link to="/login" className="text-brand-600 hover:text-brand-700 font-medium ml-1">
              立即登录
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}

export default Register;
