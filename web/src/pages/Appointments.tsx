import { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { apiGet, apiPut, apiPost, apiDelete, getApiErrorMessage } from '../api/client';

interface UserInfo { email: string; nickname: string; avatar: string; department: string; student_id: string; contact_info: string; account_status: string; }
interface TimeSlot { id: string; user_id: string; date: string; start_time: string; end_time: string; }
interface AppointmentItem { id: string; mentor_id: string; student_id: string; time_slot_id: string; message: string; status: string; reject_reason: string; created_at: string; mentor: UserInfo; student: UserInfo; time_slot: TimeSlot; }

const avatarColors = ['#e0b868','#78b880','#d4a040','#e07070','#aaaab2','#72727c','#c4a060','#b0b0b8'];
const dayNames = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat'];
const getDay = (d: string) => dayNames[new Date(d).getDay()];

const statusCfg: Record<string, { cls: string; label: string }> = {
  pending: { cls: 'text-warning', label: '待确认' },
  accepted: { cls: 'text-success', label: '已接受' },
  rejected: { cls: 'text-danger', label: '已拒绝' },
};
type Tab = '预约' | 'received';

function Appointments() {
  const [tab, setTab] = useState<Tab>('预约');
  const [appointments, setAppointments] = useState<AppointmentItem[]>([]);
  const [slots, setSlots] = useState<TimeSlot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [msg, setMsg] = useState('');
  const [date, setDate] = useState(''); const [start, setStart] = useState(''); const [end, setEnd] = useState('');

  const show = (m: string) => { setMsg(m); setTimeout(() => setMsg(''), 3000); };
  const fetchSlots = useCallback(() => apiGet<{ availabilities: TimeSlot[] }>('/availability').then((r) => { setSlots(r.availabilities); setError(''); }).catch(() => setError('加载失败')), []);
  const fetchApps = useCallback(() => apiGet<{ appointments: AppointmentItem[] }>('/appointments?role=mentor').then((r) => { setAppointments(r.appointments); setError(''); }).catch(() => setError('加载失败')), []);

  useEffect(() => { setLoading(true); if (tab === '预约') { fetchSlots().finally(() => setLoading(false)); } else { fetchApps().finally(() => setLoading(false)); const t = setInterval(fetchApps, 5000); return () => clearInterval(t); } }, [tab, fetchSlots, fetchApps]);

  const add = async (e: React.FormEvent) => { e.preventDefault(); if (!date || !start || !end) return; try { await apiPost('/availability', { date, start_time: start, end_time: end }); setDate(''); setStart(''); setEnd(''); show('成功'); fetchSlots(); } catch (err: unknown) { show(getApiErrorMessage(err, '失败')); } };
  const del = async (id: string) => { try { await apiDelete(`/availability/${id}`); show('已删除'); fetchSlots(); } catch (err: unknown) { show(getApiErrorMessage(err, '失败')); } };
  const accept = async (id: string) => { try { await apiPut(`/appointments/${id}/accept`); fetchApps(); } catch (err: unknown) { show(getApiErrorMessage(err, '失败')); } };
  const reject = async (id: string) => { const r = prompt('reason:'); try { await apiPut(`/appointments/${id}/reject`, { reason: r || '' }); fetchApps(); } catch (err: unknown) { show(getApiErrorMessage(err, '失败')); } };

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      <h1 className="text-text mb-1" style={{ fontSize: 24, fontWeight: 700 }}>my slots</h1>
      <p className="text-text-secondary mb-6" style={{ fontSize: 18 }}>发布可预约时间，管理预约请求</p>

      <div className="flex border-b border-border mb-6">
        {(['预约','received'] as Tab[]).map((t) => (
          <button key={t} onClick={() => setTab(t)}
            className={`px-4 py-2 cursor-pointer border-none bg-transparent transition-colors ${tab === t ? 'text-brand-600' : 'text-text-muted hover:text-text-secondary'}`}
            style={{ fontSize: 18, borderBottom: tab === t ? '2px solid var(--color-brand-600)' : '2px solid transparent' }}>
            [{t}]
          </button>
        ))}
      </div>

      {msg && (
        <div className={`px-3 py-2 mb-4 pixel-corners-sm ${msg.includes('失败') ? 'text-danger' : 'text-success'}`}
             style={{ fontSize: 18, background: msg.includes('失败') ? 'rgba(224,112,112,0.08)' : 'rgba(120,184,128,0.08)', border: `1px solid ${msg.includes('失败') ? 'rgba(224,112,112,0.2)' : 'rgba(120,184,128,0.2)'}` }}>
          {msg}
        </div>
      )}

      {tab === '预约' && (
        <div>
          <div className="bg-card border border-border pixel-corners p-5 mb-4">
            <h2 className="text-text mb-2" style={{ fontSize: 18, fontWeight: 700 }}>+ new slot</h2>
            <p className="text-text-muted mb-4" style={{ fontSize: 17 }}>发布后其他同学可以看到并预约</p>
            <form onSubmit={add} className="flex flex-wrap gap-3 items-end p-4 bg-surface-alt border border-border pixel-corners-sm">
              <label className="flex flex-col gap-1">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>DATE</span>
                <input type="date" className="px-2 py-1.5 bg-card border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={date} onChange={(e) => setDate(e.target.value)} required />
              </label>
              <label className="flex flex-col gap-1">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>START</span>
                <input type="time" className="px-2 py-1.5 bg-card border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={start} onChange={(e) => setStart(e.target.value)} required />
              </label>
              <label className="flex flex-col gap-1">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>END</span>
                <input type="time" className="px-2 py-1.5 bg-card border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={end} onChange={(e) => setEnd(e.target.value)} required />
              </label>
              <button type="submit" className="pixel-btn primary" style={{ fontSize: 18 }}>发布</button>
            </form>
          </div>

          {loading ? <div className="flex flex-col gap-2">{Array.from({length:3}).map((_,i) => <div key={i} className="skeleton h-12 pixel-corners" />)}</div>
            : slots.length === 0 ? (
              <div className="text-center py-16 text-text-muted">
                <p className="text-text" style={{ fontSize: 18, fontWeight: 700 }}>还没有发布时段</p>
                <p style={{ fontSize: 18, marginTop: 4 }}>发布后其他同学可以看到并预约</p>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                {slots.map((s) => (
                  <div key={s.id} className="flex items-center justify-between px-4 py-3 bg-card border border-border pixel-corners-sm">
                    <span className="text-text-secondary" style={{ fontSize: 17 }}>{s.date} {getDay(s.date)} · {s.start_time}-{s.end_time}</span>
                    <button onClick={() => del(s.id)} className="cursor-pointer border-none bg-transparent text-text-muted hover:text-danger transition-colors" style={{ fontSize: 17 }}>删除</button>
                  </div>
                ))}
              </div>
            )}
        </div>
      )}

      {tab === 'received' && (
        <>
          {loading ? <div className="flex flex-col gap-3">{Array.from({length:3}).map((_,i) => <div key={i} className="skeleton h-32 pixel-corners" />)}</div>
            : error ? <div className="p-8 text-center text-danger pixel-corners" style={{ background: 'rgba(224,112,112,0.06)', border: '1px solid rgba(224,112,112,0.2)' }}>{error}</div>
            : appointments.length === 0 ? (
              <div className="text-center py-16 text-text-muted">
                <p className="text-text" style={{ fontSize: 18, fontWeight: 700 }}>还没有收到预约</p>
                <p style={{ fontSize: 18, marginTop: 4 }}>发布时段后其他同学就可以预约你</p>
              </div>
            ) : (
              <div className="flex flex-col gap-3">
                {appointments.map((a) => {
                  const sc = statusCfg[a.status] || statusCfg.pending;
                  const abg = avatarColors[a.student.nickname.charCodeAt(0) % avatarColors.length];
                  return (
                    <div key={a.id} className="bg-card border border-border pixel-corners p-5 card-hover">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                          {a.student.avatar ? <img src={a.student.avatar} alt="" className="w-10 h-10 object-cover" />
                            : <div className="w-10 h-10 flex items-center justify-center text-white font-bold" style={{ background: abg, fontSize: 18 }}>{a.student.nickname.charAt(0)}</div>}
                          <div>
                            <Link to={`/user/${a.student.email}`} className="no-underline text-text hover:text-brand-600 transition-colors" style={{ fontSize: 17, fontWeight: 700 }}>{a.student.nickname}</Link>
                            <div className="text-text-muted" style={{ fontSize: 17 }}>{a.student.department || '--'}</div>
                          </div>
                        </div>
                        <span className={sc.cls} style={{ fontSize: 17 }}>[{sc.label}]</span>
                      </div>
                      <div className="mt-3 px-3 py-2 bg-surface-alt text-text-secondary pixel-corners-sm" style={{ fontSize: 18 }}>
                        {a.time_slot?.date} · {a.time_slot?.start_time}-{a.time_slot?.end_time}
                      </div>
                      {a.message && <div className="mt-2 text-text-secondary" style={{ fontSize: 18 }}>{a.message}</div>}
                      {a.reject_reason && <div className="mt-2 px-3 py-2 text-danger pixel-corners-sm" style={{ fontSize: 18, background: 'rgba(224,112,112,0.06)' }}>原因：{a.reject_reason}</div>}
                      {a.status === '已接受' && a.student.contact_info && (
                        <div className="mt-2 px-3 py-2 bg-surface-alt border border-border pixel-corners-sm text-text-secondary" style={{ fontSize: 18 }}>contact: {a.student.contact_info}</div>
                      )}
                      {a.status === '待确认' && (
                        <div className="flex gap-2 mt-3 pt-3 border-t border-border">
                          <button onClick={() => accept(a.id)} className="pixel-btn primary" style={{ fontSize: 18 }}>接受</button>
                          <button onClick={() => reject(a.id)} className="pixel-btn" style={{ fontSize: 18 }}>拒绝</button>
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
