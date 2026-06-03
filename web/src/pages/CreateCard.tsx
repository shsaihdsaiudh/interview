import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiGet, apiPut, getApiErrorMessage } from '../api/client';
import { getUser } from '../components/Navbar';

// ── Types ──

interface RecruitmentCard {
  id?: string;
  skills: string[];
  target_companies: string[];
  role: 'interviewee' | 'interviewer' | 'both';
  experience_years: number;
  bio: string;
  open_to_appointment: boolean;
}

// ── Presets ──

const SKILL_PRESETS = [
  'React', 'Vue', 'Angular', 'TypeScript', 'JavaScript',
  'Go', 'Java', 'Python', 'Rust', 'Node.js',
  '算法', '系统设计', 'CSS', 'HTML', 'SQL',
  'Docker', 'Kubernetes', 'AWS', 'Linux', 'Git',
  'C++', 'C#', 'Swift', 'Kotlin', 'PHP',
  'MongoDB', 'PostgreSQL', 'Redis', 'GraphQL', 'REST',
];

const COMPANY_PRESETS = [
  '字节跳动', '腾讯', '阿里巴巴', '美团',
  'Google', 'Microsoft', 'Amazon', 'Apple',
  'Meta', 'Netflix', '百度', '京东',
  '网易', '小米', '华为',
  '滴滴', '快手', '拼多多', 'B站',
];

const ROLE_OPTIONS = [
  { value: 'interviewee', label: '面试者' },
  { value: 'interviewer', label: '面试官' },
  { value: 'both', label: '两者皆可' },
] as const;

// ── MultiSelect (Notion‑style) ──

interface MultiSelectProps {
  label: string;
  placeholder: string;
  values: string[];
  onChange: (values: string[]) => void;
  presets: string[];
  error?: string;
}

function MultiSelect({ label, placeholder, values, onChange, presets, error }: MultiSelectProps) {
  const [input, setInput] = useState('');
  const [open, setOpen] = useState(false);
  const [highlightIdx, setHighlightIdx] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const filtered = presets.filter(
    (p) => p.toLowerCase().includes(input.toLowerCase()) && !values.includes(p),
  );

  const showAddNew =
    input.trim() !== '' &&
    !presets.some((p) => p.toLowerCase() === input.trim().toLowerCase()) &&
    !values.includes(input.trim());

  useEffect(() => { setHighlightIdx(0); }, [input]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const addValue = (val: string) => {
    const trimmed = val.trim();
    if (trimmed && !values.includes(trimmed)) {
      onChange([...values, trimmed]);
    }
    setInput('');
    setOpen(false);
  };

  const removeValue = (val: string) => {
    onChange(values.filter((v) => v !== val));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      const items: string[] = [...filtered];
      if (showAddNew) items.push(input.trim());
      if (open && items.length > 0 && highlightIdx < items.length) {
        addValue(items[highlightIdx]);
      } else if (input.trim()) {
        addValue(input.trim());
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      setOpen(true);
      const max = filtered.length + (showAddNew ? 1 : 0) - 1;
      setHighlightIdx((prev) => Math.min(prev + 1, Math.max(0, max)));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlightIdx((prev) => Math.max(prev - 1, 0));
    } else if (e.key === 'Escape') {
      setOpen(false);
    } else if (e.key === 'Backspace' && input === '' && values.length > 0) {
      removeValue(values[values.length - 1]);
    }
  };

  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium text-text">{label}</label>
      <div ref={containerRef} className="relative">
        <div
          className={`flex flex-wrap items-center gap-1.5 p-2 min-h-[42px] rounded-xl border bg-white cursor-pointer transition ${
            error
              ? 'border-danger'
              : open
                ? 'border-brand-400'
                : 'border-border hover:border-brand-200'
          }`}
          style={open ? { boxShadow: '0 0 0 3px rgba(99,102,241,0.1)' } : undefined}
          onClick={() => { inputRef.current?.focus(); setOpen(true); }}
        >
          {values.map((v) => (
            <span
              key={v}
              className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-brand-50 text-brand-700 text-xs font-medium"
            >
              {v}
              <button
                type="button"
                onClick={(e) => { e.stopPropagation(); removeValue(v); }}
                className="cursor-pointer border-none bg-transparent p-0 leading-none text-brand-400 hover:text-danger transition"
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </span>
          ))}
          <input
            ref={inputRef}
            type="text"
            className="flex-1 min-w-[120px] border-none outline-none bg-transparent text-sm text-text placeholder:text-text-muted py-0.5"
            placeholder={values.length === 0 ? placeholder : ''}
            value={input}
            onChange={(e) => { setInput(e.target.value); setOpen(true); }}
            onFocus={() => setOpen(true)}
            onKeyDown={handleKeyDown}
          />
        </div>
        {open && (filtered.length > 0 || showAddNew) && (
          <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-border rounded-xl shadow-sm z-50 max-h-48 overflow-y-auto py-1">
            {filtered.map((item, idx) => (
              <button
                key={item}
                type="button"
                className={`w-full text-left px-3 py-2 text-sm cursor-pointer border-none bg-transparent transition ${
                  idx === highlightIdx ? 'bg-brand-50 text-brand-700' : 'text-text hover:bg-surface-alt'
                }`}
                onMouseEnter={() => setHighlightIdx(idx)}
                onClick={() => addValue(item)}
              >
                {item}
              </button>
            ))}
            {showAddNew && (
              <button
                type="button"
                className={`w-full text-left px-3 py-2 text-sm cursor-pointer border-none bg-transparent transition flex items-center gap-2 ${
                  highlightIdx === filtered.length ? 'bg-brand-50 text-brand-700' : 'text-text-muted hover:bg-surface-alt'
                }`}
                onMouseEnter={() => setHighlightIdx(filtered.length)}
                onClick={() => addValue(input.trim())}
              >
                <span>添加</span>
                <span className="text-brand-600 font-medium">"{input.trim()}"</span>
              </button>
            )}
          </div>
        )}
      </div>
      {error && <p className="text-xs text-danger mt-1">{error}</p>}
    </div>
  );
}

// ── Page Component ──

export default function CreateCard() {
  const navigate = useNavigate();
  const user = getUser();

  const [form, setForm] = useState<RecruitmentCard>({
    skills: [],
    target_companies: [],
    role: 'both',
    experience_years: 0,
    bio: '',
    open_to_appointment: true,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [isEdit, setIsEdit] = useState(false);
  const [submitError, setSubmitError] = useState('');

  useEffect(() => {
    if (!user) {
      navigate('/login', { replace: true });
      return;
    }

    apiGet<RecruitmentCard>('/recruitment-card?user_id=self')
      .then((card) => {
        if (card) {
          setForm({
            skills: card.skills || [],
            target_companies: card.target_companies || [],
            role: card.role || 'both',
            experience_years: card.experience_years ?? 0,
            bio: card.bio || '',
            open_to_appointment: card.open_to_appointment ?? true,
          });
          setIsEdit(true);
        }
      })
      .catch((err) => {
        if (err?.response?.status !== 404) {
          console.error('Failed to fetch card:', err);
        }
      })
      .finally(() => setLoading(false));
  }, [user, navigate]);

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (form.skills.length === 0) newErrors.skills = '请至少选择 1 个技能标签';
    if (form.target_companies.length === 0) newErrors.target_companies = '请至少选择 1 个目标公司';
    if (form.experience_years < 0) newErrors.experience_years = '经验年限不能为负数';
    if (form.bio.length > 500) newErrors.bio = '个人简介不能超过 500 字';
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError('');
    if (!validate()) return;

    setSubmitting(true);
    try {
      await apiPut('/recruitment-card', form);
      navigate('/find', { replace: true });
    } catch (err) {
      setSubmitError(getApiErrorMessage(err, '保存失败，请重试'));
    } finally {
      setSubmitting(false);
    }
  };

  if (!user) return null;

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto px-6 pt-24 pb-24">
        <div className="space-y-6">
          <div className="skeleton h-8 w-48" />
          <div className="skeleton h-96 w-full" />
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto px-6 pt-12 pb-24">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-text">
          {isEdit ? '编辑我的名片' : '创建我的名片'}
        </h1>
        <p className="text-sm text-text-secondary mt-2">
          {isEdit
            ? '更新你的招募信息，让更多人找到你'
            : '创建你的招募卡片，开始寻找面试伙伴'}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-8">
        {/* Skills */}
        <MultiSelect
          label="技能标签"
          placeholder="搜索或输入技能标签..."
          values={form.skills}
          onChange={(skills) => setForm({ ...form, skills })}
          presets={SKILL_PRESETS}
          error={errors.skills}
        />

        {/* Target Companies */}
        <MultiSelect
          label="目标公司"
          placeholder="搜索或输入目标公司..."
          values={form.target_companies}
          onChange={(target_companies) => setForm({ ...form, target_companies })}
          presets={COMPANY_PRESETS}
          error={errors.target_companies}
        />

        {/* Role */}
        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-text">角色</label>
          <div className="flex gap-2">
            {ROLE_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                type="button"
                className={`px-4 py-2 rounded-lg text-sm font-medium cursor-pointer border transition ${
                  form.role === opt.value
                    ? 'bg-brand-600 text-white border-brand-600'
                    : 'bg-white text-text-secondary border-border hover:border-brand-200 hover:text-brand-600'
                }`}
                onClick={() => setForm({ ...form, role: opt.value })}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        {/* Experience Years */}
        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-text">经验年限</label>
          <div className="flex items-center gap-4">
            <input
              type="range"
              min="0"
              max="20"
              step="0.5"
              value={form.experience_years}
              onChange={(e) =>
                setForm({ ...form, experience_years: parseFloat(e.target.value) || 0 })
              }
              className="flex-1 h-2 rounded-full appearance-none bg-surface-alt cursor-pointer
                         accent-brand-600
                         [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-5
                         [&::-webkit-slider-thumb]:h-5 [&::-webkit-slider-thumb]:rounded-full
                         [&::-webkit-slider-thumb]:bg-brand-600 [&::-webkit-slider-thumb]:cursor-pointer
                         [&::-webkit-slider-thumb]:shadow-sm
                         [&::-moz-range-thumb]:w-5 [&::-moz-range-thumb]:h-5
                         [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-brand-600
                         [&::-moz-range-thumb]:border-none [&::-moz-range-thumb]:cursor-pointer"
            />
            <div className="flex items-center gap-1.5 bg-white border border-border rounded-lg px-3 py-1.5 w-20">
              <input
                type="number"
                min="0"
                max="50"
                step="0.5"
                value={form.experience_years}
                onChange={(e) =>
                  setForm({ ...form, experience_years: parseFloat(e.target.value) || 0 })
                }
                className="w-full text-sm text-center border-none outline-none bg-transparent text-text font-medium
                           [appearance:textfield]
                           [&::-webkit-outer-spin-button]:appearance-none
                           [&::-webkit-inner-spin-button]:appearance-none"
              />
              <span className="text-xs text-text-muted shrink-0">年</span>
            </div>
          </div>
          {errors.experience_years && (
            <p className="text-xs text-danger mt-1">{errors.experience_years}</p>
          )}
        </div>

        {/* Bio */}
        <div className="space-y-1.5">
          <label className="block text-sm font-medium text-text">
            个人简介
            <span className="text-text-muted font-normal ml-1">（选填）</span>
          </label>
          <textarea
            className={`w-full min-h-[120px] rounded-xl border px-4 py-3 text-sm text-text placeholder:text-text-muted bg-white transition resize-y outline-none ${
              errors.bio
                ? 'border-danger'
                : 'border-border hover:border-brand-200 focus:border-brand-400'
            }`}
            placeholder="简单介绍一下自己，让面试伙伴更好地了解你..."
            maxLength={500}
            value={form.bio}
            onChange={(e) => setForm({ ...form, bio: e.target.value })}
            onFocus={(e) => {
              if (!errors.bio) {
                e.currentTarget.style.boxShadow = '0 0 0 3px rgba(99,102,241,0.1)';
              }
            }}
            onBlur={(e) => {
              e.currentTarget.style.boxShadow = 'none';
            }}
          />
          <div className="flex justify-between items-center">
            {errors.bio ? (
              <p className="text-xs text-danger">{errors.bio}</p>
            ) : (
              <span />
            )}
            <span
              className={`text-xs ${form.bio.length > 450 ? 'text-warning' : 'text-text-muted'}`}
            >
              {form.bio.length}/500
            </span>
          </div>
        </div>

        {/* Open to appointment toggle */}
        <div className="flex items-center justify-between p-4 rounded-xl bg-white border border-border">
          <div>
            <div className="text-sm font-medium text-text">开放被预约</div>
            <div className="text-xs text-text-muted mt-0.5">
              开启后，其他用户可以预约你的面试时间
            </div>
          </div>
          <button
            type="button"
            role="switch"
            aria-checked={form.open_to_appointment}
            onClick={() =>
              setForm({ ...form, open_to_appointment: !form.open_to_appointment })
            }
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition cursor-pointer border-none ${
              form.open_to_appointment
                ? 'bg-brand-600'
                : 'bg-surface-alt border border-border'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white shadow-sm transition ${
                form.open_to_appointment ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
        </div>

        {/* Submit error */}
        {submitError && (
          <div className="p-3 rounded-xl bg-red-50 border border-red-200 text-sm text-danger">
            {submitError}
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 py-3 rounded-xl bg-brand-600 text-white font-medium text-sm
                       hover:bg-brand-700 transition disabled:opacity-50 disabled:cursor-not-allowed
                       cursor-pointer border-none"
          >
            {submitting ? '保存中...' : isEdit ? '更新名片' : '创建名片'}
          </button>
          <button
            type="button"
            onClick={() => navigate('/')}
            className="px-6 py-3 rounded-xl border border-border text-text-secondary font-medium text-sm
                       hover:border-brand-200 hover:text-brand-600 transition bg-white cursor-pointer"
          >
            取消
          </button>
        </div>
      </form>
    </div>
  );
}
