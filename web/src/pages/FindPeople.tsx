import { useEffect, useState, useRef, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { apiGet } from '../api/client';
import RecruitmentCard, { type RecruitmentCardData } from '../components/RecruitmentCard';

// ── 预设筛选标签 ──

const PRESET_SKILLS = [
  'React', 'Vue', 'TypeScript', 'JavaScript', 'Python', 'Go', 'Java',
  'Node.js', 'Rust', 'C++', '系统设计', '算法', '前端', '后端', '全栈', '机器学习',
];

const PRESET_COMPANIES = [
  'Google', 'Meta', 'Apple', 'Amazon', 'Microsoft', 'Netflix', 'Tesla',
  'Stripe', 'Airbnb', 'Uber', 'ByteDance', 'Alibaba', 'Tencent',
];

const ROLE_OPTIONS = [
  { value: '', label: '全部角色' },
  { value: 'interviewee', label: '面试者' },
  { value: 'interviewer', label: '面试官' },
  { value: 'both', label: '两者皆可' },
];

const PAGE_SIZE = 20;

function FindPeople() {
  const [searchParams, setSearchParams] = useSearchParams();

  const [keyword, setKeyword] = useState(searchParams.get('keyword') || '');
  const [selectedSkills, setSelectedSkills] = useState<string[]>(
    searchParams.get('skill')?.split(',').filter(Boolean) || []
  );
  const [selectedCompanies, setSelectedCompanies] = useState<string[]>(
    searchParams.get('company')?.split(',').filter(Boolean) || []
  );
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

  const buildParams = useCallback(
    (pageNum: number) => {
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
    },
    [keyword, selectedSkills, selectedCompanies, role, expMin, expMax]
  );

  const fetchCards = useCallback(
    async (pageNum: number, replace: boolean) => {
      setLoading(true);
      setError('');
      try {
        const params = buildParams(pageNum);
        const res = await apiGet<{ cards: RecruitmentCardData[]; total: number }>(
          `/recruitment-cards?${params.toString()}`
        );
        if (replace) {
          setCards(res.cards);
        } else {
          setCards((prev) => [...prev, ...res.cards]);
        }
        setTotal(res.total);
        setPage(pageNum);
      } catch {
        setError('加载失败，请稍后重试');
      } finally {
        setLoading(false);
        setInitialLoading(false);
      }
    },
    [buildParams]
  );

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

  // ── toast on card created/updated ──
  useEffect(() => {
    const created = searchParams.get('created');
    if (created) {
      setToast(decodeURIComponent(created));
      const newParams = new URLSearchParams(searchParams);
      newParams.delete('created');
      setSearchParams(newParams, { replace: true });
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (toast) {
      const timer = setTimeout(() => setToast(''), 3000);
      return () => clearTimeout(timer);
    }
  }, [toast]);

  const handleLoadMore = () => fetchCards(page + 1, false);

  const toggleSkill = (skill: string) => {
    setSelectedSkills((prev) =>
      prev.includes(skill) ? prev.filter((s) => s !== skill) : [...prev, skill]
    );
  };
  const toggleCompany = (company: string) => {
    setSelectedCompanies((prev) =>
      prev.includes(company) ? prev.filter((c) => c !== company) : [...prev, company]
    );
  };
  const clearAllFilters = () => {
    setKeyword(''); setSelectedSkills([]); setSelectedCompanies([]);
    setRole(''); setExpMin(''); setExpMax('');
  };

  const hasActiveFilters =
    keyword.trim() !== '' || selectedSkills.length > 0 ||
    selectedCompanies.length > 0 || role !== '' || expMin !== '' || expMax !== '';
  const hasMore = cards.length < total;

  return (
    <div className="max-w-6xl mx-auto px-6 pb-12">
      {/* ── Toast ── */}
      {toast && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-50 px-4 py-2.5 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-700 text-sm font-medium shadow-sm">
          {toast}
        </div>
      )}

      {/* ── 搜索栏（固定）── */}
      <div className="sticky top-14 z-40 -mx-6 px-6 pt-6 pb-4 bg-surface/95 backdrop-blur-sm border-b border-border">
        <div className="flex items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-[200px] max-w-xl">
            <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none"
              fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" viewBox="0 0 24 24">
              <circle cx="11" cy="11" r="8" /><path d="M21 21l-4.3-4.3" />
            </svg>
            <input type="text" placeholder="搜索技能、公司、简介..." value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 rounded-xl border border-border bg-white text-sm
                         placeholder:text-text-muted focus-visible:border-brand-400 outline-none transition" />
          </div>
          <button onClick={() => setShowFilters(!showFilters)}
            className={`flex items-center gap-1.5 px-4 py-2.5 rounded-xl border text-sm font-medium transition cursor-pointer
              ${showFilters ? 'bg-brand-50 border-brand-200 text-brand-700' : 'bg-white border-border text-text-secondary hover:border-brand-200 hover:text-brand-600'}`}>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" viewBox="0 0 24 24">
              <path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3" />
              <circle cx="4" cy="12" r="2" /><circle cx="12" cy="10" r="2" /><circle cx="20" cy="14" r="2" />
            </svg>
            筛选{hasActiveFilters && <span className="w-2 h-2 rounded-full bg-brand-600" />}
          </button>
          {hasActiveFilters && (
            <button onClick={clearAllFilters}
              className="text-sm text-text-muted hover:text-text-secondary transition cursor-pointer border-none bg-transparent">
              清除
            </button>
          )}
        </div>

        <div className={`overflow-hidden transition-all duration-250 ${showFilters ? 'max-h-[600px] opacity-100 mt-4' : 'max-h-0 opacity-0'}`}>
          <div className="bg-card rounded-2xl border border-border p-5 space-y-5">
            <div>
              <div className="text-xs font-medium text-text-muted mb-2">技能</div>
              <div className="flex flex-wrap gap-2">
                {PRESET_SKILLS.map((skill) => (
                  <button key={skill} onClick={() => toggleSkill(skill)}
                    className={`px-3 py-1.5 rounded-lg text-xs font-medium transition cursor-pointer border-none
                      ${selectedSkills.includes(skill) ? 'bg-brand-600 text-white' : 'bg-surface-alt text-text-secondary hover:bg-brand-50 hover:text-brand-700'}`}>
                    {skill}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <div className="text-xs font-medium text-text-muted mb-2">目标公司</div>
              <div className="flex flex-wrap gap-2">
                {PRESET_COMPANIES.map((company) => (
                  <button key={company} onClick={() => toggleCompany(company)}
                    className={`px-3 py-1.5 rounded-lg text-xs font-medium transition cursor-pointer border-none
                      ${selectedCompanies.includes(company) ? 'bg-emerald-600 text-white' : 'bg-surface-alt text-text-secondary hover:bg-emerald-50 hover:text-emerald-700'}`}>
                    {company}
                  </button>
                ))}
              </div>
            </div>
            <div className="flex flex-wrap items-end gap-4">
              <div>
                <div className="text-xs font-medium text-text-muted mb-2">角色</div>
                <select value={role} onChange={(e) => setRole(e.target.value)}
                  className="px-3 py-1.5 rounded-lg border border-border bg-white text-sm text-text
                             focus-visible:border-brand-400 outline-none transition cursor-pointer">
                  {ROLE_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <div className="text-xs font-medium text-text-muted mb-2">经验年限</div>
                <div className="flex items-center gap-2">
                  <input type="number" min="0" max="50" placeholder="最小" value={expMin}
                    onChange={(e) => setExpMin(e.target.value)}
                    className="w-20 px-3 py-1.5 rounded-lg border border-border bg-white text-sm
                               placeholder:text-text-muted focus-visible:border-brand-400 outline-none transition" />
                  <span className="text-text-muted text-sm">-</span>
                  <input type="number" min="0" max="50" placeholder="最大" value={expMax}
                    onChange={(e) => setExpMax(e.target.value)}
                    className="w-20 px-3 py-1.5 rounded-lg border border-border bg-white text-sm
                               placeholder:text-text-muted focus-visible:border-brand-400 outline-none transition" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* ── 结果区域 ── */}
      <div className="mt-6">
        <div className="flex items-center justify-between mb-5">
          <div>
            <h1 className="text-xl font-bold text-text">发现面试伙伴</h1>
            {!initialLoading && (<p className="text-sm text-text-muted mt-0.5">{total > 0 ? `共找到 ${total} 位伙伴` : ''}</p>)}
          </div>
        </div>

        {initialLoading && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {Array.from({ length: 6 }).map((_, i) => (<div key={i} className="skeleton h-72 rounded-2xl" />))}
          </div>
        )}

        {!initialLoading && error && (
          <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">{error}</div>
        )}

        {!initialLoading && !error && cards.length === 0 && (
          <div className="text-center py-20">
            <div className="text-5xl mb-4">🔍</div>
            <p className="text-lg font-medium text-text">没有找到合适的伙伴</p>
            <p className="text-sm text-text-muted mt-1.5 max-w-sm mx-auto leading-relaxed">
              尝试调整筛选条件，或者发布招募帖让更多人看到你
            </p>
            <Link to="/posts"
              className="inline-flex items-center gap-2 mt-6 px-6 py-3 rounded-xl bg-brand-600 text-white
                         font-medium text-sm hover:bg-brand-700 transition no-underline">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" viewBox="0 0 24 24">
                <path d="M12 5v14M5 12h14" />
              </svg>
              去发布招募帖
            </Link>
          </div>
        )}

        {cards.length > 0 && (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {cards.map((card) => (<RecruitmentCard key={card.id} card={card} />))}
            </div>
            <div className="mt-8 text-center">
              {loading && cards.length > 0 && (
                <div className="flex items-center justify-center gap-2 text-text-muted text-sm">
                  <svg className="animate-spin w-4 h-4" viewBox="0 0 24 24" fill="none">
                    <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" strokeDasharray="32" strokeLinecap="round" />
                  </svg>
                  加载中...
                </div>
              )}
              {!loading && hasMore && (
                <button onClick={handleLoadMore}
                  className="px-6 py-2.5 rounded-xl border border-border bg-white text-text-secondary text-sm font-medium
                             hover:border-brand-200 hover:text-brand-600 transition cursor-pointer">
                  加载更多
                </button>
              )}
              {!loading && !hasMore && total > 0 && (
                <p className="text-sm text-text-muted">— 已展示全部 {total} 位伙伴 —</p>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default FindPeople;
