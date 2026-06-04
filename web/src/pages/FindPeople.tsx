import { useEffect, useState, useRef, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { apiGet } from '../api/client';
import RecruitmentCard, { type RecruitmentCardData } from '../components/RecruitmentCard';

const PRESET_SKILLS = ['React','Vue','TypeScript','JavaScript','Python','Go','Java','Node.js','Rust','C++','系统设计','算法','前端','后端','全栈','机器学习'];
const PRESET_COMPANIES = ['Google','Meta','Apple','Amazon','Microsoft','Netflix','Tesla','Stripe','Airbnb','Uber','ByteDance','Alibaba','Tencent'];
const ROLE_OPTIONS = [
  { value: '', label: 'all' },
  { value: '面试者', label: '面试者' },
  { value: '面试官', label: '面试官' },
  { value: '两者皆可', label: '两者皆可' },
];
const PAGE_SIZE = 20;

function FindPeople() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [keyword, setKeyword] = useState(searchParams.get('keyword') || '');
  const [selectedSkills, setSelectedSkills] = useState<string[]>(searchParams.get('skill')?.split(',').filter(Boolean) || []);
  const [selectedCompanies, setSelectedCompanies] = useState<string[]>(searchParams.get('company')?.split(',').filter(Boolean) || []);
  const [role, setRole] = useState(searchParams.get('role') || '');
  const [expMin, setExpMin] = useState(searchParams.get('exp_min') || '');
  const [expMax, setExpMax] = useState(searchParams.get('exp_max') || '');
  const [showFilters, setShowFilters] = useState(false);
  const [cards, setCards] = useState<RecruitmentCardData[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);
  const [toast, setToast] = useState('');
  const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const buildParams = useCallback((pageNum: number) => {
    const params = new URLSearchParams();
    if (keyword.trim()) params.set('keyword', keyword.trim());
    selectedSkills.forEach((s) => params.append('skill', s));
    selectedCompanies.forEach((c) => params.append('company', c));
    if (role) params.set('role', role);
    if (expMin) params.set('exp_min', expMin);
    if (expMax) params.set('exp_max', expMax);
    params.set('page', String(pageNum));
    params.set('size', String(PAGE_SIZE));
    return params;
  }, [keyword, selectedSkills, selectedCompanies, role, expMin, expMax]);

  const fetchCards = useCallback(async (pageNum: number, replace: boolean) => {
    setLoading(true); setError('');
    try {
      const params = buildParams(pageNum);
      const res = await apiGet<{ cards: RecruitmentCardData[]; total: number }>(`/recruitment-cards?${params.toString()}`);
      if (replace) setCards(res.cards); else setCards((prev) => [...prev, ...res.cards]);
      setTotal(res.total); setPage(pageNum);
    } catch { setError('加载失败'); }
    finally { setLoading(false); setInitialLoading(false); }
  }, [buildParams]);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      fetchCards(1, true);
      const syncParams = new URLSearchParams();
      if (keyword.trim()) syncParams.set('keyword', keyword.trim());
      if (selectedSkills.length > 0) syncParams.set('skill', selectedSkills.join(','));
      if (selectedCompanies.length > 0) syncParams.set('company', selectedCompanies.join(','));
      if (role) syncParams.set('role', role);
      if (expMin) syncParams.set('exp_min', expMin);
      if (expMax) syncParams.set('exp_max', expMax);
      setSearchParams(syncParams, { replace: true });
    }, 400);
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current); };
  }, [keyword, selectedSkills, selectedCompanies, role, expMin, expMax]);

  useEffect(() => {
    const created = searchParams.get('created');
    if (created) { setToast(decodeURIComponent(created)); const np = new URLSearchParams(searchParams); np.delete('created'); setSearchParams(np, { replace: true }); }
  }, []);
  useEffect(() => { if (toast) { const t = setTimeout(() => setToast(''), 3000); return () => clearTimeout(t); } }, [toast]);

  const handleLoadMore = () => fetchCards(page + 1, false);
  const toggleSkill = (s: string) => setSelectedSkills((prev) => prev.includes(s) ? prev.filter((x) => x !== s) : [...prev, s]);
  const toggleCompany = (c: string) => setSelectedCompanies((prev) => prev.includes(c) ? prev.filter((x) => x !== c) : [...prev, c]);
  const clearAll = () => { setKeyword(''); setSelectedSkills([]); setSelectedCompanies([]); setRole(''); setExpMin(''); setExpMax(''); };
  const hasFilters = keyword.trim() !== '' || selectedSkills.length > 0 || selectedCompanies.length > 0 || role !== '' || expMin !== '' || expMax !== '';
  const hasMore = cards.length < total;

  return (
    <div className="max-w-6xl mx-auto px-6 pb-12">
      {toast && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-50 px-4 py-2 text-success pixel-corners-sm"
             style={{ fontSize: 17, background: 'rgba(120,184,128,0.1)', border: '1px solid rgba(120,184,128,0.3)' }}>
          {toast}
        </div>
      )}

      <div className="sticky top-14 z-40 -mx-6 px-6 pt-6 pb-4 bg-surface/95 backdrop-blur-sm border-b border-border">
        <div className="flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-[200px] max-w-xl">
            <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted pointer-events-none" style={{ fontSize: 18 }}>&gt;</span>
            <input type="text" placeholder="搜索..." value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="w-full pl-8 pr-4 py-2.5 bg-card border border-border text-text pixel-corners-sm"
              style={{ fontSize: 18 }} />
          </div>
          <button onClick={() => setShowFilters(!showFilters)}
            className={`pixel-btn ${showFilters ? 'primary' : ''}`} style={{ fontSize: 18 }}>
            filter{hasFilters ? ' *' : ''}
          </button>
          {hasFilters && (
            <button onClick={clearAll} className="cursor-pointer border-none bg-transparent text-text-muted hover:text-text-secondary" style={{ fontSize: 18 }}>
              clear
            </button>
          )}
        </div>

        <div style={{ overflow: 'hidden', transition: 'max-height 0.2s steps(4), opacity 0.2s steps(4)', maxHeight: showFilters ? 600 : 0, opacity: showFilters ? 1 : 0 }}>
          <div className="bg-card border border-border pixel-corners p-5 mt-4 space-y-5">
            <div>
              <div className="text-text-muted mb-2 tracking-wider" style={{ fontSize: 17 }}>技能</div>
              <div className="flex flex-wrap gap-2">
                {PRESET_SKILLS.map((s) => (
                  <button key={s} onClick={() => toggleSkill(s)}
                    className={`pixel-tag cursor-pointer border-none ${selectedSkills.includes(s) ? 'primary' : ''}`}
                    style={{ fontSize: 17, background: selectedSkills.includes(s) ? 'var(--color-brand-600)' : 'transparent', color: selectedSkills.includes(s) ? '#1a1a1e' : 'var(--color-text-secondary)' }}>
                    {s}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <div className="text-text-muted mb-2 tracking-wider" style={{ fontSize: 17 }}>目标公司</div>
              <div className="flex flex-wrap gap-2">
                {PRESET_COMPANIES.map((c) => (
                  <button key={c} onClick={() => toggleCompany(c)}
                    className={`pixel-tag cursor-pointer border-none ${selectedCompanies.includes(c) ? 'primary' : ''}`}
                    style={{ fontSize: 17, background: selectedCompanies.includes(c) ? 'var(--color-success)' : 'transparent', color: selectedCompanies.includes(c) ? '#1a1a1e' : 'var(--color-text-secondary)' }}>
                    {c}
                  </button>
                ))}
              </div>
            </div>
            <div className="flex flex-wrap items-end gap-4">
              <div>
                <div className="text-text-muted mb-2 tracking-wider" style={{ fontSize: 17 }}>角色</div>
                <select value={role} onChange={(e) => setRole(e.target.value)}
                  className="px-3 py-1.5 bg-card border border-border text-text pixel-corners-sm cursor-pointer"
                  style={{ fontSize: 18 }}>
                  {ROLE_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                </select>
              </div>
              <div>
                <div className="text-text-muted mb-2 tracking-wider" style={{ fontSize: 17 }}>经验</div>
                <div className="flex items-center gap-2">
                  <input type="number" min="0" max="50" placeholder="min" value={expMin}
                    onChange={(e) => setExpMin(e.target.value)}
                    className="w-16 px-2 py-1.5 bg-card border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} />
                  <span className="text-text-muted">-</span>
                  <input type="number" min="0" max="50" placeholder="max" value={expMax}
                    onChange={(e) => setExpMax(e.target.value)}
                    className="w-16 px-2 py-1.5 bg-card border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-6">
        <div className="flex items-center justify-between mb-5">
          <h1 className="text-text" style={{ fontSize: 22, fontWeight: 700 }}>发现伙伴</h1>
          {!initialLoading && <span className="text-text-muted" style={{ fontSize: 18 }}>{total > 0 ? `${total} 个结果` : ''}</span>}
        </div>

        {initialLoading && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {Array.from({ length: 6 }).map((_, i) => <div key={i} className="skeleton h-72 pixel-corners" />)}
          </div>
        )}
        {!initialLoading && error && (
          <div className="p-8 text-center text-danger pixel-corners" style={{ background: 'rgba(224,112,112,0.06)', border: '1px solid rgba(224,112,112,0.2)' }}>{error}</div>
        )}
        {!initialLoading && !error && cards.length === 0 && (
          <div className="text-center py-20 animate-fade-up">
            <div className="text-text-muted mb-4" style={{ fontSize: 40 }}>?</div>
            <p className="text-text" style={{ fontSize: 20, fontWeight: 700 }}>没有找到伙伴</p>
            <p className="text-text-muted mt-2" style={{ fontSize: 17 }}>尝试调整筛选条件</p>
            <Link to="/my-card" className="pixel-btn primary no-underline mt-6 inline-flex" style={{ fontSize: 18 }}>创建名片</Link>
          </div>
        )}
        {cards.length > 0 && (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {cards.map((card) => <RecruitmentCard key={card.id} card={card} />)}
            </div>
            <div className="mt-8 text-center">
              {loading && <span className="text-text-muted" style={{ fontSize: 18 }}>loading...</span>}
              {!loading && hasMore && (
                <button onClick={handleLoadMore} className="pixel-btn" style={{ fontSize: 18 }}>加载更多</button>
              )}
              {!loading && !hasMore && total > 0 && (
                <span className="text-text-muted" style={{ fontSize: 18 }}>— {total} 个结果 —</span>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default FindPeople;
