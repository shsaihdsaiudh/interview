import { useState, useMemo, useCallback } from 'react';

interface TimeSlot { id: string; user_id: string; date: string; start_time: string; end_time: string; }
interface DayInfo { date: string; label: string; dayOfWeek: string; isToday: boolean; }
interface TimeSlotInfo { label: string; hour: number; minute: number; }
interface BookingSlot { slotId: string; date: string; startTime: string; endTime: string; }

interface WeekCalendarProps {
  availabilities: TimeSlot[];
  isSelf: boolean;
  onBook: (slotId: string, message: string) => Promise<void>;
}

const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

function fmtDate(d: Date) { return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`; }
function t2m(t: string) { const [h,m]=t.split(':').map(Number); return h*60+m; }

function buildDays(): DayInfo[] {
  const days: DayInfo[] = [];
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const todayS = fmtDate(today);
  for (let i = 0; i < 7; i++) {
    const d = new Date(today); d.setDate(d.getDate() + i);
    const ds = fmtDate(d);
    days.push({ date: ds, label: `${d.getMonth()+1}/${d.getDate()}`, dayOfWeek: dayNames[d.getDay()], isToday: ds === todayS });
  }
  return days;
}

function buildTimeSlots(): TimeSlotInfo[] {
  const s: TimeSlotInfo[] = [];
  for (let h = 8; h < 22; h++)
    for (let m = 0; m < 60; m += 30)
      s.push({ label: `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}`, hour: h, minute: m });
  return s;
}

export default function WeekCalendar({ availabilities, isSelf, onBook }: WeekCalendarProps) {
  const days = useMemo(() => buildDays(), []);
  const timeSlots = useMemo(() => buildTimeSlots(), []);

  const availableCells = useMemo(() => {
    const map = new Map<string, Set<number>>();
    for (const a of availabilities) {
      if (!map.has(a.date)) map.set(a.date, new Set());
      const sm = t2m(a.start_time), em = t2m(a.end_time);
      const set = map.get(a.date)!;
      for (let m = sm; m < em; m += 30) {
        const idx = Math.floor((m - 480) / 30);
        if (idx >= 0 && idx < timeSlots.length) set.add(idx);
      }
    }
    return map;
  }, [availabilities, timeSlots.length]);

  const findAvailability = useCallback((day: string, idx: number): TimeSlot | undefined => {
    const sm = 480 + idx * 30, se = sm + 30;
    return availabilities.find(a => a.date === day && t2m(a.start_time) <= sm && t2m(a.end_time) >= se);
  }, [availabilities]);

  const [modalOpen, setModalOpen] = useState(false);
  const [sel, setSel] = useState<BookingSlot | null>(null);
  const [msg, setMsg] = useState('');
  const [bkErr, setBkErr] = useState('');
  const [bkOk, setBkOk] = useState('');
  const [bking, setBking] = useState(false);

  const click = (day: DayInfo, idx: number) => {
    if (isSelf || !availableCells.get(day.date)?.has(idx)) return;
    const avail = findAvailability(day.date, idx);
    if (!avail) return;
    const endLabel = idx + 1 < timeSlots.length ? timeSlots[idx + 1].label : '22:00';
    setSel({ slotId: avail.id, date: day.date, startTime: timeSlots[idx].label, endTime: endLabel });
    setMsg('希望预约一场模拟面试'); setBkErr(''); setBkOk(''); setModalOpen(true);
  };

  const confirm = async () => {
    if (!sel) return;
    setBking(true); setBkErr(''); setBkOk('');
    try {
      await onBook(sel.slotId, msg || '希望预约一场模拟面试');
      setBkOk('预约成功，请等待对方确认');
      setTimeout(() => { setModalOpen(false); setSel(null); setBkOk(''); }, 1500);
    } catch (err: unknown) {
      setBkErr((err as { response?: { data?: { error?: string } } })?.response?.data?.error || '预约失败');
    } finally { setBking(false); }
  };

  const close = () => { setModalOpen(false); setSel(null); setBkErr(''); setBkOk(''); };
  const dr = `${days[0].label} - ${days[days.length - 1].label}`;

  return (
    <>
      <div className="bg-card rounded-2xl border border-border shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-border-light flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1">
          <h2 className="text-lg font-bold text-text">空闲时间</h2>
          <span className="text-sm text-text-muted">{dr}</span>
        </div>

        <div className="overflow-auto max-h-[520px]">
          <div className="inline-block min-w-full">
            <table className="w-full border-collapse">
              <thead>
                <tr>
                  <th className="sticky top-0 left-0 z-20 bg-card border-b border-border-light w-16 min-w-[64px]" />
                  {days.map((day) => (
                    <th
                      key={day.date}
                      className={`sticky top-0 z-10 border-b border-border-light px-1 py-2 text-center min-w-[60px] sm:min-w-[80px] ${
                        day.isToday ? 'bg-brand-50/40' : 'bg-card'
                      }`}
                    >
                      <div className="text-xs font-semibold text-text-secondary">{day.dayOfWeek}</div>
                      <div className={`text-xs mt-0.5 ${day.isToday ? 'text-brand-600 font-bold' : 'text-text-muted'}`}>{day.label}</div>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {timeSlots.map((slot, si) => {
                  const isHour = slot.minute === 0;
                  return (
                    <tr key={slot.label}>
                      <td className={`sticky left-0 z-10 border-b border-border-light py-1 px-2 text-right text-[10px] sm:text-xs text-text-muted font-mono ${isHour ? 'bg-surface-alt/80' : 'bg-card'}`}>
                        {isHour ? slot.label : ''}
                      </td>
                      {days.map((day) => {
                        const isAvail = availableCells.get(day.date)?.has(si) ?? false;
                        const hourLine = slot.minute === 0;
                        return (
                          <td key={`${day.date}-${slot.label}`} className={`border-b border-border-light p-0 relative ${day.isToday ? 'bg-brand-50/20' : ''}`}>
                            <div
                              onClick={() => click(day, si)}
                              className={`h-7 sm:h-8 mx-0.5 my-px rounded-md transition-all duration-150 ${
                                isAvail
                                  ? isSelf ? 'bg-emerald-100/70 cursor-default' : 'bg-emerald-400 hover:bg-emerald-500 hover:scale-105 hover:shadow-sm cursor-pointer'
                                  : 'bg-transparent'
                              } ${!isAvail && hourLine ? 'border-t border-border-light/50' : ''}`}
                              title={isAvail && !isSelf ? `${day.date} ${slot.label}` : undefined}
                            />
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>

        {isSelf && (
          <div className="px-6 py-4 border-t border-border-light bg-surface-alt/50">
            <p className="text-sm text-text-muted text-center">
              前往 <a href="/appointments" className="text-brand-600 hover:text-brand-700 font-medium mx-1">预约管理</a> 设置你的空闲时间
            </p>
          </div>
        )}
      </div>

      {/* Booking Modal */}
      {modalOpen && sel && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          style={{ background: 'rgba(0,0,0,0.3)', backdropFilter: 'blur(6px)', WebkitBackdropFilter: 'blur(6px)' }}
          onClick={close}
        >
          <div
            className="bg-card rounded-2xl border border-border w-full max-w-sm overflow-hidden"
            style={{ boxShadow: '0 20px 60px rgba(0,0,0,0.12), 0 4px 16px rgba(0,0,0,0.06)' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="px-6 pt-6 pb-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-lg font-bold text-text">发起预约</h3>
                <button onClick={close} className="w-8 h-8 rounded-full flex items-center justify-center text-text-muted hover:text-text hover:bg-surface-alt transition cursor-pointer border-none bg-transparent">
                  ✕
                </button>
              </div>
              <div className="flex items-center gap-3 p-3 rounded-xl bg-emerald-50 border border-emerald-100">
                <div className="w-2 h-2 rounded-full bg-emerald-500 flex-shrink-0" />
                <div>
                  <div className="text-sm font-semibold text-text">
                    {sel.date}
                    <span className="text-text-muted font-normal ml-1">{dayNames[new Date(sel.date).getDay()]}</span>
                  </div>
                  <div className="text-sm text-text-secondary mt-0.5">{sel.startTime} - {sel.endTime}</div>
                </div>
              </div>
            </div>
            <div className="px-6 pb-4">
              <label className="flex flex-col gap-1.5">
                <span className="text-xs font-medium text-text-secondary">附言</span>
                <textarea
                  value={msg}
                  onChange={(e) => setMsg(e.target.value)}
                  placeholder="简单介绍一下你想练习的方向..."
                  className="w-full px-3 py-2.5 rounded-xl border border-border text-sm resize-y min-h-[80px] focus:outline-none focus:border-brand-400"
                  rows={3}
                />
              </label>
              {bkErr && (
                <div className="mt-3 px-3 py-2 rounded-lg bg-red-50 border border-red-100 text-sm text-danger">{bkErr}</div>
              )}
              {bkOk && (
                <div className="mt-3 px-3 py-2 rounded-lg bg-emerald-50 border border-emerald-100 text-sm text-success font-medium">{bkOk}</div>
              )}
            </div>
            <div className="px-6 pb-6 flex gap-3">
              <button onClick={close} className="flex-1 px-4 py-2.5 rounded-xl border border-border text-text-secondary text-sm font-medium hover:bg-surface-alt transition cursor-pointer bg-transparent">取消</button>
              <button onClick={confirm} disabled={bking || !!bkOk} className="flex-1 px-4 py-2.5 rounded-xl bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50">
                {bking ? '预约中...' : '确认预约'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
