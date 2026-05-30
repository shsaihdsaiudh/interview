import { Link } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { apiGet } from '../api/client';
import { getUser } from '../components/Navbar';

function Home() {
  const [status, setStatus] = useState<string>('');
  const user = getUser();

  useEffect(() => {
    apiGet<{ message: string }>('/ping')
      .then((res) => setStatus(res.message))
      .catch(() => setStatus('后端未连接'));
  }, []);

  return (
    <div>
      {/* ── Hero ── */}
      <section className="px-6 pt-32 pb-24 md:pt-40 md:pb-32">
        <div className="max-w-3xl mx-auto">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-border bg-surface-alt text-xs text-text-secondary mb-8">
            <span className={`w-1.5 h-1.5 rounded-full ${status ? 'bg-success' : 'bg-warning'}`} />
            {status ? '系统运行正常' : '正在连接...'}
          </div>

          <h1 className="text-5xl md:text-6xl font-bold text-text leading-tight tracking-tight">
            和同学一起
            <br />
            练习面试
          </h1>

          <p className="text-lg text-text-secondary mt-6 max-w-xl leading-relaxed">
            找到志同道合的同学，进行一对一模拟面试。在真实的对话中积累经验，在彼此的反馈中成长。
          </p>

          <div className="flex items-center gap-3 mt-10">
            {user ? (
              <>
                <Link
                  to="/find"
                  className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-brand-600 text-white font-medium text-sm
                             hover:bg-brand-700 transition no-underline"
                >
                  找人面试
                </Link>
                <Link
                  to="/appointments"
                  className="inline-flex items-center gap-2 px-6 py-3 rounded-xl border border-border text-text-secondary font-medium text-sm
                             hover:border-brand-200 hover:text-brand-600 transition no-underline bg-white"
                >
                  查看预约
                </Link>
              </>
            ) : (
              <>
                <Link
                  to="/register"
                  className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-brand-600 text-white font-medium text-sm
                             hover:bg-brand-700 transition no-underline"
                >
                  立即注册
                </Link>
                <Link
                  to="/login"
                  className="inline-flex items-center gap-2 px-6 py-3 rounded-xl border border-border text-text-secondary font-medium text-sm
                             hover:border-brand-200 hover:text-brand-600 transition no-underline bg-white"
                >
                  登录
                </Link>
              </>
            )}
          </div>
        </div>
      </section>

      {/* ── 功能介绍 ── */}
      <section className="max-w-5xl mx-auto px-6 pb-24">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { title: '找人面试', desc: '浏览可预约的面试官，找到方向匹配的同学', to: '/find' },
            { title: '管理预约', desc: '查看收到的和发出的面试预约，一键接受或拒绝', to: '/appointments' },
            { title: '设置时间', desc: '编辑个人资料，管理空闲时间段方便他人预约', to: '/settings/availability' },
            { title: '帖子广场', desc: '分享面试经验，讨论求职话题', to: '/posts' },
          ].map((f) => (
            <Link
              key={f.to}
              to={f.to}
              className="group bg-card rounded-2xl border border-border p-6 no-underline text-inherit
                         hover:border-brand-200 hover:shadow-sm transition"
            >
              <h3 className="font-semibold text-text group-hover:text-brand-600 transition-colors">
                {f.title}
              </h3>
              <p className="text-sm text-text-secondary mt-1.5 leading-relaxed">{f.desc}</p>
            </Link>
          ))}
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="border-t border-border py-8 text-center text-xs text-text-muted">
        面试互助平台 · 让每一次模拟面试都成为成长的阶梯
      </footer>
    </div>
  );
}

export default Home;
