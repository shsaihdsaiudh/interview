import { useEffect, useState } from 'react';
import { apiGet, apiDelete } from '../../api/client';

interface CardData {
  id: string;
  user_id: string;
  nickname: string;
  skills: string[];
  target_companies: string[];
  role: string;
  experience_years: number;
  bio: string;
  is_active: boolean;
  created_at: string;
}

interface CardListResponse {
  cards: CardData[];
  total: number;
  page: number;
  page_size: number;
}

const PAGE_SIZE = 20;
const ROLE_MAP: Record<string, string> = {
  interviewer: '面试官',
  interviewee: '面试者',
  both: '两者皆可',
};

export default function AdminCards() {
  const [cards, setCards] = useState<CardData[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  const fetchCards = (p: number, kw: string) => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ page: String(p), page_size: String(PAGE_SIZE) });
    if (kw) params.set('keyword', kw);

    apiGet<CardListResponse>(`/admin/cards?${params}`)
      .then((data) => {
        setCards(data.cards);
        setTotal(data.total);
        setLoading(false);
      })
      .catch(() => {
        setError('加载名片列表失败');
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchCards(page, keyword);
  }, [page, keyword]);

  const handleSearch = () => {
    setKeyword(searchInput);
    setPage(1);
  };

  const handleDelete = async (id: string, nickname: string) => {
    if (!confirm(`确定删除 "${nickname}" 的名片吗？此操作不可撤销。`)) return;
    try {
      await apiDelete(`/admin/cards/${encodeURIComponent(id)}`);
      setCards((prev) => prev.filter((c) => c.id !== id));
      setTotal((t) => t - 1);
      setActionMsg('名片已删除');
      setTimeout(() => setActionMsg(''), 3000);
    } catch {
      setActionMsg('删除失败');
      setTimeout(() => setActionMsg(''), 3000);
    }
  };

  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div>
      <h1 className="text-text font-bold mb-6" style={{ fontSize: 26 }}>名片管理</h1>

      {actionMsg && (
        <div className="mb-4 px-3 py-2 text-text pixel-corners-sm" style={{ fontSize: 18, background: 'rgba(224,184,104,0.1)', border: '1px solid rgba(224,184,104,0.25)' }}>
          {actionMsg}
        </div>
      )}

      <div className="flex gap-2 mb-5">
        <input
          type="text"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          placeholder="搜索用户昵称 / 邮箱"
          className="flex-1 px-3 py-2 bg-surface border border-border text-text pixel-corners-sm"
          style={{ fontSize: 17, maxWidth: 360 }}
        />
        <button onClick={handleSearch} className="pixel-btn primary" style={{ fontSize: 17 }}>
          搜索
        </button>
      </div>

      {error && (
        <div className="mb-5 px-3 py-2 text-danger pixel-corners-sm"
             style={{ fontSize: 17, background: 'rgba(224,112,112,0.08)', border: '1px solid rgba(224,112,112,0.2)' }}>
          {error}
        </div>
      )}

      {loading ? (
        <p className="text-text-muted" style={{ fontSize: 18 }}>加载中...</p>
      ) : cards.length === 0 ? (
        <p className="text-text-muted" style={{ fontSize: 18 }}>没有找到名片</p>
      ) : (
        <>
          <div className="space-y-3">
            {cards.map((card) => (
              <div
                key={card.id}
                className="bg-card border border-border pixel-corners p-4"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="font-bold text-text" style={{ fontSize: 18 }}>{card.nickname}</span>
                      <span className="text-text-muted" style={{ fontSize: 14 }}>{card.user_id}</span>
                      <span className="px-2 py-0.5 text-text-muted" style={{ fontSize: 13, background: 'rgba(255,255,255,0.03)', border: '1px solid var(--color-border)' }}>
                        {ROLE_MAP[card.role] || card.role}
                      </span>
                      {!card.is_active && (
                        <span className="px-2 py-0.5 text-danger" style={{ fontSize: 13, background: 'rgba(224,112,112,0.1)', border: '1px solid rgba(224,112,112,0.25)' }}>
                          已下架
                        </span>
                      )}
                    </div>

                    {card.skills?.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 mb-2">
                        {card.skills.map((s) => (
                          <span key={s} className="pixel-tag text-brand-600" style={{ borderColor: 'rgba(224,184,104,0.25)', fontSize: 13 }}>
                            {s}
                          </span>
                        ))}
                      </div>
                    )}

                    {card.target_companies?.length > 0 && (
                      <div className="flex flex-wrap gap-1.5 mb-2">
                        {card.target_companies.map((c) => (
                          <span key={c} className="pixel-tag text-text-muted" style={{ fontSize: 13 }}>
                            🎯 {c}
                          </span>
                        ))}
                      </div>
                    )}

                    {card.bio && (
                      <p className="text-text-secondary mt-1" style={{ fontSize: 15, lineHeight: 1.5 }}>
                        {card.bio.length > 120 ? card.bio.slice(0, 120) + '...' : card.bio}
                      </p>
                    )}
                  </div>

                  <button
                    onClick={() => handleDelete(card.id, card.nickname)}
                    className="text-danger no-underline hover:underline flex-shrink-0"
                    style={{ fontSize: 15, background: 'none', border: 'none', cursor: 'pointer' }}
                  >
                    删除
                  </button>
                </div>
              </div>
            ))}
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
            共 {total} 张名片
          </p>
        </>
      )}
    </div>
  );
}
