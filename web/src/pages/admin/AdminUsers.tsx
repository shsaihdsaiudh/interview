import { useEffect, useState } from 'react';
import { apiGet, apiPut } from '../../api/client';

interface AdminUser {
  email: string;
  nickname: string;
  student_id: string;
  role: string;
  account_status: string;
  created_at: string;
  department: string;
}

interface UserListResponse {
  users: AdminUser[];
  total: number;
  page: number;
  page_size: number;
}

const PAGE_SIZE = 20;

export default function AdminUsers() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actionMsg, setActionMsg] = useState('');

  const fetchUsers = (p: number, kw: string) => {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ page: String(p), page_size: String(PAGE_SIZE) });
    if (kw) params.set('keyword', kw);

    apiGet<UserListResponse>(`/admin/users?${params}`)
      .then((data) => {
        setUsers(data.users);
        setTotal(data.total);
        setLoading(false);
      })
      .catch(() => {
        setError('加载用户列表失败');
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchUsers(page, keyword);
  }, [page, keyword]);

  const handleSearch = () => {
    setKeyword(searchInput);
    setPage(1);
  };

  const handleBan = async (email: string) => {
    try {
      const updated = await apiPut<AdminUser>(`/admin/users/${encodeURIComponent(email)}/ban`);
      setUsers((prev) => prev.map((u) => (u.email === email ? updated : u)));
      setActionMsg(`已封禁 ${email}`);
      setTimeout(() => setActionMsg(''), 3000);
    } catch {
      setActionMsg('操作失败');
      setTimeout(() => setActionMsg(''), 3000);
    }
  };

  const handleUnban = async (email: string) => {
    try {
      const updated = await apiPut<AdminUser>(`/admin/users/${encodeURIComponent(email)}/unban`);
      setUsers((prev) => prev.map((u) => (u.email === email ? updated : u)));
      setActionMsg(`已解封 ${email}`);
      setTimeout(() => setActionMsg(''), 3000);
    } catch {
      setActionMsg('操作失败');
      setTimeout(() => setActionMsg(''), 3000);
    }
  };

  const isActive = (u: AdminUser) => u.account_status === 'active';
  const totalPages = Math.ceil(total / PAGE_SIZE);

  const statusBadge = (u: AdminUser) => {
    if (u.account_status === 'banned') {
      return (
        <span className="px-2 py-0.5 text-danger" style={{ fontSize: 13, background: 'rgba(224,112,112,0.1)', border: '1px solid rgba(224,112,112,0.25)' }}>
          已封禁
        </span>
      );
    }
    return (
      <span className="px-2 py-0.5 text-success" style={{ fontSize: 13, background: 'rgba(132,195,149,0.1)', border: '1px solid rgba(132,195,149,0.25)' }}>
        正常
      </span>
    );
  };

  return (
    <div>
      <h1 className="text-text font-bold mb-6" style={{ fontSize: 26 }}>用户管理</h1>

      {/* 消息提示 */}
      {actionMsg && (
        <div className="mb-4 px-3 py-2 text-text pixel-corners-sm" style={{ fontSize: 18, background: 'rgba(224,184,104,0.1)', border: '1px solid rgba(224,184,104,0.25)' }}>
          {actionMsg}
        </div>
      )}

      {/* 搜索栏 */}
      <div className="flex gap-2 mb-5">
        <input
          type="text"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
          placeholder="搜索邮箱 / 昵称 / 学号"
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
      ) : users.length === 0 ? (
        <p className="text-text-muted" style={{ fontSize: 18 }}>没有找到用户</p>
      ) : (
        <>
          {/* 表格 */}
          <div className="overflow-auto">
            <table className="w-full text-text" style={{ fontSize: 15, borderCollapse: 'collapse' }}>
              <thead>
                <tr className="text-text-muted" style={{ textAlign: 'left', borderBottom: '1px solid var(--color-border)' }}>
                  <th className="py-2 px-3 font-normal">昵称</th>
                  <th className="py-2 px-3 font-normal">邮箱</th>
                  <th className="py-2 px-3 font-normal">学号</th>
                  <th className="py-2 px-3 font-normal">院系</th>
                  <th className="py-2 px-3 font-normal">角色</th>
                  <th className="py-2 px-3 font-normal">状态</th>
                  <th className="py-2 px-3 font-normal">注册时间</th>
                  <th className="py-2 px-3 font-normal">操作</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.email} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td className="py-2 px-3 font-bold">{u.nickname}</td>
                    <td className="py-2 px-3 text-text-muted" style={{ fontSize: 14 }}>{u.email}</td>
                    <td className="py-2 px-3 text-text-muted">{u.student_id || '-'}</td>
                    <td className="py-2 px-3 text-text-muted">{u.department || '-'}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 ${u.role === 'admin' ? 'text-brand-600' : 'text-text-muted'}`}
                            style={{ fontSize: 13, background: u.role === 'admin' ? 'rgba(224,184,104,0.1)' : 'rgba(255,255,255,0.03)', border: `1px solid ${u.role === 'admin' ? 'rgba(224,184,104,0.25)' : 'var(--color-border)'}` }}>
                        {u.role === 'admin' ? '管理员' : '用户'}
                      </span>
                    </td>
                    <td className="py-2 px-3">{statusBadge(u)}</td>
                    <td className="py-2 px-3 text-text-muted" style={{ fontSize: 13 }}>
                      {new Date(u.created_at).toLocaleDateString('zh-CN')}
                    </td>
                    <td className="py-2 px-3">
                      {isActive(u) ? (
                        <button
                          onClick={() => handleBan(u.email)}
                          className="text-danger no-underline hover:underline"
                          style={{ fontSize: 15, background: 'none', border: 'none', cursor: 'pointer' }}
                        >
                          封禁
                        </button>
                      ) : (
                        <button
                          onClick={() => handleUnban(u.email)}
                          className="text-success no-underline hover:underline"
                          style={{ fontSize: 15, background: 'none', border: 'none', cursor: 'pointer' }}
                        >
                          解封
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* 分页 */}
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
            共 {total} 个用户
          </p>
        </>
      )}
    </div>
  );
}
