import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { apiGet, apiPut } from '../api/client';

interface UserInfo {
  email: string;
  nickname: string;
  avatar: string;
  department: string;
  student_id: string;
  contact_info: string;
}

interface TimeSlot {
  id: string;
  date: string;
  start_time: string;
  end_time: string;
}

interface AppointmentItem {
  id: string;
  mentor_id: string;
  student_id: string;
  time_slot_id: string;
  message: string;
  status: string;
  reject_reason: string;
  created_at: string;
  mentor: UserInfo;
  student: UserInfo;
  time_slot: TimeSlot;
}

function Appointments() {
  const [tab, setTab] = useState<'received' | 'sent'>('received');
  const [appointments, setAppointments] = useState<AppointmentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionError, setActionError] = useState('');

  const fetch = useCallback(() => {
    const role = tab === 'received' ? 'mentor' : 'student';
    apiGet<{ appointments: AppointmentItem[] }>(`/appointments?role=${role}`)
      .then((res) => {
        setAppointments(res.appointments);
        setError('');
      })
      .catch(() => setError('加载预约列表失败'))
      .finally(() => setLoading(false));
  }, [tab]);

  useEffect(() => {
    setLoading(true);
    fetch();
    // 每 5 秒自动刷新
    const timer = setInterval(fetch, 5000);
    return () => clearInterval(timer);
  }, [fetch]);

  const handleAccept = async (id: string) => {
    setActionError('');
    try {
      await apiPut(`/appointments/${id}/accept`);
      fetch();
    } catch (err: any) {
      setActionError(err?.response?.data?.error || '操作失败');
    }
  };

  const handleReject = async (id: string) => {
    const reason = prompt('拒绝原因（可选）：');
    setActionError('');
    try {
      await apiPut(`/appointments/${id}/reject`, { reason: reason || '' });
      fetch();
    } catch (err: any) {
      setActionError(err?.response?.data?.error || '操作失败');
    }
  };

  const statusBadge = (status: string): React.CSSProperties => {
    const colors: Record<string, { bg: string; color: string; text: string }> = {
      pending: { bg: '#fff7e6', color: '#fa8c16', text: '待确认' },
      accepted: { bg: '#f6ffed', color: '#52c41a', text: '已接受' },
      rejected: { bg: '#fff2f0', color: '#f5222d', text: '已拒绝' },
    };
    const c = colors[status] || colors.pending;
    return {
      ...badgeBase,
      background: c.bg,
      color: c.color,
    };
  };

  // ── 渲染 ──
  return (
    <div style={container}>
      <h1>📋 我的预约</h1>

      {/* Tab 切换 */}
      <div style={{ display: 'flex', gap: 0, marginBottom: 20 }}>
        <button
          onClick={() => setTab('received')}
          style={tab === 'received' ? tabActive : tabInactive}
        >
          📥 收到的预约
        </button>
        <button
          onClick={() => setTab('sent')}
          style={tab === 'sent' ? tabActive : tabInactive}
        >
          📤 发出的预约
        </button>
      </div>

      {actionError && (
        <div style={errorBanner}>⚠️ {actionError}</div>
      )}

      {loading ? (
        <p>⏳ 加载中...</p>
      ) : error ? (
        <p style={{ color: '#f5222d' }}>❌ {error}</p>
      ) : appointments.length === 0 ? (
        <div style={emptyStyle}>
          {tab === 'received' ? '📭 还没有收到预约' : '📭 还没有发出预约'}
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {appointments.map((a) => {
            const other = tab === 'received' ? a.student : a.mentor;
            return (
              <div key={a.id} style={cardStyle}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <Avatar name={other.nickname} url={other.avatar} />
                    <div>
                      <Link to={`/user/${other.email}`} style={{ fontWeight: 600, color: '#1677ff', textDecoration: 'none' }}>
                        {other.nickname}
                      </Link>
                      <div style={{ fontSize: 13, color: '#999' }}>{other.department || '未设置院系'}</div>
                    </div>
                  </div>
                  <span style={statusBadge(a.status)}>
                    {a.status === 'pending' ? '待确认' : a.status === 'accepted' ? '已接受' : '已拒绝'}
                  </span>
                </div>

                <div style={slotInfoStyle}>
                  📅 {a.time_slot?.date} · 🕐 {a.time_slot?.start_time} - {a.time_slot?.end_time}
                </div>

                {a.message && <div style={msgStyle}>💬 "{a.message}"</div>}
                {a.reject_reason && <div style={rejectStyle}>❌ 拒绝原因：{a.reject_reason}</div>}

                {a.status === 'accepted' && other.contact_info && (
                  <div style={contactStyle}>📞 联系方式：{other.contact_info}</div>
                )}

                {/* 收到的待确认预约：操作按钮 */}
                {tab === 'received' && a.status === 'pending' && (
                  <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
                    <button onClick={() => handleAccept(a.id)} style={btnAccept}>✅ 接受</button>
                    <button onClick={() => handleReject(a.id)} style={btnReject}>❌ 拒绝</button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ── 小组件：字母头像 ──
function Avatar({ name, url }: { name: string; url: string }) {
  const colors = ['#1677ff', '#52c41a', '#fa8c16', '#eb2f96', '#722ed1', '#13c2c2', '#f5222d', '#faad14'];
  const idx = name.charCodeAt(0) % colors.length;
  return (
    <div style={{
      width: 40, height: 40, borderRadius: '50%',
      background: url ? `url(${url}) center/cover` : colors[idx],
      color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center',
      fontSize: 16, fontWeight: 700, flexShrink: 0,
    }}>
      {url ? '' : name.charAt(0)}
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

const tabActive: React.CSSProperties = {
  padding: '10px 24px',
  border: 'none',
  background: '#1677ff',
  color: '#fff',
  fontSize: 15,
  cursor: 'pointer',
  borderRadius: 0,
};

const tabInactive: React.CSSProperties = {
  padding: '10px 24px',
  border: '1px solid #d9d9d9',
  background: '#fff',
  color: '#666',
  fontSize: 15,
  cursor: 'pointer',
  borderRadius: 0,
};

const cardStyle: React.CSSProperties = {
  background: '#fff',
  border: '1px solid #e8e8e8',
  borderRadius: 10,
  padding: 20,
};

const badgeBase: React.CSSProperties = {
  padding: '3px 12px',
  borderRadius: 12,
  fontSize: 12,
  fontWeight: 500,
};

const slotInfoStyle: React.CSSProperties = {
  marginTop: 10,
  padding: '8px 12px',
  background: '#fafafa',
  borderRadius: 6,
  fontSize: 14,
  color: '#555',
};

const msgStyle: React.CSSProperties = {
  marginTop: 8,
  color: '#666',
  fontSize: 14,
  fontStyle: 'italic',
};

const rejectStyle: React.CSSProperties = {
  marginTop: 8,
  padding: '6px 10px',
  background: '#fff2f0',
  borderRadius: 6,
  fontSize: 13,
  color: '#f5222d',
};

const contactStyle: React.CSSProperties = {
  marginTop: 10,
  padding: '8px 12px',
  background: '#f6ffed',
  border: '1px solid #b7eb8f',
  borderRadius: 6,
  fontSize: 14,
  color: '#389e0d',
};

const emptyStyle: React.CSSProperties = {
  textAlign: 'center',
  padding: 60,
  color: '#999',
  fontSize: 16,
};

const errorBanner: React.CSSProperties = {
  background: '#fff2f0',
  border: '1px solid #ffccc7',
  borderRadius: 8,
  padding: '8px 16px',
  marginBottom: 16,
  color: '#f5222d',
  fontSize: 14,
};

const btnAccept: React.CSSProperties = {
  padding: '6px 18px',
  borderRadius: 6,
  border: 'none',
  background: '#52c41a',
  color: '#fff',
  fontSize: 14,
  cursor: 'pointer',
};

const btnReject: React.CSSProperties = {
  padding: '6px 18px',
  borderRadius: 6,
  border: '1px solid #d9d9d9',
  background: '#fff',
  color: '#666',
  fontSize: 14,
  cursor: 'pointer',
};

export default Appointments;
