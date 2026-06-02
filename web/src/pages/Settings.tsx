import { useEffect, useState } from 'react';
import { apiGet, apiPut, getApiErrorMessage } from '../api/client';
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
    account_status: string;
  };
}

function Settings() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [msg, setMsg] = useState('');
  const [msgType, setMsgType] = useState<'success' | 'error'>('success');

  const [nickname, setNickname] = useState('');
  const [studentId, setStudentId] = useState('');
  const [department, setDepartment] = useState('');
  const [tagsStr, setTagsStr] = useState('');
  const [avatar, setAvatar] = useState('');
  const [contactInfo, setContactInfo] = useState('');

  // ── 修改密码 ──
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [pwdSaving, setPwdSaving] = useState(false);
  const [pwdMsg, setPwdMsg] = useState('');
  const [pwdMsgType, setPwdMsgType] = useState<'success' | 'error'>('success');

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

  const showMsg = (text: string, type: 'success' | 'error' = 'success') => {
    setMsg(text);
    setMsgType(type);
    setTimeout(() => setMsg(''), 3000);
  };

  const showPwdMsg = (text: string, type: 'success' | 'error' = 'success') => {
    setPwdMsg(text);
    setPwdMsgType(type);
    setTimeout(() => setPwdMsg(''), 3000);
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPwdMsg('');

    if (!oldPassword) { showPwdMsg('请输入旧密码', 'error'); return; }
    if (!newPassword) { showPwdMsg('请输入新密码', 'error'); return; }
    if (newPassword.length < 6) { showPwdMsg('新密码至少 6 位', 'error'); return; }
    if (newPassword !== confirmPassword) { showPwdMsg('两次输入的新密码不一致', 'error'); return; }

    setPwdSaving(true);
    try {
      await apiPut('/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword,
      });
      showPwdMsg('密码修改成功', 'success');
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: unknown) {
      showPwdMsg(getApiErrorMessage(err, '修改密码失败'), 'error');
    } finally {
      setPwdSaving(false);
    }
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
    } catch (err: unknown) {
      showMsg(getApiErrorMessage(err, '保存失败'));
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
            msgType === 'success'
              ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
              : 'bg-red-50 border-red-200 text-red-600'
          }`}
        >
          {msg}
        </div>
      )}

      <div className="bg-card rounded-2xl border border-border shadow-sm p-8 mb-6">
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

      {/* ── 修改密码 ── */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
        <h2 className="text-lg font-bold text-text mb-5">修改密码</h2>

        {pwdMsg && (
          <div
            className={`px-4 py-2.5 rounded-xl text-sm mb-4 border ${
              pwdMsgType === 'success'
                ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
                : 'bg-red-50 border-red-200 text-red-600'
            }`}
          >
            {pwdMsg}
          </div>
        )}

        <form onSubmit={handleChangePassword}>
          <div className="grid grid-cols-1 gap-4 mb-6">
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">旧密码</span>
              <input type="password" className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} placeholder="输入当前密码" disabled={pwdSaving} />
            </label>
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">新密码</span>
              <input type="password" className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} placeholder="至少 6 位" disabled={pwdSaving} />
            </label>
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium text-text-secondary">确认新密码</span>
              <input type="password" className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} placeholder="再次输入新密码" disabled={pwdSaving} />
            </label>
          </div>
          <button type="submit" disabled={pwdSaving} className="px-5 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50">
            {pwdSaving ? '修改中...' : '修改密码'}
          </button>
        </form>
      </div>
    </div>
  );
}

export default Settings;
