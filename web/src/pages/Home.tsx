import { Link } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { apiGet } from '../api/client';
import { getUser } from '../components/Navbar';

function Home() {
  const [status, setStatus] = useState<string>('');
  const user = getUser();

  useEffect(() => {
    apiGet<{ message: string }>('/ping')
      .then((res) => setStatus(res.message))
      .catch(() => setStatus('后端未连接'));
  }, []);

  return (
    <div style={{ padding: 40, fontFamily: 'system-ui' }}>
      <h1>🎯 面试互助平台</h1>
      <p>校内模拟面试，互相帮助积累经验</p>

      {user && (
        <div style={welcomeStyle}>
          👋 欢迎回来，<strong>{user.nickname}</strong>
        </div>
      )}

      <div style={{ background: '#f5f5f5', padding: 10, borderRadius: 8, marginBottom: 20 }}>
        后端状态：{status ? `✅ ${status}` : '⏳ 连接中...'}
      </div>

      <nav style={{ display: 'flex', gap: 16 }}>
        <Link to="/find" style={linkStyle}>
          🔍 找人面试
        </Link>
        <Link to="/posts" style={linkStyle}>
          📝 帖子广场
        </Link>
      </nav>
    </div>
  );
}

const welcomeStyle: React.CSSProperties = {
  background: '#e6f7ff',
  border: '1px solid #91d5ff',
  padding: '12px 20px',
  borderRadius: 8,
  marginBottom: 16,
  fontSize: 16,
  color: '#1677ff',
};

const linkStyle: React.CSSProperties = {
  padding: '12px 24px',
  background: '#1677ff',
  color: '#fff',
  borderRadius: 8,
  textDecoration: 'none',
  fontSize: 18,
};

export default Home;
