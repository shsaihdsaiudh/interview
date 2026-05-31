import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { apiGet } from '../api/client';
import { getUser } from '../components/Navbar';

interface UserItem {
  email: string;
  nickname: string;
  student_id: string;
  department: string;
  tags: string[];
  avatar: string;
}

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

interface OutgoingAppointment {
  id: string;
  mentor_id: string;
  student_id: string;
  time_slot_id: string;
  message: string;
  status: string;
  reject_reason: string;
  created_at: string;
  mentor: UserInfo;
  time_slot: TimeSlot;
}

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4', '#ef4444', '#f97316'];
const statusConfig: Record<string, { bg: string; label: string }> = {
  pending: { bg: 'bg-amber-50 text-amber-700', label: '待确认' },
  accepted: { bg: 'bg-emerald-50 text-emerald-700', label: '已接受' },
  rejected: { bg: 'bg-red-50 text-red-600', label: '已拒绝' },
};

function FindPeople() {
  const currentUser = getUser();
  const [users, setUsers] = useState<UserItem[]>([]);
  const [outgoing, setOutgoing] = useState<OutgoingAppointment[]>([]);
  const [loading, setLoading] = useState(true);
  const [outgoingLoading, setOutgoingLoading] = useState(!!currentUser);
  const [error, setError] = useState('');

  useEffect(() => {
    apiGet<{ users: UserItem[] }>('/users')
      .then((res) => setUsers(res.users))
      .catch(() => setError('加载用户列表失败'))
      .finally(() => setLoading(false));
  }, []);

  const fetchOutgoing = useCallback(() => {
    if (!currentUser) return;
    setOutgoingLoading(true);
    apiGet<{ appointments: OutgoingAppointment[] }>('/appointments?role=student')
      .then((res) => setOutgoing(res.appointments))
      .catch(() => {})
      .finally(() => setOutgoingLoading(false));
  }, [currentUser]);

  useEffect(() => {
    if (!currentUser) return;
    fetchOutgoing();
    const timer = setInterval(fetchOutgoing, 5000);
    return () => clearInterval(timer);
  }, [currentUser?.email]);

  if (loading) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-10">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => <div key={i} className="skeleton h-36" />)}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-10">
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">{error}</div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-6 py-10">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-text">找人预约</h1>
        <p className="text-text-secondary text-sm mt-1">浏览可预约的同学，点击进入主页选择时段发起预约</p>
      </div>

      {users.length === 0 ? (
        <div className="text-center py-20 text-text-muted">
          <p className="text-lg font-medium">还没有已验证的用户</p>
          <p className="text-sm mt-1">快去邀请同学注册吧</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {users.map((u) => {
            const bg = avatarColors[u.nickname.charCodeAt(0) % avatarColors.length];
            return (
              <Link
                to={`/user/${u.email}`}
                key={u.email}
                className="group bg-card rounded-2xl border border-border p-5 no-underline text-inherit hover:border-brand-200 hover:shadow-sm transition"
              >
                <div className="flex items-center gap-3">
                  {u.avatar ? (
                    <img src={u.avatar} alt={u.nickname} className="w-12 h-12 rounded-full object-cover" />
                  ) : (
                    <div className="w-12 h-12 rounded-full flex items-center justify-center text-white text-lg font-bold flex-shrink-0" style={{ background: bg }}>
                      {u.nickname.charAt(0)}
                    </div>
                  )}
                  <div className="min-w-0">
                    <div className="font-semibold text-text group-hover:text-brand-600 transition-colors truncate">{u.nickname}</div>
                    <div className="text-sm text-text-muted truncate mt-0.5">{u.department || '未设置院系'}</div>
                  </div>
                </div>
                {u.tags && u.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1.5 mt-3 pt-3 border-t border-border">
                    {u.tags.map((t) => (
                      <span key={t} className="px-2 py-0.5 rounded-md bg-brand-50 text-brand-700 text-xs font-medium">{t}</span>
                    ))}
                  </div>
                )}
              </Link>
            );
          })}
        </div>
      )}

      {/* ── 我发出的预约 ── */}
      {currentUser && (
        <div className="mt-12">
          <h2 className="text-xl font-bold text-text mb-4">我发出的预约</h2>
          {outgoingLoading ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 2 }).map((_, i) => <div key={i} className="skeleton h-24 rounded-xl" />)}
            </div>
          ) : outgoing.length === 0 ? (
            <div className="text-center py-10 text-text-muted bg-card rounded-2xl border border-border">
              <p className="text-sm">还没有发出预约</p>
              <p className="text-xs mt-1">在上方浏览同学，进入主页选择时段发起预约</p>
            </div>
          ) : (
            <div className="flex flex-col gap-2">
              {outgoing.map((a) => {
                const sc = statusConfig[a.status] || statusConfig.pending;
                const avatarBg = avatarColors[a.mentor.nickname.charCodeAt(0) % avatarColors.length];
                return (
                  <div key={a.id} className="bg-card rounded-2xl border border-border shadow-sm p-4 hover:border-gray-300 transition">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        {a.mentor.avatar ? (
                          <img src={a.mentor.avatar} alt={a.mentor.nickname} className="w-9 h-9 rounded-full object-cover" />
                        ) : (
                          <div className="w-9 h-9 rounded-full flex items-center justify-center text-white text-xs font-bold flex-shrink-0" style={{ background: avatarBg }}>
                            {a.mentor.nickname.charAt(0)}
                          </div>
                        )}
                        <div>
                          <Link to={`/user/${a.mentor.email}`} className="font-semibold text-text hover:text-brand-600 transition-colors no-underline text-sm">
                            {a.mentor.nickname}
                          </Link>
                          <div className="text-xs text-text-muted">{a.mentor.department || '未设置院系'}</div>
                        </div>
                      </div>
                      <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${sc.bg}`}>{sc.label}</span>
                    </div>
                    <div className="mt-2 px-3 py-1.5 rounded-lg bg-gray-50 text-sm text-text-secondary">
                      {a.time_slot?.date} · {a.time_slot?.start_time} - {a.time_slot?.end_time}
                    </div>
                    {a.message && <div className="mt-1.5 text-sm text-text-secondary">{a.message}</div>}
                    {a.reject_reason && (
                      <div className="mt-2 p-2 rounded-lg bg-red-50 text-sm text-red-600">拒绝原因：{a.reject_reason}</div>
                    )}
                    {a.status === 'accepted' && a.mentor.contact_info && (
                      <div className="mt-2 p-2 rounded-lg bg-gray-50 border border-border text-sm text-text-secondary">联系方式：{a.mentor.contact_info}</div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default FindPeople;
