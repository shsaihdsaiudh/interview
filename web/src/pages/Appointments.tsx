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

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4', '#ef4444', '#f97316'];

const statusConfig: Record<string, { bg: string; text: string; label: string }> = {
  pending: { bg: 'bg-amber-50 text-amber-700', text: 'text-amber-700', label: '待确认' },
  accepted: { bg: 'bg-emerald-50 text-emerald-700', text: 'text-emerald-700', label: '已接受' },
  rejected: { bg: 'bg-red-50 text-red-600', text: 'text-red-600', label: '已拒绝' },
};

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

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      <h1 className="text-2xl font-bold text-text mb-6">我的预约</h1>

      {/* Tab */}
      <div className="flex bg-gray-100 rounded-lg p-1 mb-6 w-fit">
        {(['received', 'sent'] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-1.5 rounded-md text-sm font-medium transition cursor-pointer border-none ${
              tab === t
                ? 'bg-white text-text shadow-sm'
                : 'text-text-secondary hover:text-text bg-transparent'
            }`}
          >
            {t === 'received' ? '收到的预约' : '发出的预约'}
          </button>
        ))}
      </div>

      {actionError && (
        <div className="bg-red-50 border border-red-200 text-red-600 px-4 py-3 rounded-xl text-sm mb-4">
          {actionError}
        </div>
      )}

      {loading ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-32 rounded-2xl" />
          ))}
        </div>
      ) : error ? (
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">
          {error}
        </div>
      ) : appointments.length === 0 ? (
        <div className="text-center py-16 text-text-muted">
          <p className="text-lg font-medium">
            {tab === 'received' ? '还没有收到预约' : '还没有发出预约'}
          </p>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {appointments.map((a) => {
            const other = tab === 'received' ? a.student : a.mentor;
            const sc = statusConfig[a.status] || statusConfig.pending;
            const avatarBg = avatarColors[other.nickname.charCodeAt(0) % avatarColors.length];
            return (
              <div key={a.id} className="bg-card rounded-2xl border border-border shadow-sm p-5 hover:border-gray-300 transition">
                {/* 头部 */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    {other.avatar ? (
                      <img src={other.avatar} alt={other.nickname} className="w-10 h-10 rounded-full object-cover" />
                    ) : (
                      <div
                        className="w-10 h-10 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0"
                        style={{ background: avatarBg }}
                      >
                        {other.nickname.charAt(0)}
                      </div>
                    )}
                    <div>
                      <Link
                        to={`/user/${other.email}`}
                        className="font-semibold text-text hover:text-brand-600 transition-colors no-underline text-sm"
                      >
                        {other.nickname}
                      </Link>
                      <div className="text-xs text-text-muted">{other.department || '未设置院系'}</div>
                    </div>
                  </div>
                  <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${sc.bg}`}>
                    {sc.label}
                  </span>
                </div>

                {/* 时间 */}
                <div className="mt-3 px-3 py-2 rounded-lg bg-gray-50 text-sm text-text-secondary">
                  {a.time_slot?.date} · {a.time_slot?.start_time} - {a.time_slot?.end_time}
                </div>

                {/* 附言 */}
                {a.message && (
                  <div className="mt-2 text-sm text-text-secondary">{a.message}</div>
                )}

                {/* 拒绝原因 */}
                {a.reject_reason && (
                  <div className="mt-2 p-2.5 rounded-lg bg-red-50 text-sm text-red-600">
                    拒绝原因：{a.reject_reason}
                  </div>
                )}

                {/* 联系方式 */}
                {a.status === 'accepted' && other.contact_info && (
                  <div className="mt-2 p-3 rounded-lg bg-gray-50 border border-border text-sm text-text-secondary">
                    联系方式：{other.contact_info}
                  </div>
                )}

                {/* 操作 */}
                {tab === 'received' && a.status === 'pending' && (
                  <div className="flex gap-2 mt-3 pt-3 border-t border-border">
                    <button
                      onClick={() => handleAccept(a.id)}
                      className="px-4 py-1.5 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium
                                 transition cursor-pointer border-none"
                    >
                      接受
                    </button>
                    <button
                      onClick={() => handleReject(a.id)}
                      className="px-4 py-1.5 rounded-lg border border-border bg-white text-text-secondary text-sm
                                 font-medium hover:bg-red-50 hover:text-danger hover:border-red-200 transition cursor-pointer"
                    >
                      拒绝
                    </button>
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

export default Appointments;
