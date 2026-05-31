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
  return dayNames[new Date(dateStr).getDay()];
}

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4'];

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
      setBookingSuccess('预约成功，请等待对方确认');
      setBookingId(null);
      setMessage('');
    } catch (err: any) {
      const msg = err?.response?.data?.error || '预约失败';
      setBookingError(msg);
    }
  };

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto px-6 py-10">
        <div className="skeleton h-48 rounded-2xl mb-4" />
        <div className="skeleton h-64 rounded-2xl" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-3xl mx-auto px-6 py-10">
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">
          {error}
        </div>
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="max-w-3xl mx-auto px-6 py-10 text-center text-text-muted">
        <p className="text-lg">用户不存在</p>
      </div>
    );
  }

  const { user, availabilities } = detail;
  const isSelf = currentUser?.email === user.email;
  const avatarBg = avatarColors[user.nickname.charCodeAt(0) % avatarColors.length];

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      {/* 用户信息 */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8 mb-4">
        <div className="flex items-start gap-4">
          {user.avatar ? (
            <img src={user.avatar} alt={user.nickname} className="w-16 h-16 rounded-full object-cover" />
          ) : (
            <div
              className="w-16 h-16 rounded-full flex items-center justify-center text-white text-2xl font-bold flex-shrink-0"
              style={{ background: avatarBg }}
            >
              {user.nickname.charAt(0)}
            </div>
          )}
          <div>
            <h1 className="text-2xl font-bold text-text">{user.nickname}</h1>
            <p className="text-text-secondary text-sm mt-1">
              {user.department || '未设置院系'} · {user.student_id}
            </p>
            {user.tags && user.tags.length > 0 && (
              <div className="flex flex-wrap gap-1.5 mt-2">
                {user.tags.map((t) => (
                  <span key={t} className="px-2 py-0.5 rounded-md bg-brand-50 text-brand-700 text-xs font-medium">
                    {t}
                  </span>
                ))}
              </div>
            )}
          </div>
        </div>

        {user.contact_info && (
          <div className="mt-4 p-3 rounded-xl bg-gray-50 border border-border text-sm text-text-secondary">
            联系方式：{user.contact_info}
          </div>
        )}
      </div>

      {/* 空闲时间 */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
        <h2 className="text-lg font-bold text-text mb-4">空闲时间</h2>

        {availabilities.length === 0 ? (
          <div className="text-center py-12 text-text-muted">
            <p className="font-medium">暂无空闲时间</p>
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {availabilities.map((slot) => {
              const selected = bookingId === slot.id;
              return (
                <div
                  key={slot.id}
                  onClick={() => { if (!isSelf) setBookingId(selected ? null : slot.id); }}
                  className={`rounded-xl p-4 border transition ${
                    selected
                      ? 'border-brand-400 bg-brand-50/50'
                      : 'border-border bg-gray-50/50 hover:border-gray-300'
                  } ${!isSelf ? 'cursor-pointer' : ''}`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-semibold text-text text-sm">
                        {slot.date} <span className="text-text-muted font-normal">{getDayOfWeek(slot.date)}</span>
                      </div>
                      <div className="text-sm text-text-secondary mt-0.5">
                        {slot.start_time} - {slot.end_time}
                      </div>
                    </div>
                    {!isSelf && !selected && (
                      <span className="text-xs text-text-muted">选择</span>
                    )}
                    {selected && (
                      <span className="text-xs text-brand-600 font-medium">已选中</span>
                    )}
                  </div>

                  {selected && (
                    <div className="mt-3 pt-3 border-t border-brand-200">
                      <textarea
                        placeholder="附言：简单介绍一下你想练习的方向..."
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        className="w-full px-3 py-2 rounded-lg border border-border text-sm resize-y min-h-[72px]"
                        rows={3}
                      />
                      <div className="flex items-center gap-2 mt-2">
                        <button
                          onClick={handleBook}
                          className="px-4 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium
                                     transition cursor-pointer border-none"
                        >
                          发起预约
                        </button>
                        <button
                          onClick={() => setBookingId(null)}
                          className="px-4 py-2 rounded-lg border border-border bg-white text-text-secondary text-sm
                                     font-medium hover:bg-gray-50 transition cursor-pointer"
                        >
                          取消
                        </button>
                        {bookingError && <span className="text-sm text-danger">{bookingError}</span>}
                        {bookingSuccess && <span className="text-sm text-success">{bookingSuccess}</span>}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}

        {isSelf && (
          <button
            onClick={() => navigate('/settings')}
            className="mt-4 w-full py-2.5 rounded-xl border border-dashed border-border text-text-secondary text-sm
                       font-medium hover:border-brand-300 hover:text-brand-600 transition cursor-pointer bg-transparent"
          >
            管理我的空闲时间
          </button>
        )}
      </div>
    </div>
  );
}

export default UserDetail;
