import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiGet } from '../api/client';

interface UserItem {
  email: string;
  nickname: string;
  student_id: string;
  department: string;
  tags: string[];
  avatar: string;
}

function FindPeople() {
  const [users, setUsers] = useState<UserItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    apiGet<{ users: UserItem[] }>('/users')
      .then((res) => setUsers(res.users))
      .catch(() => setError('加载用户列表失败'))
      .finally(() => setLoading(false));
  }, []);

  // ── 字母头像生成 ──
  const avatarStyle = (name: string, avatarUrl: string): React.CSSProperties => {
    const colors = ['#1677ff', '#52c41a', '#fa8c16', '#eb2f96', '#722ed1', '#13c2c2', '#f5222d', '#faad14'];
    const idx = name.charCodeAt(0) % colors.length;
    return {
      width: 48,
      height: 48,
      borderRadius: '50%',
      background: avatarUrl ? `url(${avatarUrl}) center/cover` : colors[idx],
      color: '#fff',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: 20,
      fontWeight: 700,
      flexShrink: 0,
    };
  };

  // ── 渲染 ──
  if (loading) {
    return <div style={containerStyle}><p>⏳ 加载中...</p></div>;
  }

  if (error) {
    return <div style={containerStyle}><p style={{ color: '#f5222d' }}>❌ {error}</p></div>;
  }

  return (
    <div style={containerStyle}>
      <h1>🔍 找人面试</h1>
      <p style={{ color: '#666' }}>浏览可预约的面试官，找到和你方向匹配的人</p>

      {users.length === 0 ? (
        <div style={emptyStyle}>📭 还没有已验证的用户，快去邀请同学注册吧</div>
      ) : (
        <div style={gridStyle}>
          {users.map((u) => (
            <Link to={`/user/${u.email}`} key={u.email} style={cardStyle}>
              <div style={cardHeader}>
                <div style={avatarStyle(u.nickname, u.avatar)}>
                  {u.avatar ? '' : u.nickname.charAt(0)}
                </div>
                <div>
                  <div style={{ fontWeight: 600, fontSize: 16, color: '#333' }}>{u.nickname}</div>
                  <div style={{ color: '#999', fontSize: 13 }}>{u.department || '未设置院系'}</div>
                </div>
              </div>
              {u.tags && u.tags.length > 0 && (
                <div style={tagsWrap}>
                  {u.tags.map((t) => (
                    <span key={t} style={tagStyle}>{t}</span>
                  ))}
                </div>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

// ── 样式 ──

const containerStyle: React.CSSProperties = {
  padding: 40,
  maxWidth: 1000,
  margin: '0 auto',
  fontFamily: 'system-ui',
};

const emptyStyle: React.CSSProperties = {
  textAlign: 'center',
  padding: 60,
  color: '#999',
  fontSize: 16,
};

const gridStyle: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
  gap: 16,
  marginTop: 24,
};

const cardStyle: React.CSSProperties = {
  background: '#fff',
  border: '1px solid #e8e8e8',
  borderRadius: 10,
  padding: 20,
  textDecoration: 'none',
  color: 'inherit',
  transition: 'box-shadow 0.2s',
};

const cardHeader: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 12,
};

const tagsWrap: React.CSSProperties = {
  display: 'flex',
  flexWrap: 'wrap',
  gap: 6,
  marginTop: 12,
};

const tagStyle: React.CSSProperties = {
  display: 'inline-block',
  padding: '2px 10px',
  borderRadius: 12,
  background: '#e6f7ff',
  color: '#1677ff',
  fontSize: 12,
};

export default FindPeople;
