import { useEffect, useState } from 'react';
import { apiGet } from '../../api/client';

interface AdminStats {
  total_users: number;
  total_cards: number;
  total_appointments: number;
  new_users_today: number;
}

const STAT_CARDS = [
  { key: 'total_users', label: '总用户数', icon: '👥', color: 'var(--color-brand-600)' },
  { key: 'total_cards', label: '名片数量', icon: '🃏', color: 'var(--color-success)' },
  { key: 'total_appointments', label: '预约次数', icon: '📅', color: 'var(--color-info)' },
  { key: 'new_users_today', label: '今日新增', icon: '✨', color: 'var(--color-warning)' },
];

export default function AdminDashboard() {
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    apiGet<AdminStats>('/admin/stats')
      .then((data) => {
        setStats(data);
        setLoading(false);
      })
      .catch((err) => {
        setError((err as { response?: { data?: { error?: string } } })?.response?.data?.error || '加载失败');
        setLoading(false);
      });
  }, []);

  if (loading) {
    return <p className="text-text-muted" style={{ fontSize: 18 }}>加载中...</p>;
  }

  if (error) {
    return (
      <div className="px-3 py-2 text-danger" style={{ fontSize: 17, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>
        {error}
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-text font-bold mb-6" style={{ fontSize: 26 }}>仪表盘</h1>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-10">
        {STAT_CARDS.map((card) => (
          <div
            key={card.key}
            className="bg-card border border-border pixel-corners p-5"
          >
            <div className="flex items-center gap-2 mb-3">
              <span style={{ fontSize: 22 }}>{card.icon}</span>
              <span className="text-text-muted tracking-wider" style={{ fontSize: 14 }}>
                {card.label}
              </span>
            </div>
            <div className="font-bold" style={{ fontSize: 32, color: card.color }}>
              {stats ? stats[card.key as keyof AdminStats] : '-'}
            </div>
          </div>
        ))}
      </div>

      {/* 快捷入口 */}
      <h2 className="text-text font-bold mb-4" style={{ fontSize: 20 }}>快捷操作</h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        {[
          { label: '查看所有用户', desc: '用户列表，支持搜索、封禁', path: '/admin/users' },
          { label: '管理名片', desc: '查看和删除招募名片', path: '/admin/cards' },
          { label: '管理预约', desc: '查看和取消预约记录', path: '/admin/appointments' },
        ].map((item) => (
          <a
            key={item.path}
            href={item.path}
            className="no-underline text-inherit bg-card border border-border pixel-corners p-4 card-hover block"
          >
            <h3 className="font-bold text-text mb-1" style={{ fontSize: 17 }}>{item.label}</h3>
            <p className="text-text-muted" style={{ fontSize: 15 }}>{item.desc}</p>
          </a>
        ))}
      </div>
    </div>
  );
}
