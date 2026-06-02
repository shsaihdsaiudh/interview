import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { apiPost, getApiErrorMessage } from '../api/client';
import { getUser } from '../components/Navbar';

const EMAIL_SUFFIX = '@std.uestc.edu.cn';

type Step = 'input-email' | 'reset' | 'done';

function ForgotPassword() {
  const navigate = useNavigate();
  const currentUser = getUser();

  const [step, setStep] = useState<Step>('input-email');
  const [studentId, setStudentId] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [sending, setSending] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);

  const fullEmail = studentId + EMAIL_SUFFIX;

  // 已登录用户跳转首页
  if (currentUser) {
    navigate('/', { replace: true });
    return null;
  }

  const handleSendCode = async () => {
    setError('');
    if (!studentId) {
      setError('请填写学号');
      return;
    }

    setSending(true);
    try {
      await apiPost('/auth/forgot-password', { email: fullEmail });
      // 倒计时
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) { clearInterval(timer); return 0; }
          return prev - 1;
        });
      }, 1000);
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, '发送失败'));
    } finally {
      setSending(false);
    }
  };

  const handleNext = () => {
    if (!studentId) { setError('请填写学号'); return; }
    handleSendCode();
    setStep('reset');
  };

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!code || !password) {
      setError('请填写验证码和新密码');
      return;
    }
    if (password.length < 6) {
      setError('新密码至少 6 位');
      return;
    }

    setLoading(true);
    try {
      await apiPost('/auth/reset-password', {
        email: fullEmail,
        code,
        password,
      });
      setStep('done');
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, '重置失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
      <div className="w-full max-w-lg">
        <h1 className="text-2xl font-bold text-text text-center mb-8">找回密码</h1>

        <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
          {error && (
            <div className="bg-red-50 border border-red-200 rounded-xl px-4 py-3 text-sm text-red-600 mb-5">
              {error}
            </div>
          )}

          {step === 'input-email' && (
            <div>
              <p className="text-sm text-text-secondary mb-4">请输入你的学号，我们将发送重置验证码到你的学校邮箱</p>
              <label className="flex flex-col gap-1.5 mb-4">
                <span className="text-sm font-medium text-text-secondary">学号</span>
                <div className="flex items-center">
                  <input
                    type="text"
                    value={studentId}
                    onChange={(e) => { setStudentId(e.target.value); setError(''); }}
                    placeholder="2024010914026"
                    className="flex-1 min-w-0 px-4 py-2.5 border border-border rounded-l-xl border-r-0 outline-none bg-surface-alt text-sm"
                    onKeyDown={(e) => { if (e.key === 'Enter') handleNext(); }}
                  />
                  <span className="px-3 py-2.5 border border-border rounded-r-xl bg-surface-alt text-sm text-text-secondary select-none whitespace-nowrap flex-shrink-0">
                    @std.uestc.edu.cn
                  </span>
                </div>
              </label>
              <button
                onClick={handleNext}
                disabled={sending}
                className="w-full py-2.5 rounded-xl bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50"
              >
                {sending ? '发送中...' : '发送重置验证码'}
              </button>
              <div className="text-center mt-4">
                <Link to="/login" className="text-sm text-text-muted hover:text-brand-600 transition no-underline">
                  返回登录
                </Link>
              </div>
            </div>
          )}

          {step === 'reset' && (
            <form onSubmit={handleReset}>
              <p className="text-sm text-text-secondary mb-4">
                验证码已发送到 <strong>{fullEmail}</strong>，请查收邮件
              </p>

              <label className="flex flex-col gap-1.5 mb-3">
                <span className="text-sm font-medium text-text-secondary">验证码</span>
                <div className="flex gap-2 items-center">
                  <input
                    type="text"
                    value={code}
                    onChange={(e) => { setCode(e.target.value); setError(''); }}
                    placeholder="6 位验证码"
                    maxLength={6}
                    className="flex-1 px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt outline-none disabled:opacity-50"
                    disabled={loading}
                    autoFocus
                  />
                  <button
                    type="button"
                    onClick={handleSendCode}
                    disabled={countdown > 0 || sending}
                    className="px-3 py-2.5 rounded-xl border border-border bg-surface-alt text-sm text-text-secondary hover:text-brand-600 hover:border-brand-200 transition cursor-pointer disabled:opacity-40 whitespace-nowrap flex-shrink-0"
                  >
                    {countdown > 0 ? `${countdown}s` : '重新发送'}
                  </button>
                </div>
              </label>

              <label className="flex flex-col gap-1.5 mb-5">
                <span className="text-sm font-medium text-text-secondary">新密码</span>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => { setPassword(e.target.value); setError(''); }}
                  placeholder="至少 6 位"
                  className="px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt outline-none disabled:opacity-50"
                  disabled={loading}
                />
              </label>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 rounded-xl bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50"
              >
                {loading ? '重置中...' : '重置密码'}
              </button>
            </form>
          )}

          {step === 'done' && (
            <div className="text-center">
              <div className="w-12 h-12 rounded-full bg-emerald-100 flex items-center justify-center mx-auto mb-3">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              </div>
              <h3 className="text-lg font-bold text-text mb-2">密码重置成功</h3>
              <p className="text-sm text-text-secondary mb-5">请使用新密码登录</p>
              <button
                onClick={() => navigate('/login')}
                className="px-6 py-2.5 rounded-xl bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none"
              >
                去登录
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default ForgotPassword;