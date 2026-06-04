import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { apiPost, getApiErrorMessage } from '../api/client';
import { getUser } from '../components/Navbar';

const EMAIL_SUFFIX = '@std.uestc.edu.cn';
type Step = 'input-email' | '重置密码' | 'done';

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
  if (currentUser) { navigate('/', { replace: true }); return null; }

  const handleSendCode = async () => {
    setError('');
    if (!studentId) { setError('请填写学号'); return; }
    setSending(true);
    try {
      await apiPost('/auth/forgot-password', { email: fullEmail });
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((prev) => { if (prev <= 1) { clearInterval(timer); return 0; } return prev - 1; });
      }, 1000);
    } catch (err: unknown) { setError(getApiErrorMessage(err, '发送失败')); }
    finally { setSending(false); }
  };

  const handleNext = () => { if (!studentId) { setError('请填写学号'); return; } handleSendCode(); setStep('重置密码'); };

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault(); setError('');
    if (!code || !password) { setError('请填写验证码和新密码'); return; }
    if (password.length < 6) { setError('新密码至少 6 位'); return; }
    setLoading(true);
    try { await apiPost('/auth/reset-password', { email: fullEmail, code, password }); setStep('done'); }
    catch (err: unknown) { setError(getApiErrorMessage(err, '重置失败')); }
    finally { setLoading(false); }
  };

  return (
    <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
      <div className="w-full max-w-sm animate-fade-up">
        <div className="text-center mb-8">
          <h1 className="text-text" style={{ fontSize: 26, fontWeight: 700 }}>reset</h1>
          <p className="text-text-muted mt-2" style={{ fontSize: 18 }}>重置密码</p>
        </div>

        <div className="bg-card border border-border pixel-corners p-6">
          {error && (
            <div className="mb-5 px-3 py-2 text-danger pixel-corners-sm"
                 style={{ fontSize: 17, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>
              {error}
            </div>
          )}

          {step === 'input-email' && (
            <div>
              <p className="text-text-secondary mb-4" style={{ fontSize: 17 }}>输入学号，发送重置验证码到学校邮箱</p>
              <label className="flex flex-col gap-1 mb-4">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>学号</span>
                <div className="flex">
                  <input type="text" value={studentId}
                    onChange={(e) => { setStudentId(e.target.value); setError(''); }}
                    placeholder="2024010914026"
                    className="flex-1 min-w-0 px-3 py-2 bg-surface border border-border text-text"
                    style={{ fontSize: 18, borderRight: 0, clipPath: 'polygon(0 2px, 2px 2px, 2px 0, 100% 0, 100% 100%, 0 100%)' }}
                    onKeyDown={(e) => { if (e.key === 'Enter') handleNext(); }} />
                  <span className="px-2 py-2 bg-surface border border-border text-text-muted select-none whitespace-nowrap"
                        style={{ fontSize: 17, borderLeft: 0 }}>
                    @std.uestc.edu.cn
                  </span>
                </div>
              </label>
              <button onClick={handleNext} disabled={sending}
                className="pixel-btn primary w-full justify-center" style={{ fontSize: 18, padding: '10px' }}>
                {sending ? '...' : '发送验证码'}
              </button>
              <div className="text-center mt-4">
                <Link to="/login" className="no-underline text-text-muted hover:text-text-secondary" style={{ fontSize: 18 }}>返回登录</Link>
              </div>
            </div>
          )}

          {step === '重置密码' && (
            <form onSubmit={handleReset}>
              <p className="text-text-secondary mb-4" style={{ fontSize: 17 }}>验证码已发送到 <strong className="text-text">{fullEmail}</strong></p>
              <label className="flex flex-col gap-1 mb-3">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>验证码</span>
                <div className="flex gap-2">
                  <input type="text" value={code} onChange={(e) => { setCode(e.target.value); setError(''); }}
                    placeholder="6 位验证码" maxLength={6} disabled={loading} autoFocus
                    className="flex-1 px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} />
                  <button type="button" onClick={handleSendCode} disabled={countdown > 0 || sending}
                    className="pixel-btn whitespace-nowrap flex-shrink-0" style={{ fontSize: 17, padding: '8px 12px' }}>
                    {countdown > 0 ? `${countdown}s` : '重新发送'}
                  </button>
                </div>
              </label>
              <label className="flex flex-col gap-1 mb-5">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>新密码</span>
                <input type="password" value={password} onChange={(e) => { setPassword(e.target.value); setError(''); }}
                  placeholder="至少 6 位" disabled={loading}
                  className="w-full px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} />
              </label>
              <button type="submit" disabled={loading}
                className="pixel-btn primary w-full justify-center" style={{ fontSize: 18, padding: '10px' }}>
                {loading ? '...' : '重置'}
              </button>
            </form>
          )}

          {step === 'done' && (
            <div className="text-center animate-fade-up">
              <div className="w-12 h-12 flex items-center justify-center mx-auto mb-3"
                   style={{ border: '1px solid var(--color-success)' }}>
                <span style={{ fontSize: 20, color: 'var(--color-success)' }}>OK</span>
              </div>
              <h3 className="text-text mb-2" style={{ fontSize: 20, fontWeight: 700 }}>密码重置成功</h3>
              <p className="text-text-secondary mb-5" style={{ fontSize: 17 }}>请使用新密码登录</p>
              <button onClick={() => navigate('/login')}
                className="pixel-btn primary" style={{ fontSize: 18, padding: '8px 24px' }}>
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
