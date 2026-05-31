import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { apiGet, apiPut, apiPost, apiDelete } from '../api/client';

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
  user_id: string;
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

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4', '#ef4444', '#f97316'];
const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
const getDayOfWeek = (dateStr: string) => dayNames[new Date(dateStr).getDay()];

const statusConfig: Record<string, { bg: string; text: string; label: string }> = {
  pending: { bg: 'bg-amber-50 text-amber-700', text: 'text-amber-700', label: '待确认' },
  accepted: { bg: 'bg-emerald-50 text-emerald-700', text: 'text-emerald-700', label: '已接受' },
  rejected: { bg: 'bg-red-50 text-red-600', text: 'text-red-600', label: '已拒绝' },
};

type Tab = 'slots' | 'received';

function Appointments() {
  const [tab, setTab] = useState<Tab>('slots');
  const [appointments, setAppointments] = useState<AppointmentItem[]>([]);
  const [slots, setSlots] = useState<TimeSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  const [date, setDate] = useState('');
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');

  const showMsg = (msg: string) => {
    setActionMsg(msg);
    setTimeout(() => setActionMsg(''), 3000);
  };

  const fetchSlots = useCallback(() => {
    return apiGet<{ availabilities: TimeSlot[] }>('/availability')
      .then((res) => { setSlots(res.availabilities); setError(''); })
      .catch(() => setError('加载时段失败'));
  }, []);

  const fetchAppointments = useCallback(() => {
    return apiGet<{ appointments: AppointmentItem[] }>('/appointments?role=mentor')
      .then((res) => {
        setAppointments(res.appointments);
        setError('');
      })
      .catch(() => setError('加载预约列表失败'));
  }, []);

  useEffect(() => {
    setLoading(true);
    if (tab === 'slots') {
      fetchSlots().finally(() => setLoading(false));
    } else {
      fetchAppointments().finally(() => setLoading(false));
      const timer = setInterval(fetchAppointments, 5000);
      return () => clearInterval(timer);
    }
  }, [tab, fetchSlots, fetchAppointments]);

  const handleAddSlot = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!date || !startTime || !endTime) return;
    try {
      await apiPost('/availability', { date, start_time: startTime, end_time: endTime });
      setDate(''); setStartTime(''); setEndTime('');
      showMsg('已发布');
      fetchSlots();
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '发布失败');
    }
  };

  const handleDeleteSlot = async (id: string) => {
    try {
      await apiDelete(`/availability/${id}`);
      showMsg('已删除');
      fetchSlots();
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '删除失败');
    }
  };

  const handleAccept = async (id: string) => {
    try {
      await apiPut(`/appointments/${id}/accept`);
      fetchAppointments();
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '操作失败');
    }
  };

  const handleReject = async (id: string) => {
    const reason = prompt('拒绝原因（可选）：');
    try {
      await apiPut(`/appointments/${id}/reject`, { reason: reason || '' });
      fetchAppointments();
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '操作失败');
    }
  };

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      <h1 className="text-2xl font-bold text-text mb-2">我的面试间</h1>
      <p className="text-text-secondary text-sm mb-6">发布可预约时间，管理收到的预约请求</p>

      <div className="flex bg-gray-100 rounded-lg p-1 mb-6 w-fit">
        {([
          { key: 'slots' as Tab, label: '可预约时段' },
          { key: 'received' as Tab, label: '收到的预约' },
        ]).map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`px-4 py-1.5 rounded-md text-sm font-medium transition cursor-pointer border-none ${
              tab === t.key
                ? 'bg-white text-text shadow-sm'
                : 'text-text-secondary hover:text-text bg-transparent'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {actionMsg && (
        <div className={`px-4 py-2.5 rounded-xl text-sm mb-4 border ${
          actionMsg.includes('失败') ? 'bg-red-50 border-red-200 text-red-600' : 'bg-emerald-50 border-emerald-200 text-emerald-700'
        }`}>
          {actionMsg}
        </div>
      )}

      {tab === 'slots' && (
        <div>
          <div className="bg-card rounded-2xl border border-border shadow-sm p-6 mb-4">
            <h2 className="text-base font-bold text-text mb-3">发布可预约时间</h2>
            <p className="text-xs text-text-muted mb-4">发布后，其他同学可以在你的主页看到这些时段并向你发起预约</p>
            <form onSubmit={handleAddSlot} className="flex flex-wrap gap-2 items-end p-3 rounded-xl bg-surface-alt border border-border">
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
                发布
              </button>
            </form>
          </div>

          {loading ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-12 rounded-xl" />)}
            </div>
          ) : slots.length === 0 ? (
            <div className="text-center py-12 text-text-muted">
              <p className="text-lg font-medium">还没有发布可预约时间</p>
              <p className="text-sm mt-1">发布后其他同学可以看到并预约</p>
            </div>
          ) : (
            <div className="flex flex-col gap-1.5">
              {slots.map((s) => (
                <div key={s.id} className="flex items-center justify-between px-4 py-3 rounded-xl bg-card border border-border text-sm">
                  <span className="text-text-secondary">{s.date} {getDayOfWeek(s.date)} · {s.start_time} - {s.end_time}</span>
                  <button onClick={() => handleDeleteSlot(s.id)} className="text-xs text-text-muted hover:text-danger transition cursor-pointer border-none bg-transparent">删除</button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {tab === 'received' && (
        <>
          {loading ? (
            <div className="flex flex-col gap-3">
              {Array.from({ length: 3 }).map((_, i) => <div key={i} className="skeleton h-32 rounded-2xl" />)}
            </div>
          ) : error ? (
            <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">{error}</div>
          ) : appointments.length === 0 ? (
            <div className="text-center py-16 text-text-muted">
              <p className="text-lg font-medium">还没有收到预约</p>
              <p className="text-sm mt-1">发布可预约时段后，其他同学就可以预约你</p>
            </div>
          ) : (
            <div className="flex flex-col gap-3">
              {appointments.map((a) => {
                const sc = statusConfig[a.status] || statusConfig.pending;
                const avatarBg = avatarColors[a.student.nickname.charCodeAt(0) % avatarColors.length];
                return (
                  <div key={a.id} className="bg-card rounded-2xl border border-border shadow-sm p-5 hover:border-gray-300 transition">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        {a.student.avatar ? (
                          <img src={a.student.avatar} alt={a.student.nickname} className="w-10 h-10 rounded-full object-cover" />
                        ) : (
                          <div className="w-10 h-10 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0" style={{ background: avatarBg }}>
                            {a.student.nickname.charAt(0)}
                          </div>
                        )}
                        <div>
                          <Link to={`/user/${a.student.email}`} className="font-semibold text-text hover:text-brand-600 transition-colors no-underline text-sm">
                            {a.student.nickname}
                          </Link>
                          <div className="text-xs text-text-muted">{a.student.department || '未设置院系'}</div>
                        </div>
                      </div>
                      <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${sc.bg}`}>{sc.label}</span>
                    </div>
                    <div className="mt-3 px-3 py-2 rounded-lg bg-gray-50 text-sm text-text-secondary">
                      {a.time_slot?.date} · {a.time_slot?.start_time} - {a.time_slot?.end_time}
                    </div>
                    {a.message && <div className="mt-2 text-sm text-text-secondary">{a.message}</div>}
                    {a.reject_reason && <div className="mt-2 p-2.5 rounded-lg bg-red-50 text-sm text-red-600">拒绝原因：{a.reject_reason}</div>}
                    {a.status === 'accepted' && a.student.contact_info && (
                      <div className="mt-2 p-3 rounded-lg bg-gray-50 border border-border text-sm text-text-secondary">联系方式：{a.student.contact_info}</div>
                    )}
                    {a.status === 'pending' && (
                      <div className="flex gap-2 mt-3 pt-3 border-t border-border">
                        <button onClick={() => handleAccept(a.id)} className="px-4 py-1.5 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none">接受</button>
                        <button onClick={() => handleReject(a.id)} className="px-4 py-1.5 rounded-lg border border-border bg-white text-text-secondary text-sm font-medium hover:bg-red-50 hover:text-danger hover:border-red-200 transition cursor-pointer">拒绝</button>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default Appointments;
