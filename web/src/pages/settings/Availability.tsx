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

const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
const getDayOfWeek = (dateStr: string) => dayNames[new Date(dateStr).getDay()];

function AvailabilitySettings() {
  const [slots, setSlots] = useState<TimeSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  const [date, setDate] = useState('');
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');

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

  useEffect(() => { fetchData(); }, []);

  const showMsg = (msg: string) => {
    setActionMsg(msg);
    setTimeout(() => setActionMsg(''), 3000);
  };

  const handleAddSlot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!date || !startTime || !endTime) return;
    try {
      await apiPost('/availability', { date, start_time: startTime, end_time: endTime });
      setDate(''); setStartTime(''); setEndTime('');
      showMsg('已添加');
      fetchData();
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '添加失败');
    }
  };

  const handleDeleteSlot = async (id: string) => {
    try {
      await apiDelete(`/availability/${id}`);
      showMsg('已删除');
      fetchData();
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '删除失败');
    }
  };

  const handleSaveProfile = async () => {
    const tags = tagsStr.split(/[,，]/).map((t) => t.trim()).filter(Boolean);
    try {
      await apiPut('/profile', { nickname, student_id: studentId, department, tags, avatar, contact_info: contactInfo });
      const u = getUser();
      if (u) { setUser({ ...u, nickname }); notifyAuthChange(); }
      showMsg('已保存');
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '保存失败');
    }
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto px-6 py-10">
        <div className="skeleton h-96 rounded-2xl" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-2xl mx-auto px-6 py-10">
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">{error}</div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto px-6 py-10">
      <h1 className="text-2xl font-bold text-text mb-8">设置</h1>

      {actionMsg && (
        <div className={`px-4 py-2.5 rounded-xl text-sm mb-4 border ${
          actionMsg === '已保存' || actionMsg === '已添加' || actionMsg === '已删除'
            ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
            : 'bg-red-50 border-red-200 text-red-600'
        }`}>
          {actionMsg}
        </div>
      )}

      {/* 个人资料 */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8 mb-4">
        <h2 className="text-lg font-bold text-text mb-5">个人资料</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">昵称</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={nickname} onChange={(e) => setNickname(e.target.value)} />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">学号</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={studentId} onChange={(e) => setStudentId(e.target.value)} />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">院系</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={department} onChange={(e) => setDepartment(e.target.value)} placeholder="如：计算机学院" />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">面试方向标签</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={tagsStr} onChange={(e) => setTagsStr(e.target.value)} placeholder="产品, 前端, 后端" />
          </label>
          <label className="flex flex-col gap-1.5 sm:col-span-2">
            <span className="text-sm font-medium text-text-secondary">头像 URL</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={avatar} onChange={(e) => setAvatar(e.target.value)} placeholder="https://..." />
          </label>
          <label className="flex flex-col gap-1.5 sm:col-span-2">
            <span className="text-sm font-medium text-text-secondary">联系方式</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={contactInfo} onChange={(e) => setContactInfo(e.target.value)} placeholder="微信 / QQ / 手机" />
          </label>
        </div>
        <button
          onClick={handleSaveProfile}
          className="px-5 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none"
        >
          保存资料
        </button>
      </div>

      {/* 空闲时间 */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
        <h2 className="text-lg font-bold text-text mb-4">空闲时间</h2>

        <form onSubmit={handleAddSlot} className="flex flex-wrap gap-2 items-end mb-5 p-3 rounded-xl bg-surface-alt border border-border">
          <label className="flex flex-col gap-1">
            <span className="text-xs text-text-muted">日期</span>
            <input type="date" className="px-2 py-1.5 rounded-md border border-border bg-white text-sm" value={date} onChange={(e) => setDate(e.target.value)} required />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-xs text-text-muted">开始</span>
            <input type="time" className="px-2 py-1.5 rounded-md border border-border bg-white text-sm" value={startTime} onChange={(e) => setStartTime(e.target.value)} required />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-xs text-text-muted">结束</span>
            <input type="time" className="px-2 py-1.5 rounded-md border border-border bg-white text-sm" value={endTime} onChange={(e) => setEndTime(e.target.value)} required />
          </label>
          <button type="submit" className="px-3 py-1.5 rounded-md bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none">
            添加
          </button>
        </form>

        {slots.length === 0 ? (
          <div className="text-center py-10 text-text-muted text-sm">还没有添加空闲时间</div>
        ) : (
          <div className="flex flex-col gap-1.5">
            {slots.map((s) => (
              <div key={s.id} className="flex items-center justify-between px-3 py-2.5 rounded-lg bg-surface-alt border border-border text-sm">
                <span className="text-text-secondary">
                  {s.date} {getDayOfWeek(s.date)} · {s.start_time} - {s.end_time}
                </span>
                <button
                  onClick={() => handleDeleteSlot(s.id)}
                  className="text-xs text-text-muted hover:text-danger transition cursor-pointer border-none bg-transparent"
                >
                  删除
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default AvailabilitySettings;
