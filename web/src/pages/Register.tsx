import { useState } from 'react';
import { Link } from 'react-router-dom';
import { apiPost } from '../api/client';

interface RegisterResponse {
  message: string;
  user: {
    email: string;
    nickname: string;
  };
}

function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [nickname, setNickname] = useState('');
  const [studentId, setStudentId] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!email || !password || !nickname || !studentId) {
      setError('请填写所有字段');
      return;
    }

    if (!email.endsWith('.edu')) {
      setError('仅支持 .edu 结尾的学生邮箱注册');
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
        password,
        nickname,
        student_id: studentId,
      });
      setSuccess(true);
      console.log('注册成功:', data);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ||
        '注册失败，请稍后重试';
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-[calc(100vh-56px)] flex items-center justify-center px-4">
        <div className="w-full max-w-sm">
          <h1 className="text-2xl font-bold text-text text-center mb-8">注册成功</h1>
          <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
            <div className="text-sm text-text-secondary leading-relaxed space-y-3">
              <p>
                请查看你的邮箱 <strong className="text-text">{email}</strong> 中的验证链接。
              </p>
              <p className="text-text-muted text-xs">
                （开发阶段：验证链接已打印到后端控制台日志）
              </p>
              <p className="text-text-muted text-xs">
                点击验证链接后即可
                <Link to="/login" className="text-brand-600 font-medium ml-1">登录</Link>
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

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

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">
                邮箱 <span className="text-text-muted font-normal">（.edu 学生邮箱）</span>
              </span>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="student@university.edu"
                className="px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt disabled:opacity-50"
                disabled={loading}
              />
            </label>

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

            <div className="grid grid-cols-2 gap-3">
              <label className="flex flex-col gap-1.5">
                <span className="text-sm font-medium text-text-secondary">昵称</span>
                <input
                  type="text"
                  value={nickname}
                  onChange={(e) => setNickname(e.target.value)}
                  placeholder="你的昵称"
                  className="px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt disabled:opacity-50"
                  disabled={loading}
                />
              </label>
              <label className="flex flex-col gap-1.5">
                <span className="text-sm font-medium text-text-secondary">学号</span>
                <input
                  type="text"
                  value={studentId}
                  onChange={(e) => setStudentId(e.target.value)}
                  placeholder="你的学号"
                  className="px-4 py-2.5 rounded-xl border border-border text-sm bg-surface-alt disabled:opacity-50"
                  disabled={loading}
                />
              </label>
            </div>

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
