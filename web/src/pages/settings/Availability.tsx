import { useEffect, useState } from 'react';
import { apiGet, apiPost, apiDelete, apiPut } from '../../api/client';
import { getUser, setUser, notifyAuthChange } from '../../components/Navbar';

interface TimeSlot {
  id: string;
  user_id: string;
  date: string;
  start_time: string;
  end_time: string;
}

interface ProfileData {
  user: {
    email: string;
    nickname: string;
    student_id: string;
    department: string;
    tags: string[];
    avatar: string;
    contact_info: string;
    email_verified: boolean;
  };
  availabilities: TimeSlot[];
}

function AvailabilitySettings() {
  const [slots, setSlots] = useState<TimeSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  // 表单：添加空闲时间
  const [date, setDate] = useState('');
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');

  // 表单：编辑资料
  const [nickname, setNickname] = useState('');
  const [department, setDepartment] = useState('');
  const [tagsStr, setTagsStr] = useState('');
  const [avatar, setAvatar] = useState('');
  const [contactInfo, setContactInfo] = useState('');
  const [studentId, setStudentId] = useState('');

  const fetchData = () => {
    apiGet<ProfileData>('/profile')
      .then((res) => {
        setSlots(res.availabilities);
        setNickname(res.user.nickname);
        setStudentId(res.user.student_id);
        setDepartment(res.user.department || '');
        setTagsStr((res.user.tags || []).join(', '));
        setAvatar(res.user.avatar || '');
        setContactInfo(res.user.contact_info || '');
      })
      .catch(() => setError('加载资料失败'))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchData();
  }, []);

  const showMsg = (msg: string) => {
    setActionMsg(msg);
    setTimeout(() => setActionMsg(''), 3000);
  };

  // 添加空闲时间
  const handleAddSlot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!date || !startTime || !endTime) return;
    try {
      await apiPost('/availability', { date, start_time: startTime, end_time: endTime });
      setDate('');
      setStartTime('');
      setEndTime('');
      showMsg('✅ 空闲时间已添加');
      fetchData();
    } catch (err: any) {
      showMsg('❌ ' + (err?.response?.data?.error || '添加失败'));
    }
  };

  // 删除空闲时间
  const handleDeleteSlot = async (id: string) => {
    try {
      await apiDelete(`/availability/${id}`);
      showMsg('✅ 已删除');
      fetchData();
    } catch (err: any) {
      showMsg('❌ ' + (err?.response?.data?.error || '删除失败'));
    }
  };

  // 保存资料
  const handleSaveProfile = async () => {
    const tags = tagsStr
      .split(/[,，]/)
      .map((t) => t.trim())
      .filter(Boolean);
    try {
      await apiPut('/profile', {
        nickname,
        student_id: studentId,
        department,
        tags,
        avatar,
        contact_info: contactInfo,
      });
      // 更新本地存储的用户信息
      const u = getUser();
      if (u) {
        setUser({ ...u, nickname });
        notifyAuthChange();
      }
      showMsg('✅ 资料已保存');
    } catch (err: any) {
      showMsg('❌ ' + (err?.response?.data?.error || '保存失败'));
    }
  };

  const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
  const getDayOfWeek = (dateStr: string) => {
    const d = new Date(dateStr);
    return dayNames[d.getDay()];
  };

  if (loading) return <div style={container}><p>⏳ 加载中...</p></div>;
  if (error) return <div style={container}><p style={{ color: '#f5222d' }}>❌ {error}</p></div>;

  return (
    <div style={container}>
      <h1>⚙️ 设置</h1>

      {actionMsg && (
        <div style={{
          ...msgBanner,
          background: actionMsg.startsWith('✅') ? '#f6ffed' : '#fff2f0',
          borderColor: actionMsg.startsWith('✅') ? '#b7eb8f' : '#ffccc7',
        }}>
          {actionMsg}
        </div>
      )}

      {/* ── 个人资料编辑 ── */}
      <div style={sectionStyle}>
        <h2>👤 个人资料</h2>
        <div style={formGrid}>
          <label style={labelStyle}>
            昵称
            <input style={inputStyle} value={nickname} onChange={(e) => setNickname(e.target.value)} />
          </label>
          <label style={labelStyle}>
            学号
            <input style={inputStyle} value={studentId} onChange={(e) => setStudentId(e.target.value)} />
          </label>
          <label style={labelStyle}>
            院系
            <input style={inputStyle} value={department} onChange={(e) => setDepartment(e.target.value)} placeholder="如：计算机学院" />
          </label>
          <label style={labelStyle}>
            面试方向标签（逗号分隔）
            <input style={inputStyle} value={tagsStr} onChange={(e) => setTagsStr(e.target.value)} placeholder="如：产品, 前端, 后端" />
          </label>
          <label style={labelStyle}>
            头像 URL
            <input style={inputStyle} value={avatar} onChange={(e) => setAvatar(e.target.value)} placeholder="https://..." />
          </label>
          <label style={labelStyle}>
            联系方式
            <input style={inputStyle} value={contactInfo} onChange={(e) => setContactInfo(e.target.value)} placeholder="微信 / QQ / 手机" />
          </label>
        </div>
        <button onClick={handleSaveProfile} style={btnPrimary}>💾 保存资料</button>
      </div>

      {/* ── 空闲时间管理 ── */}
      <div style={sectionStyle}>
        <h2>📅 空闲时间</h2>

        {/* 添加表单 */}
        <form onSubmit={handleAddSlot} style={{ display: 'flex', gap: 10, flexWrap: 'wrap', alignItems: 'end', marginBottom: 20 }}>
          <label style={miniLabel}>
            日期
            <input type="date" style={inputStyle} value={date} onChange={(e) => setDate(e.target.value)} required />
          </label>
          <label style={miniLabel}>
            开始
            <input type="time" style={inputStyle} value={startTime} onChange={(e) => setStartTime(e.target.value)} required />
          </label>
          <label style={miniLabel}>
            结束
            <input type="time" style={inputStyle} value={endTime} onChange={(e) => setEndTime(e.target.value)} required />
          </label>
          <button type="submit" style={{ ...btnPrimary, height: 38 }}>➕ 添加</button>
        </form>

        {/* 已有时间列表 */}
        {slots.length === 0 ? (
          <div style={emptyStyle}>还没有添加空闲时间</div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {slots.map((s) => (
              <div key={s.id} style={slotItemStyle}>
                <span>
                  📍 {s.date} {getDayOfWeek(s.date)} · 🕐 {s.start_time} - {s.end_time}
                </span>
                <button onClick={() => handleDeleteSlot(s.id)} style={btnDelete}>删除</button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// ── 样式 ──

const container: React.CSSProperties = {
  padding: 40,
  maxWidth: 700,
  margin: '0 auto',
  fontFamily: 'system-ui',
};

const sectionStyle: React.CSSProperties = {
  background: '#fff',
  border: '1px solid #e8e8e8',
  borderRadius: 12,
  padding: 24,
  marginTop: 20,
};

const formGrid: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: '1fr 1fr',
  gap: '12px 20px',
  marginBottom: 16,
};

const labelStyle: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 4,
  fontSize: 14,
  color: '#333',
};

const miniLabel: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 4,
  fontSize: 14,
  color: '#333',
};

const inputStyle: React.CSSProperties = {
  padding: '8px 10px',
  borderRadius: 6,
  border: '1px solid #d9d9d9',
  fontSize: 14,
  fontFamily: 'system-ui',
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

const btnDelete: React.CSSProperties = {
  padding: '4px 14px',
  borderRadius: 6,
  border: '1px solid #ffccc7',
  background: '#fff',
  color: '#f5222d',
  fontSize: 13,
  cursor: 'pointer',
};

const slotItemStyle: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  padding: '10px 14px',
  background: '#fafafa',
  borderRadius: 8,
  fontSize: 14,
};

const emptyStyle: React.CSSProperties = {
  textAlign: 'center',
  padding: 30,
  color: '#999',
};

const msgBanner: React.CSSProperties = {
  padding: '10px 16px',
  borderRadius: 8,
  marginBottom: 16,
  border: '1px solid',
  fontSize: 14,
};

export default AvailabilitySettings;
