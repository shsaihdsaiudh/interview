import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiGet, apiPut, getApiErrorMessage } from '../api/client';
import { getUser } from '../components/Navbar';

interface RecruitmentCard { id?: string; skills: string[]; target_companies: string[]; role: '面试者'|'面试官'|'两者皆可'; experience_years: number; bio: string; open_to_appointment: boolean; }

const SKILLS = ['React','Vue','Angular','TypeScript','JavaScript','Go','Java','Python','Rust','Node.js','算法','系统设计','CSS','HTML','SQL','Docker','Kubernetes','AWS','Linux','Git','C++','C#','Swift','Kotlin','PHP','MongoDB','PostgreSQL','Redis','GraphQL','REST'];
const COMPANIES = ['字节跳动','腾讯','阿里巴巴','美团','Google','Microsoft','Amazon','Apple','Meta','Netflix','百度','京东','网易','小米','华为','滴滴','快手','拼多多','B站'];
const ROLES = [{v:'面试者',l:'面试者'},{v:'面试官',l:'面试官'},{v:'两者皆可',l:'两者皆可'}] as const;

function MultiSelect({ label, placeholder, values, onChange, presets, error }: { label: string; placeholder: string; values: string[]; onChange: (v: string[]) => void; presets: string[]; error?: string; }) {
  const [input, setInput] = useState(''); const [open, setOpen] = useState(false); const [hi, setHi] = useState(0);
  const cr = useRef<HTMLDivElement>(null); const ir = useRef<HTMLInputElement>(null);
  const filtered = presets.filter((p) => p.toLowerCase().includes(input.toLowerCase()) && !values.includes(p));
  const showNew = input.trim() !== '' && !presets.some((p) => p.toLowerCase() === input.trim().toLowerCase()) && !values.includes(input.trim());
  useEffect(() => { setHi(0); }, [input]);
  useEffect(() => { const h = (e: MouseEvent) => { if (cr.current && !cr.current.contains(e.target as Node)) setOpen(false); }; document.addEventListener('mousedown', h); return () => document.removeEventListener('mousedown', h); }, []);
  const add = (v: string) => { const t = v.trim(); if (t && !values.includes(t)) onChange([...values, t]); setInput(''); setOpen(false); };
  const remove = (v: string) => { onChange(values.filter((x) => x !== v)); };
  const kd = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') { e.preventDefault(); const items = [...filtered]; if (showNew) items.push(input.trim()); if (open && items.length > 0 && hi < items.length) add(items[hi]); else if (input.trim()) add(input.trim()); }
    else if (e.key === 'ArrowDown') { e.preventDefault(); setOpen(true); setHi((p) => Math.min(p+1, Math.max(0, filtered.length + (showNew?1:0) - 1))); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setHi((p) => Math.max(p-1, 0)); }
    else if (e.key === 'Escape') setOpen(false);
    else if (e.key === 'Backspace' && input === '' && values.length > 0) remove(values[values.length-1]);
  };
  return (
    <div style={{ marginBottom: 8 }}>
      <label className="block text-text-muted mb-1 tracking-wider" style={{ fontSize: 17 }}>{label}</label>
      <div ref={cr} style={{ position: 'relative' }}>
        <div onClick={() => { ir.current?.focus(); setOpen(true); }}
          className={`flex flex-wrap items-center gap-1.5 p-2 min-h-[42px] bg-card border cursor-pointer pixel-corners-sm ${error ? 'border-danger' : open ? 'border-brand-400' : 'border-border hover:border-brand-400'}`}>
          {values.map((v) => (
            <span key={v} className="inline-flex items-center gap-1 px-2 py-0.5 text-brand-600 pixel-tag" style={{ borderColor: 'rgba(224,184,104,0.2)' }}>
              {v}
              <button type="button" onClick={(e) => { e.stopPropagation(); remove(v); }} className="cursor-pointer border-none bg-transparent p-0 text-text-muted hover:text-danger" style={{ fontSize: 18 }}>x</button>
            </span>
          ))}
          <input ref={ir} type="text" className="flex-1 min-w-[100px] border-none outline-none bg-transparent text-text placeholder:text-text-muted" style={{ fontSize: 18 }} placeholder={values.length===0?placeholder:''} value={input} onChange={(e) => { setInput(e.target.value); setOpen(true); }} onFocus={() => setOpen(true)} onKeyDown={kd} />
        </div>
        {open && (filtered.length > 0 || showNew) && (
          <div className="absolute top-full left-0 right-0 mt-1 bg-card border border-border pixel-corners-sm z-50 max-h-48 overflow-y-auto py-1">
            {filtered.map((item, idx) => (
              <button key={item} type="button" className={`w-full text-left px-3 py-2 cursor-pointer border-none bg-transparent ${idx===hi?'text-brand-600 bg-brand-50':'text-text hover:bg-surface-alt'}`}
                style={{ fontSize: 17 }} onMouseEnter={() => setHi(idx)} onClick={() => add(item)}>{item}</button>
            ))}
            {showNew && (
              <button type="button" className={`w-full text-left px-3 py-2 cursor-pointer border-none bg-transparent flex items-center gap-2 ${hi===filtered.length?'text-brand-600 bg-brand-50':'text-text-muted hover:bg-surface-alt'}`}
                style={{ fontSize: 17 }} onMouseEnter={() => setHi(filtered.length)} onClick={() => add(input.trim())}>
                + "{input.trim()}"
              </button>
            )}
          </div>
        )}
      </div>
      {error && <p className="text-danger mt-1" style={{ fontSize: 18 }}>{error}</p>}
    </div>
  );
}

export default function CreateCard() {
  const navigate = useNavigate(); const user = getUser();
  const [form, setForm] = useState<RecruitmentCard>({ skills: [], target_companies: [], role: '两者皆可', experience_years: 0, bio: '', open_to_appointment: true });
  const [errors, setErrors] = useState<Record<string,string>>({});
  const [loading, setLoading] = useState(true); const [submitting, setSubmitting] = useState(false);
  const [isEdit, setIsEdit] = useState(false); const [submitError, setSubmitError] = useState('');

  useEffect(() => { if (!user) { navigate('/login', { replace: true }); return; }
    apiGet<RecruitmentCard>('/recruitment-card?user_id=self').then((card) => { if (card) { setForm({ skills: card.skills||[], target_companies: card.target_companies||[], role: card.role||'两者皆可', experience_years: card.experience_years??0, bio: card.bio||'', open_to_appointment: card.open_to_appointment??true }); setIsEdit(true); } }).catch((err) => { if (err?.response?.status !== 404) console.error(err); }).finally(() => setLoading(false));
  }, [user, navigate]);

  const validate = (): boolean => { const ne: Record<string,string> = {}; if (form.skills.length===0) ne.skills='至少选择 1 个技能'; if (form.target_companies.length===0) ne.target_companies='至少选择 1 个公司'; if (form.experience_years<0) ne.experience_years='无效值'; if (form.bio.length>500) ne.bio='最多 500 字'; setErrors(ne); return Object.keys(ne).length===0; };

  const submit = async (e: React.FormEvent) => { e.preventDefault(); setSubmitError(''); if (!validate()) return; setSubmitting(true);
    try { await apiPut('/recruitment-card', form); navigate(`/find?created=${encodeURIComponent(isEdit?'名片已更新':'名片已发布')}`, { replace: true }); }
    catch (err) { setSubmitError(getApiErrorMessage(err, '保存失败')); } finally { setSubmitting(false); } };

  if (!user) return null;
  if (loading) return <div className="max-w-2xl mx-auto px-6 pt-24"><div className="skeleton h-8 w-48 pixel-corners" /><div className="skeleton h-96 pixel-corners mt-6" /></div>;

  return (
    <div className="max-w-2xl mx-auto px-6 pt-12 pb-24">
      <div className="mb-8 animate-fade-up">
        <div className="text-text-muted tracking-wider mb-2" style={{ fontSize: 18 }}>{isEdit ? '[编辑]' : '[新建]'}</div>
        <h1 className="text-text" style={{ fontSize: 24, fontWeight: 700 }}>{isEdit ? '编辑名片' : '创建名片'}</h1>
        <p className="text-text-secondary mt-2" style={{ fontSize: 17 }}>{isEdit ? '更新你的招募信息' : '创建你的招募卡片'}</p>
      </div>

      <form onSubmit={submit} className="space-y-6">
        <MultiSelect label="SKILLS" placeholder="搜索..." values={form.skills} onChange={(s) => setForm({...form, skills: s})} presets={SKILLS} error={errors.skills} />
        <MultiSelect label="TARGET COMPANIES" placeholder="搜索..." values={form.target_companies} onChange={(c) => setForm({...form, target_companies: c})} presets={COMPANIES} error={errors.target_companies} />

        <div>
          <label className="block text-text-muted mb-1 tracking-wider" style={{ fontSize: 17 }}>角色</label>
          <div className="flex gap-2">
            {ROLES.map((o) => (
              <button key={o.v} type="button"
                className={`pixel-btn ${form.role===o.v?'primary':''}`} style={{ fontSize: 18 }}
                onClick={() => setForm({...form, role: o.v})}>{o.l}</button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-text-muted mb-1 tracking-wider" style={{ fontSize: 17 }}>经验</label>
          <div className="flex items-center gap-4">
            <input type="range" min="0" max="20" step="0.5" value={form.experience_years}
              onChange={(e) => setForm({...form, experience_years: parseFloat(e.target.value)||0})}
              className="flex-1 h-2 bg-surface-alt cursor-pointer" style={{ accentColor: 'var(--color-brand-600)' }} />
            <div className="flex items-center gap-1 bg-card border border-border px-3 py-1.5 pixel-corners-sm w-18">
              <input type="number" min="0" max="50" step="0.5" value={form.experience_years}
                onChange={(e) => setForm({...form, experience_years: parseFloat(e.target.value)||0})}
                className="w-full text-center border-none outline-none bg-transparent text-text" style={{ fontSize: 18 }} />
              <span className="text-text-muted" style={{ fontSize: 18 }}>y</span>
            </div>
          </div>
        </div>

        <div>
          <label className="block text-text-muted mb-1 tracking-wider" style={{ fontSize: 17 }}>BIO <span className="text-text-muted">(（选填）)</span></label>
          <textarea className={`w-full min-h-[120px] bg-card border px-3 py-3 text-text placeholder:text-text-muted resize-y outline-none pixel-corners-sm ${errors.bio?'border-danger':'border-border hover:border-brand-400'}`}
            style={{ fontSize: 18 }} placeholder="介绍自己..."
            maxLength={500} value={form.bio} onChange={(e) => setForm({...form, bio: e.target.value})} />
          <div className="flex justify-end"><span className={form.bio.length>450?'text-warning':'text-text-muted'} style={{ fontSize: 18 }}>{form.bio.length}/500</span></div>
        </div>

        <div className="flex items-center justify-between p-4 bg-card border border-border pixel-corners">
          <div><div className="text-text" style={{ fontSize: 17, fontWeight: 700 }}>开放被预约</div><div className="text-text-muted mt-0.5" style={{ fontSize: 18 }}>开启后其他用户可以预约你的面试时间</div></div>
          <button type="button" role="switch" aria-checked={form.open_to_appointment}
            onClick={() => setForm({...form, open_to_appointment: !form.open_to_appointment})}
            className="relative inline-flex h-6 w-11 items-center transition cursor-pointer border-none"
            style={{ background: form.open_to_appointment?'var(--color-brand-600)':'var(--color-surface-alt)', border: form.open_to_appointment?'none':'1px solid var(--color-border)' }}>
            <span className={`inline-block h-4 w-4 bg-white shadow-sm transition ${form.open_to_appointment?'translate-x-6':'translate-x-1'}`} />
          </button>
        </div>

        {submitError && <div className="p-3 text-danger pixel-corners-sm" style={{ fontSize: 18, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>{submitError}</div>}

        <div className="flex items-center gap-3 pt-2">
          <button type="submit" disabled={submitting} className="pixel-btn primary flex-1 justify-center" style={{ fontSize: 18, padding: '12px' }}>{submitting?'...':isEdit?'更新名片':'创建名片'}</button>
          <button type="button" onClick={() => navigate('/')} className="pixel-btn" style={{ fontSize: 18 }}>取消</button>
        </div>
      </form>
    </div>
  );
}
