import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { apiPost, getApiErrorMessage, setToken } from '../api/client';
import { setUser, notifyAuthChange, getUser } from '../components/Navbar';

interface RegisterResponse {
  token: string;
  user: { email: string; nickname: string; email_verified: boolean; account_status: string; };
}

const EMAIL_SUFFIX = '@std.uestc.edu.cn';

function Register() {
  const navigate = useNavigate();
  const currentUser = getUser();

  const [studentId, setStudentId] = useState('');
  const [code, setCode] = useState('');
  const [password, setPassword] = useState('');
  const [sending, setSending] = useState(false);
  const [codeSent, setCodeSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [countdown, setCountdown] = useState(0);

  if (currentUser) { navigate('/', { replace: true }); return null; }

  const fullEmail = studentId + EMAIL_SUFFIX;

  const handleSendCode = async () => {
    setError('');
    if (!studentId) { setError('请填写学号'); return; }
    setSending(true);
    try {
      await apiPost('/auth/send-code', { email: fullEmail });
      setCodeSent(true);
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((prev) => { if (prev <= 1) { clearInterval(timer); return 0; } return prev - 1; });
      }, 1000);
    } catch (err: unknown) { setError(getApiErrorMessage(err, '发送失败')); }
    finally { setSending(false); }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!studentId || !code || !password) { setError('请填写所有字段'); return; }
    if (password.length < 6) { setError('密码至少需要 6 个字符'); return; }
    setLoading(true);
    try {
      const data = await apiPost<RegisterResponse>('/auth/register', { email: fullEmail, code, password });
      setToken(data.token);
      setUser({ email: data.user.email, nickname: data.user.nickname });
      notifyAuthChange();
      navigate('/');
    } catch (err: unknown) { setError(getApiErrorMessage(err, '注册失败')); }
    finally { setLoading(false); }
  };

  return (
    <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
      <div className="w-full max-w-sm animate-fade-up">
        <div className="text-center mb-8">
          <h1 className="text-text" style={{ fontSize: 26, fontWeight: 700 }}>注册</h1>
          <p className="text-text-muted mt-2" style={{ fontSize: 18 }}>创建你的账号</p>
        </div>

        <div className="bg-card border border-border pixel-corners p-6">
          {error && (
            <div className="mb-5 px-3 py-2 text-danger pixel-corners-sm"
                 style={{ fontSize: 17, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>
              {error}
            </div>
          )}

          <form onSubmit={handleRegister} className="flex flex-col gap-4">
            <label className="flex flex-col gap-1">
              <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>学号</span>
              <div className="flex gap-2">
                <div className="flex flex-1 min-w-0">
                  <input type="text" value={studentId}
                    onChange={(e) => { setStudentId(e.target.value); setCodeSent(false); }}
                    placeholder="2024010914026" disabled={loading}
                    className="flex-1 min-w-0 px-3 py-2 bg-surface border border-border text-text pixel-corners-sm"
                    style={{ fontSize: 18, borderRight: 0, clipPath: 'polygon(0 2px, 2px 2px, 2px 0, 100% 0, 100% 100%, 0 100%)' }} />
                  <span className="px-2 py-2 bg-surface border border-border text-text-muted select-none whitespace-nowrap"
                        style={{ fontSize: 17, borderLeft: 0 }}>
                    @std.uestc.edu.cn
                  </span>
                </div>
                <button type="button" onClick={handleSendCode}
                  disabled={sending || countdown > 0 || loading}
                  className="pixel-btn primary whitespace-nowrap flex-shrink-0"
                  style={{ fontSize: 17, padding: '8px 12px' }}>
                  {sending ? '...' : countdown > 0 ? `${countdown}s` : '发送验证码'}
                </button>
              </div>
            </label>

            <label className="flex flex-col gap-1">
              <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>验证码</span>
              <input type="text" value={code} onChange={(e) => setCode(e.target.value)}
                placeholder="6 位验证码" maxLength={6} disabled={loading}
                className="w-full px-3 py-2 bg-surface border border-border text-text pixel-corners-sm"
                style={{ fontSize: 18 }} />
              {codeSent && <span className="text-text-muted" style={{ fontSize: 18 }}>开发阶段：验证码已打印到后端控制台</span>}
            </label>

            <label className="flex flex-col gap-1">
              <span className="text-text-muted tracking-wider" style={{ fontSize: 18 }}>密码</span>
              <input type="password" value={password} onChange={(e) => setPassword(e.target.value)}
                placeholder="至少 6 位" disabled={loading}
                className="w-full px-3 py-2 bg-surface border border-border text-text pixel-corners-sm"
                style={{ fontSize: 18 }} />
            </label>

            <button type="submit" disabled={loading}
              className="pixel-btn primary w-full justify-center mt-1"
              style={{ fontSize: 18, padding: '10px' }}>
              {loading ? '...' : '注册'}
            </button>
          </form>

          <p className="text-center mt-5 text-text-muted" style={{ fontSize: 17 }}>
            已有账号?
            <Link to="/login" className="no-underline text-brand-600 hover:text-brand-700 ml-1">登录</Link>
          </p>
        </div>
      </div>
    </div>
  );
}

export default Register;
