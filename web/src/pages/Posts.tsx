function Posts() {
  return (
    <div className="max-w-3xl mx-auto px-6 py-16">
      <div className="text-center animate-fade-up">
        <h1 className="text-text mb-2" style={{ fontSize: 24, fontWeight: 700 }}>feed</h1>
        <p className="text-text-secondary mb-8" style={{ fontSize: 18 }}>
          发布你的面试需求，或者浏览其他人的帖子
        </p>
        <div className="p-8 pixel-corners" style={{ background: 'rgba(212,160,64,0.06)', border: '1px solid rgba(212,160,64,0.2)' }}>
          <span className="text-warning mb-3" style={{ fontSize: 32 }}>[ ]</span>
          <p className="text-text" style={{ fontSize: 18, fontWeight: 700 }}>帖子功能即将上线</p>
          <p className="text-text-muted mt-2" style={{ fontSize: 18 }}>敬请期待...</p>
        </div>
      </div>
    </div>
  );
}

export default Posts;
