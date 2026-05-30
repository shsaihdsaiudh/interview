import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiGet, apiPost } from '../api/client';
import { getUser } from '../components/Navbar';

interface UserInfo {
  email: string;
  nickname: string;
  student_id: string;
  department: string;
  tags: string[];
  avatar: string;
  contact_info: string;
  email_verified: boolean;
}

interface TimeSlot {
  id: string;
  user_id: string;
  date: string;
  start_time: string;
  end_time: string;
}

interface DetailData {
  user: UserInfo;
  availabilities: TimeSlot[];
}

const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

function getDayOfWeek(dateStr: string): string {
  const d = new Date(dateStr);
  return dayNames[d.getDay()];
}

function UserDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const currentUser = getUser();

  const [detail, setDetail] = useState<DetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [bookingId, setBookingId] = useState<string | null>(null);
  const [message, setMessage] = useState('');
  const [bookingError, setBookingError] = useState('');
  const [bookingSuccess, setBookingSuccess] = useState('');

  useEffect(() => {
    if (!id) return;
    apiGet<DetailData>(`/users/${id}`)
      .then(setDetail)
      .catch(() => setError('加载用户详情失败'))
      .finally(() => setLoading(false));
  }, [id]);

  const handleBook = async () => {
    if (!bookingId) return;
    if (!currentUser) {
      navigate('/login');
      return;
    }
    setBookingError('');
    setBookingSuccess('');
    try {
      await apiPost('/appointments', {
        time_slot_id: bookingId,
        message: message || '希望预约一场模拟面试',
      });
      setBookingSuccess('预约成功！请等待对方确认');
      setBookingId(null);
      setMessage('');
    } catch (err: any) {
      const msg = err?.response?.data?.error || '预约失败';
      setBookingError(msg);
    }
  };

  // ── 头像 ──
  const avatarStyle = (name: string, avatarUrl: string): React.CSSProperties => {
    const colors = ['#1677ff', '#52c41a', '#fa8c16', '#eb2f96', '#722ed1', '#13c2c2'];
    const idx = name.charCodeAt(0) % colors.length;
    return {
      width: 64,
      height: 64,
      borderRadius: '50%',
      background: avatarUrl ? `url(${avatarUrl}) center/cover` : colors[idx],
      color: '#fff',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontSize: 28,
      fontWeight: 700,
    };
  };

  // ── 渲染 ──
  if (loading) return <div style={container}><p>⏳ 加载中...</p></div>;
  if (error) return <div style={container}><p style={{ color: '#f5222d' }}>❌ {error}</p></div>;
  if (!detail) return <div style={container}><p>用户不存在</p></div>;

  const { user, availabilities } = detail;
  const isSelf = currentUser?.email === user.email;

  return (
    <div style={container}>
      {/* 用户信息卡片 */}
      <div style={profileCard}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div style={avatarStyle(user.nickname, user.avatar)}>
            {user.avatar ? '' : user.nickname.charAt(0)}
          </div>
          <div>
            <h1 style={{ margin: 0, fontSize: 24 }}>{user.nickname}</h1>
            <p style={{ margin: '4px 0 0', color: '#666' }}>
              {user.department || '未设置院系'} · {user.student_id}
            </p>
          </div>
        </div>

        {user.tags && user.tags.length > 0 && (
          <div style={{ display: 'flex', gap: 8, marginTop: 16, flexWrap: 'wrap' }}>
            {user.tags.map((t) => (
              <span key={t} style={tagStyle}>{t}</span>
            ))}
          </div>
        )}

        {user.contact_info && (
          <div style={contactStyle}>📞 联系方式：{user.contact_info}</div>
        )}
      </div>

      {/* 空闲时间列表 */}
      <div style={sectionStyle}>
        <h2>📅 空闲时间</h2>
        {availabilities.length === 0 ? (
          <div style={emptyStyle}>暂无空闲时间</div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 16 }}>
            {availabilities.map((slot) => (
              <div
                key={slot.id}
                style={{
                  ...slotCardStyle,
                  borderColor: bookingId === slot.id ? '#1677ff' : '#e8e8e8',
                  background: bookingId === slot.id ? '#e6f7ff' : '#fff',
                }}
                onClick={() => {
                  if (!isSelf) setBookingId(slot.id === bookingId ? null : slot.id);
                }}
              >
                <div style={{ fontWeight: 600, fontSize: 15 }}>
                  📍 {slot.date} {getDayOfWeek(slot.date)}
                </div>
                <div style={{ color: '#666', fontSize: 14 }}>
                  🕐 {slot.start_time} - {slot.end_time}
                </div>
                {!isSelf && bookingId === slot.id && (
                  <div style={{ marginTop: 12 }}>
                    <textarea
                      placeholder="附言：简单介绍一下你想练习的方向..."
                      value={message}
                      onChange={(e) => setMessage(e.target.value)}
                      style={textareaStyle}
                      rows={3}
                    />
                    <div style={{ marginTop: 8, display: 'flex', gap: 8, alignItems: 'center' }}>
                      <button onClick={handleBook} style={btnPrimary}>发起预约</button>
                      <button onClick={() => setBookingId(null)} style={btnCancel}>取消</button>
                      {bookingError && <span style={{ color: '#f5222d', fontSize: 13 }}>{bookingError}</span>}
                      {bookingSuccess && <span style={{ color: '#52c41a', fontSize: 13 }}>{bookingSuccess}</span>}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
        {isSelf && (
          <div style={{ marginTop: 12 }}>
            <button onClick={() => navigate('/settings/availability')} style={btnSecondary}>
              ⚙️ 管理我的空闲时间
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

// ── 样式 ──

const container: React.CSSProperties = {
  padding: 40,
  maxWidth: 800,
  margin: '0 auto',
  fontFamily: 'system-ui',
};

const profileCard: React.CSSProperties = {
  background: '#fff',
  border: '1px solid #e8e8e8',
  borderRadius: 12,
  padding: 28,
  marginBottom: 24,
};

const sectionStyle: React.CSSProperties = {
  background: '#fff',
  border: '1px solid #e8e8e8',
  borderRadius: 12,
  padding: 28,
};

const slotCardStyle: React.CSSProperties = {
  border: '1px solid #e8e8e8',
  borderRadius: 8,
  padding: 14,
  cursor: 'pointer',
  transition: 'all 0.2s',
};

const tagStyle: React.CSSProperties = {
  padding: '4px 14px',
  borderRadius: 14,
  background: '#e6f7ff',
  color: '#1677ff',
  fontSize: 13,
  fontWeight: 500,
};

const contactStyle: React.CSSProperties = {
  marginTop: 12,
  padding: '8px 14px',
  background: '#f6ffed',
  border: '1px solid #b7eb8f',
  borderRadius: 8,
  fontSize: 14,
  color: '#389e0d',
};

const emptyStyle: React.CSSProperties = {
  textAlign: 'center',
  padding: 40,
  color: '#999',
};

const textareaStyle: React.CSSProperties = {
  width: '100%',
  padding: 10,
  borderRadius: 6,
  border: '1px solid #d9d9d9',
  fontFamily: 'system-ui',
  fontSize: 14,
  resize: 'vertical',
  boxSizing: 'border-box',
};

const btnPrimary: React.CSSProperties = {
  padding: '8px 20px',
  borderRadius: 6,
  border: 'none',
  background: '#1677ff',
  color: '#fff',
  fontSize: 14,
  cursor: 'pointer',
};

const btnCancel: React.CSSProperties = {
  padding: '8px 20px',
  borderRadius: 6,
  border: '1px solid #d9d9d9',
  background: '#fff',
  fontSize: 14,
  cursor: 'pointer',
  color: '#666',
};

const btnSecondary: React.CSSProperties = {
  padding: '8px 20px',
  borderRadius: 6,
  border: '1px solid #1677ff',
  background: '#fff',
  color: '#1677ff',
  fontSize: 14,
  cursor: 'pointer',
};

export default UserDetail;
