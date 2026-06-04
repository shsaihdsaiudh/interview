import { Link } from 'react-router-dom';
import { useEffect, useState, useRef } from 'react';
import { apiGet } from '../api/client';
import { getUser } from '../components/Navbar';

/* ═══════════════════════════════════════
   像素文字生成 — MATCH
   ═══════════════════════════════════════ */

const FONT: Record<string, string[]> = {
  M: ['10001','11011','10101','10101','10001','10001','10001'],
  A: ['01110','10001','10001','11111','10001','10001','10001'],
  T: ['11111','00100','00100','00100','00100','00100','00100'],
  C: ['01110','10001','10000','10000','10000','10001','01110'],
  H: ['10001','10001','10001','11111','10001','10001','10001'],
};
const WORD = 'MATCH';
const COLS = WORD.length * 5 + (WORD.length - 1);
const ROWS = 7;
const FLICKERS = ['f1','f2','f3'];

function PixelWord() {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    el.style.gridTemplateColumns = `repeat(${COLS}, 8px)`;
    el.style.gridTemplateRows = `repeat(${ROWS}, 8px)`;

    for (let row = 0; row < ROWS; row++) {
      for (let li = 0; li < WORD.length; li++) {
        const glyph = FONT[WORD[li]];
        for (let lc = 0; lc < 5; lc++) {
          const span = document.createElement('span');
          span.style.display = 'block';
          if (glyph[row][lc] === '1') {
            span.className = FLICKERS[Math.floor(Math.random() * FLICKERS.length)];
          }
          el.appendChild(span);
        }
        if (li < WORD.length - 1) {
          const gap = document.createElement('span');
          gap.style.display = 'block';
          el.appendChild(gap);
        }
      }
    }
  }, []);

  return (
    <div
      ref={ref}
      className="grid"
      style={{ gap: 2, opacity: 0.55, marginLeft: 'auto' }}
    />
  );
}

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
      <section className="px-6 pt-28 pb-20 md:pt-40 md:pb-28">
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center gap-3 mb-10 animate-fade-up">
            <span
              style={{
                display: 'inline-block',
                width: 5,
                height: 5,
                background: status ? 'var(--color-success)' : 'var(--color-warning)',
              }}
            />
            <span className="text-text-muted tracking-wider" style={{ fontSize: 13 }}>
              {status ? '系统运行中' : '连接中...'}
            </span>
            <span style={{ flex: 1, maxWidth: 40, height: 1, background: 'var(--color-border)' }} />
            <span className="text-text-muted" style={{ fontSize: 13 }}>
              1,284 人在线
            </span>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-16 items-start">
            <div className="lg:col-span-7 animate-fade-up" style={{ animationDelay: '0.1s' }}>
              <h1
                className="font-bold text-text leading-tight tracking-tight"
                style={{ fontSize: 'clamp(32px, 5vw, 52px)', fontFamily: 'var(--font-sans)' }}
              >
                和同学一起
                <br />
                <span className="text-brand-600">练习面试</span>
              </h1>

              <p className="text-text-secondary mt-6 max-w-lg" style={{ fontSize: 20, lineHeight: 1.7 }}>
                找到志同道合的伙伴，进行一对一模拟面试。
                在真实的对话中积累经验，在彼此的反馈中共同成长。
              </p>

              <div className="flex items-center gap-3 mt-8 animate-fade-up" style={{ animationDelay: '0.2s' }}>
                {user ? (
                  <>
                    <Link
                      to="/find"
                      className="pixel-btn primary no-underline"
                    >
                      开始寻找
                    </Link>
                    <Link
                      to="/appointments"
                      className="pixel-btn no-underline"
                    >
                      查看预约
                    </Link>
                  </>
                ) : (
                  <>
                    <Link
                      to="/register"
                      className="pixel-btn primary no-underline"
                    >
                      立即注册
                    </Link>
                    <Link
                      to="/login"
                      className="pixel-btn no-underline"
                    >
                      登录
                    </Link>
                  </>
                )}
              </div>
            </div>

            <div className="hidden lg:flex lg:col-span-5 justify-end">
              <PixelWord />
            </div>
          </div>
        </div>
      </section>

      {/* ── 分隔线 ── */}
      <div className="max-w-6xl mx-auto px-6">
        <hr style={{ border: 'none', borderTop: '1px solid var(--color-border)', margin: 0 }} />
      </div>

      {/* ── 功能入口 ── */}
      <section className="max-w-6xl mx-auto px-6 py-16">
        <div className="mb-8 animate-fade-up">
          <div className="flex items-center gap-2 text-text-muted" style={{ fontSize: 13, letterSpacing: '0.05em' }}>
            · menu ·
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          {[
            { title: '[1] 发现伙伴', desc: '按技能栈与目标公司精确筛选面试搭档', to: '/find', tag: 'react', tag2: 'go' },
            { title: '[2] 面试间', desc: '发布空闲时段，管理预约请求', to: '/appointments', tag: '可预约' },
            { title: '[3] 我的名片', desc: '创建招募卡片，让他人找到你', to: '/my-card', tag: '名片' },
            { title: '[4] 社区', desc: '分享面试经验，讨论求职策略', to: '/posts', tag: '帖子广场' },
          ].map((f) => (
            <Link
              key={f.to}
              to={f.to}
              className="no-underline text-inherit bg-card border border-border pixel-corners p-5 card-hover"
            >
              <h3
                className="font-bold text-text mb-2"
                style={{ fontSize: 17 }}
              >
                {f.title}
              </h3>
              <p className="text-text-secondary mb-3" style={{ fontSize: 15, lineHeight: 1.6 }}>
                {f.desc}
              </p>
              <span className="pixel-tag text-brand-600" style={{ borderColor: 'rgba(224,184,104,0.25)' }}>
                {f.tag}
              </span>
              {f.tag2 && (
                <span className="pixel-tag text-text-muted" style={{ marginLeft: 4 }}>
                  {f.tag2}
                </span>
              )}
            </Link>
          ))}
        </div>
      </section>

      {/* ── Footer ── */}
      <footer className="border-t border-border py-8 text-center">
        <p className="text-text-muted tracking-wider" style={{ fontSize: 13 }}>
          <span className="text-brand-600">mock·io</span>
          &nbsp;·&nbsp;面试互助平台&nbsp;·&nbsp;让每一次面试都成为成长的阶梯
        </p>
      </footer>
    </div>
  );
}

export default Home;
