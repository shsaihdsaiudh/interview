import { useEffect, useState } from 'react';
import { apiGet, apiPut } from '../api/client';
import { getUser, setUser, notifyAuthChange } from '../components/Navbar';

interface ProfileData {
  user: {
    email: string;
    nickname: string;
    student_id: string;
    department: string;
    tags: string[];
    avatar: string;
    contact_info: string;
    email_verified: boolean;
  };
}

function Settings() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [msg, setMsg] = useState('');

  const [nickname, setNickname] = useState('');
  const [studentId, setStudentId] = useState('');
  const [department, setDepartment] = useState('');
  const [tagsStr, setTagsStr] = useState('');
  const [avatar, setAvatar] = useState('');
  const [contactInfo, setContactInfo] = useState('');

  useEffect(() => {
    apiGet<ProfileData>('/profile')
      .then((res) => {
        setNickname(res.user.nickname);
        setStudentId(res.user.student_id);
        setDepartment(res.user.department || '');
        setTagsStr((res.user.tags || []).join(', '));
        setAvatar(res.user.avatar || '');
        setContactInfo(res.user.contact_info || '');
      })
      .catch(() => setError('加载资料失败'))
      .finally(() => setLoading(false));
  }, []);

  const showMsg = (text: string) => {
    setMsg(text);
    setTimeout(() => setMsg(''), 3000);
  };

  const handleSave = async () => {
    const tags = tagsStr.split(/[,，]/).map((t) => t.trim()).filter(Boolean);
    try {
      await apiPut('/profile', {
        nickname,
        student_id: studentId,
        department,
        tags,
        avatar,
        contact_info: contactInfo,
      });
      const u = getUser();
      if (u) { setUser({ ...u, nickname }); notifyAuthChange(); }
      showMsg('保存成功');
    } catch (err: any) {
      showMsg(err?.response?.data?.error || '保存失败');
    }
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto px-6 py-10">
        <div className="skeleton h-96 rounded-2xl" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-2xl mx-auto px-6 py-10">
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">{error}</div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto px-6 py-10">
      <h1 className="text-2xl font-bold text-text mb-8">个人设置</h1>

      {msg && (
        <div
          className={`px-4 py-2.5 rounded-xl text-sm mb-4 border ${
            msg === '保存成功'
              ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
              : 'bg-red-50 border-red-200 text-red-600'
          }`}
        >
          {msg}
        </div>
      )}

      <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
        <h2 className="text-lg font-bold text-text mb-5">个人资料</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">昵称</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={nickname} onChange={(e) => setNickname(e.target.value)} />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">学号</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={studentId} onChange={(e) => setStudentId(e.target.value)} />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">院系</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={department} onChange={(e) => setDepartment(e.target.value)} placeholder="如：计算机学院" />
          </label>
          <label className="flex flex-col gap-1.5">
            <span className="text-sm font-medium text-text-secondary">面试方向标签</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={tagsStr} onChange={(e) => setTagsStr(e.target.value)} placeholder="产品, 前端, 后端" />
          </label>
          <label className="flex flex-col gap-1.5 sm:col-span-2">
            <span className="text-sm font-medium text-text-secondary">头像 URL</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={avatar} onChange={(e) => setAvatar(e.target.value)} placeholder="https://..." />
          </label>
          <label className="flex flex-col gap-1.5 sm:col-span-2">
            <span className="text-sm font-medium text-text-secondary">联系方式</span>
            <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={contactInfo} onChange={(e) => setContactInfo(e.target.value)} placeholder="微信 / QQ / 手机" />
          </label>
        </div>
        <button
          onClick={handleSave}
          className="px-5 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none"
        >
          保存
        </button>
      </div>
    </div>
  );
}

export default Settings;
