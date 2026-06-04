import { useState, useMemo, useCallback } from 'react';

interface TimeSlot { id: string; user_id: string; date: string; start_time: string; end_time: string; }
interface DayInfo { date: string; label: string; dayOfWeek: string; isToday: boolean; }
interface TimeSlotInfo { label: string; hour: number; minute: number; }
interface BookingSlot { slotId: string; date: string; startTime: string; endTime: string; }

interface WeekCalendarProps { availabilities: TimeSlot[]; isSelf: boolean; onBook: (slotId: string, message: string) => Promise<void>; }

const dayNames = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat'];
function fmtDate(d: Date) { return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`; }
function t2m(t: string) { const [h,m]=t.split(':').map(Number); return h*60+m; }
function buildDays(): DayInfo[] {
  const days: DayInfo[] = []; const now = new Date(); const today = new Date(now.getFullYear(), now.getMonth(), now.getDate()); const todayS = fmtDate(today);
  for (let i=0;i<7;i++) { const d = new Date(today); d.setDate(d.getDate()+i); const ds = fmtDate(d); days.push({ date: ds, label: `${d.getMonth()+1}/${d.getDate()}`, dayOfWeek: dayNames[d.getDay()], isToday: ds===todayS }); }
  return days;
}
function buildTimeSlots(): TimeSlotInfo[] { const s: TimeSlotInfo[] = []; for (let h=8;h<22;h++) for (let m=0;m<60;m+=30) s.push({ label: `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}`, hour: h, minute: m }); return s; }

export default function WeekCalendar({ availabilities, isSelf, onBook }: WeekCalendarProps) {
  const days = useMemo(() => buildDays(), []);
  const timeSlots = useMemo(() => buildTimeSlots(), []);
  const availableCells = useMemo(() => {
    const map = new Map<string, Set<number>>();
    for (const a of availabilities) { if (!map.has(a.date)) map.set(a.date, new Set()); const sm=t2m(a.start_time), em=t2m(a.end_time); const set=map.get(a.date)!; for (let m=sm;m<em;m+=30) { const idx=Math.floor((m-480)/30); if (idx>=0 && idx<timeSlots.length) set.add(idx); } }
    return map;
  }, [availabilities, timeSlots.length]);
  const findAvailability = useCallback((day: string, idx: number): TimeSlot|undefined => { const sm=480+idx*30, se=sm+30; return availabilities.find(a => a.date===day && t2m(a.start_time)<=sm && t2m(a.end_time)>=se); }, [availabilities]);
  const [modalOpen, setModalOpen] = useState(false); const [sel, setSel] = useState<BookingSlot|null>(null);
  const [msg, setMsg] = useState(''); const [bkErr, setBkErr] = useState(''); const [bkOk, setBkOk] = useState(''); const [bking, setBking] = useState(false);

  const click = (day: DayInfo, idx: number) => { if (isSelf||!availableCells.get(day.date)?.has(idx)) return; const avail=findAvailability(day.date, idx); if (!avail) return; setSel({ slotId: avail.id, date: day.date, startTime: timeSlots[idx].label, endTime: idx+1<timeSlots.length?timeSlots[idx+1].label:'22:00' }); setMsg('希望预约一场模拟面试'); setBkErr(''); setBkOk(''); setModalOpen(true); };
  const confirm = async () => { if (!sel) return; setBking(true); setBkErr(''); setBkOk(''); try { await onBook(sel.slotId, msg); setBkOk('成功'); setTimeout(() => { setModalOpen(false); setSel(null); setBkOk(''); }, 1500); } catch (err: unknown) { setBkErr((err as {response?:{data?:{error?:string}}})?.response?.data?.error||'失败'); } finally { setBking(false); } };
  const close = () => { setModalOpen(false); setSel(null); setBkErr(''); setBkOk(''); };
  const dr = `${days[0].label} - ${days[days.length-1].label}`;

  return (
    <>
      <div className="bg-card border border-border pixel-corners overflow-hidden">
        <div className="px-5 py-3 border-b border-border flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1">
          <h2 className="text-text" style={{ fontSize: 18, fontWeight: 700 }}>time slots</h2>
          <span className="text-text-muted" style={{ fontSize: 17 }}>{dr}</span>
        </div>
        <div className="overflow-auto max-h-[520px]">
          <div className="inline-block min-w-full">
            <table className="w-full border-collapse">
              <thead><tr>
                <th className="sticky top-0 left-0 z-20 bg-card border-b border-border w-16 min-w-[64px]" />
                {days.map((day) => (
                  <th key={day.date} className={`sticky top-0 z-10 border-b border-border px-1 py-2 text-center min-w-[60px] sm:min-w-[80px] ${day.isToday?'bg-brand-50/5':'bg-card'}`}>
                    <div className="text-text-muted tracking-wider" style={{ fontSize: 10 }}>{day.dayOfWeek}</div>
                    <div className={`mt-0.5 ${day.isToday?'text-brand-600 font-bold':'text-text-muted'}`} style={{ fontSize: 17 }}>{day.label}</div>
                  </th>
                ))}
              </tr></thead>
              <tbody>
                {timeSlots.map((slot, si) => {
                  const isHour = slot.minute===0;
                  return (
                    <tr key={slot.label}>
                      <td className={`sticky left-0 z-10 border-b border-border py-1 px-2 text-right text-text-muted ${isHour?'bg-surface-alt/80':'bg-card'}`} style={{ fontSize: 10 }}>
                        {isHour?slot.label:''}
                      </td>
                      {days.map((day) => {
                        const isAvail = availableCells.get(day.date)?.has(si)??false;
                        return (
                          <td key={`${day.date}-${slot.label}`} className={`border-b border-border-light p-0 relative ${day.isToday?'bg-brand-50/3':''}`}>
                            <div onClick={() => click(day, si)}
                              className={`h-7 sm:h-8 mx-0.5 my-px transition-all duration-100 ${isAvail?(isSelf?'bg-emerald-500/15 cursor-default':'bg-emerald-500/70 hover:bg-emerald-500/90 hover:scale-105 cursor-pointer'):'bg-transparent'}`}
                              style={{ clipPath: 'polygon(0 2px, 2px 2px, 2px 0, calc(100% - 2px) 0, calc(100% - 2px) 2px, 100% 2px, 100% calc(100% - 2px), calc(100% - 2px) calc(100% - 2px), calc(100% - 2px) 100%, 2px 100%, 2px calc(100% - 2px), 0 calc(100% - 2px))' }}
                              title={isAvail&&!isSelf?`${day.date} ${slot.label}`:undefined} />
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
          <div className="px-5 py-3 border-t border-border bg-surface-alt/50">
            <p className="text-text-muted text-center" style={{ fontSize: 17 }}>
              前往 <a href="/appointments" className="text-brand-600 hover:text-brand-700 font-bold">预约管理</a> 设置空闲时间
            </p>
          </div>
        )}
      </div>

      {modalOpen && sel && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4"
          style={{ background: 'rgba(0,0,0,0.5)', backdropFilter: 'blur(4px)' }}
          onClick={close}>
          <div className="bg-card border border-border pixel-corners w-full max-w-sm" onClick={(e) => e.stopPropagation()}>
            <div className="px-5 pt-5 pb-3">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-text" style={{ fontSize: 20, fontWeight: 700 }}>book slot</h3>
                <button onClick={close} className="w-7 h-7 flex items-center justify-center text-text-muted hover:text-text cursor-pointer border-none bg-transparent" style={{ fontSize: 18 }}>x</button>
              </div>
              <div className="flex items-center gap-3 px-3 py-2 pixel-corners-sm" style={{ background: 'rgba(120,184,128,0.08)', border: '1px solid rgba(120,184,128,0.2)' }}>
                <span style={{ width: 6, height: 6, background: 'var(--color-success)' }} />
                <div>
                  <div className="text-text" style={{ fontSize: 17, fontWeight: 700 }}>{sel.date} <span className="text-text-muted">{dayNames[new Date(sel.date).getDay()]}</span></div>
                  <div className="text-text-secondary" style={{ fontSize: 17 }}>{sel.startTime}-{sel.endTime}</div>
                </div>
              </div>
            </div>
            <div className="px-5 pb-3">
              <label className="flex flex-col gap-1">
                <span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>附言</span>
                <textarea value={msg} onChange={(e) => setMsg(e.target.value)}
                  className="w-full px-3 py-2 bg-surface border border-border text-text resize-y min-h-[70px] outline-none pixel-corners-sm"
                  style={{ fontSize: 18 }} rows={3} />
              </label>
              {bkErr && <div className="mt-2 px-3 py-2 text-danger pixel-corners-sm" style={{ fontSize: 17, background: 'rgba(224,112,112,0.06)' }}>{bkErr}</div>}
              {bkOk && <div className="mt-2 px-3 py-2 text-success pixel-corners-sm" style={{ fontSize: 17, background: 'rgba(120,184,128,0.06)' }}>{bkOk}</div>}
            </div>
            <div className="px-5 pb-5 flex gap-3">
              <button onClick={close} className="pixel-btn flex-1 justify-center" style={{ fontSize: 18 }}>取消</button>
              <button onClick={confirm} disabled={bking||!!bkOk} className="pixel-btn primary flex-1 justify-center" style={{ fontSize: 18 }}>{bking?'...':'confirm'}</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
