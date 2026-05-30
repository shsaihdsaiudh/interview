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

    // 前端校验
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
      // 开发提示
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

  // 注册成功后的提示页
  if (success) {
    return (
      <div style={containerStyle}>
        <h1 style={titleStyle}>🎉 注册成功</h1>
        <div style={successBoxStyle}>
          <p>请查看你的邮箱 <strong>{email}</strong> 中的验证链接。</p>
          <p style={{ color: '#888', fontSize: 14 }}>
            （开发阶段：验证链接已打印到后端控制台日志，请查看服务端输出）
          </p>
          <p style={{ color: '#888', fontSize: 14 }}>
            点击验证链接后即可
            <Link to="/login"> 登录</Link>
          </p>
        </div>
      </div>
    );
  }

  return (
    <div style={containerStyle}>
      <h1 style={titleStyle}>注册</h1>

      {error && <div style={errorStyle}>{error}</div>}

      <form onSubmit={handleSubmit} style={formStyle}>
        <label style={labelStyle}>
          邮箱 <span style={{ color: '#999' }}>（仅限 .edu 学生邮箱）</span>
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
            placeholder="至少 6 位密码"
            style={inputStyle}
            disabled={loading}
          />
        </label>

        <label style={labelStyle}>
          昵称
          <input
            type="text"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            placeholder="你的昵称"
            style={inputStyle}
            disabled={loading}
          />
        </label>

        <label style={labelStyle}>
          学号
          <input
            type="text"
            value={studentId}
            onChange={(e) => setStudentId(e.target.value)}
            placeholder="你的学号"
            style={inputStyle}
            disabled={loading}
          />
        </label>

        <button type="submit" disabled={loading} style={submitStyle}>
          {loading ? '注册中...' : '注册'}
        </button>
      </form>

      <p style={hintStyle}>
        已有账号？<Link to="/login">立即登录</Link>
      </p>
    </div>
  );
}

// ── 样式 ──

const containerStyle: React.CSSProperties = {
  maxWidth: 440,
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

const successBoxStyle: React.CSSProperties = {
  background: '#f6ffed',
  border: '1px solid #b7eb8f',
  padding: 24,
  borderRadius: 8,
  lineHeight: 1.8,
  fontSize: 15,
};

const hintStyle: React.CSSProperties = {
  textAlign: 'center',
  marginTop: 20,
  fontSize: 14,
  color: '#888',
};

export default Register;
