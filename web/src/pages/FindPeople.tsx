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

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4', '#ef4444', '#f97316'];

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

  if (loading) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-10">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton h-36" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-10">
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-6 py-10">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-text">找人面试</h1>
        <p className="text-text-secondary text-sm mt-1">浏览可预约的面试官，找到和你方向匹配的人</p>
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
                className="group bg-card rounded-2xl border border-border p-5 no-underline text-inherit
                           hover:border-brand-200 hover:shadow-sm transition"
              >
                <div className="flex items-center gap-3">
                  {u.avatar ? (
                    <img src={u.avatar} alt={u.nickname} className="w-12 h-12 rounded-full object-cover" />
                  ) : (
                    <div
                      className="w-12 h-12 rounded-full flex items-center justify-center text-white text-lg font-bold flex-shrink-0"
                      style={{ background: bg }}
                    >
                      {u.nickname.charAt(0)}
                    </div>
                  )}
                  <div className="min-w-0">
                    <div className="font-semibold text-text group-hover:text-brand-600 transition-colors truncate">
                      {u.nickname}
                    </div>
                    <div className="text-sm text-text-muted truncate mt-0.5">
                      {u.department || '未设置院系'}
                    </div>
                  </div>
                </div>
                {u.tags && u.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1.5 mt-3 pt-3 border-t border-border">
                    {u.tags.map((t) => (
                      <span key={t} className="px-2 py-0.5 rounded-md bg-brand-50 text-brand-700 text-xs font-medium">
                        {t}
                      </span>
                    ))}
                  </div>
                )}
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default FindPeople;
