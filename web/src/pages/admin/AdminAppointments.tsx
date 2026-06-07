import { useEffect, useState } from 'react';
import { apiGet, apiDelete } from '../../api/client';

interface AppointmentData {
  id: string;
  mentor_id: string;
  mentor_nickname: string;
  student_id: string;
  student_nickname: string;
  time_slot_date: string;
  time_slot_start: string;
  time_slot_end: string;
  message: string;
  status: string;
  reject_reason: string;
  created_at: string;
}

interface ApptListResponse {
  appointments: AppointmentData[];
  total: number;
  page: number;
  page_size: number;
}

const PAGE_SIZE = 20;
const STATUS_MAP: Record<string, { label: string; color: string; bg: string; border: string }> = {
  pending: { label: '待确认', color: 'var(--color-warning)', bg: 'rgba(224,184,104,0.1)', border: 'rgba(224,184,104,0.25)' },
  accepted: { label: '已接受', color: 'var(--color-success)', bg: 'rgba(132,195,149,0.1)', border: 'rgba(132,195,149,0.25)' },
  rejected: { label: '已拒绝', color: 'var(--color-danger)', bg: 'rgba(224,112,112,0.1)', border: 'rgba(224,112,112,0.25)' },
};

export default function AdminAppointments() {
  const [appts, setAppts] = useState<AppointmentData[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  const fetchAppts = (p: number) => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ page: String(p), page_size: String(PAGE_SIZE) });

    apiGet<ApptListResponse>(`/admin/appointments?${params}`)
      .then((data) => {
        setAppts(data.appointments);
        setTotal(data.total);
        setLoading(false);
      })
      .catch(() => {
        setError('加载预约列表失败');
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchAppts(page);
  }, [page]);

  const handleDelete = async (id: string) => {
    if (!confirm('确定取消此预约吗？')) return;
    try {
      await apiDelete(`/admin/appointments/${encodeURIComponent(id)}`);
      setAppts((prev) => prev.filter((a) => a.id !== id));
      setTotal((t) => t - 1);
      setActionMsg('预约已取消');
      setTimeout(() => setActionMsg(''), 3000);
    } catch {
      setActionMsg('取消失败');
      setTimeout(() => setActionMsg(''), 3000);
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div>
      <h1 className="text-text font-bold mb-6" style={{ fontSize: 26 }}>预约管理</h1>

      {actionMsg && (
        <div className="mb-4 px-3 py-2 text-text pixel-corners-sm" style={{ fontSize: 18, background: 'rgba(224,184,104,0.1)', border: '1px solid rgba(224,184,104,0.25)' }}>
          {actionMsg}
        </div>
      )}

      {error && (
        <div className="mb-5 px-3 py-2 text-danger pixel-corners-sm"
             style={{ fontSize: 17, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>
          {error}
        </div>
      )}

      {loading ? (
        <p className="text-text-muted" style={{ fontSize: 18 }}>加载中...</p>
      ) : appts.length === 0 ? (
        <p className="text-text-muted" style={{ fontSize: 18 }}>没有预约记录</p>
      ) : (
        <>
          <div className="space-y-3">
            {appts.map((a) => {
              const status = STATUS_MAP[a.status] || STATUS_MAP.pending;
              return (
                <div
                  key={a.id}
                  className="bg-card border border-border pixel-corners p-4"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      {/* 面试官 → 面试者 */}
                      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mb-2">
                        <span style={{ fontSize: 16 }}>
                          👨‍🏫 <span className="text-text font-bold">{a.mentor_nickname || a.mentor_id}</span>
                          <span className="text-text-muted ml-1" style={{ fontSize: 13 }}>({a.mentor_id})</span>
                        </span>
                        <span className="text-text-muted" style={{ fontSize: 14 }}>→</span>
                        <span style={{ fontSize: 16 }}>
                          🎓 <span className="text-text font-bold">{a.student_nickname || a.student_id}</span>
                          <span className="text-text-muted ml-1" style={{ fontSize: 13 }}>({a.student_id})</span>
                        </span>
                      </div>

                      {/* 时间 */}
                      {a.time_slot_date && (
                        <p className="text-text-secondary mb-2" style={{ fontSize: 15 }}>
                          📅 {a.time_slot_date} · {a.time_slot_start} ~ {a.time_slot_end}
                        </p>
                      )}

                      {a.message && (
                        <p className="text-text-secondary mb-2" style={{ fontSize: 15 }}>
                          💬 {a.message}
                        </p>
                      )}

                      {a.reject_reason && (
                        <p className="text-danger mb-2" style={{ fontSize: 15 }}>
                          ❌ 拒绝原因：{a.reject_reason}
                        </p>
                      )}

                      <div className="flex items-center gap-2">
                        <span className="px-2 py-0.5" style={{
                          fontSize: 13,
                          color: status.color,
                          background: status.bg,
                          border: `1px solid ${status.border}`,
                        }}>
                          {status.label}
                        </span>
                        <span className="text-text-muted" style={{ fontSize: 13 }}>
                          {new Date(a.created_at).toLocaleString('zh-CN')}
                        </span>
                      </div>
                    </div>

                    {a.status !== 'rejected' && (
                      <button
                        onClick={() => handleDelete(a.id)}
                        className="text-danger no-underline hover:underline flex-shrink-0"
                        style={{ fontSize: 15, background: 'none', border: 'none', cursor: 'pointer' }}
                      >
                        取消预约
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-6">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="pixel-btn"
                style={{ fontSize: 15, opacity: page <= 1 ? 0.4 : 1 }}
              >
                上一页
              </button>
              <span className="text-text-muted px-3" style={{ fontSize: 15 }}>
                {page} / {totalPages}
              </span>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                className="pixel-btn"
                style={{ fontSize: 15, opacity: page >= totalPages ? 0.4 : 1 }}
              >
                下一页
              </button>
            </div>
          )}

          <p className="text-text-muted mt-3" style={{ fontSize: 15 }}>
            共 {total} 条预约
          </p>
        </>
      )}
    </div>
  );
}
