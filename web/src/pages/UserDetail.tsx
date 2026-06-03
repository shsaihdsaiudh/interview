import { useEffect, useRef, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiGet, apiPost, apiPut, apiDelete, apiUpload, getApiErrorMessage, removeToken } from '../api/client';
import { getUser, setUser, notifyAuthChange, clearUser } from '../components/Navbar';

interface UserInfo {
  email: string;
  nickname: string;
  student_id: string;
  department: string;
  tags: string[];
  avatar: string;
  contact_info: string;
  email_verified: boolean;
  account_status: string;
}

interface TimeSlot {
  id: string;
  user_id: string;
  date: string;
  start_time: string;
  end_time: string;
}

interface DetailData {
  user: UserInfo;
  availabilities: TimeSlot[];
}

const dayNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];
function getDayOfWeek(dateStr: string): string {
  return dayNames[new Date(dateStr).getDay()];
}

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4'];

function UserDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const currentUser = getUser();

  const [detail, setDetail] = useState<DetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [bookingId, setBookingId] = useState<string | null>(null);
  const [message, setMessage] = useState('');
  const [bookingError, setBookingError] = useState('');
  const [bookingSuccess, setBookingSuccess] = useState('');

  // ── edit mode state ──
  const [editNickname, setEditNickname] = useState('');
  const [editStudentId, setEditStudentId] = useState('');
  const [editDepartment, setEditDepartment] = useState('');
  const [editTagsStr, setEditTagsStr] = useState('');
  const [editAvatar, setEditAvatar] = useState('');
  const [editContactInfo, setEditContactInfo] = useState('');
  const [saving, setSaving] = useState(false);

  // ── toast ──
  const [toast, setToast] = useState('');
  const [toastType, setToastType] = useState<'success' | 'error'>('success');

  // ── avatar upload ──
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [avatarUploading, setAvatarUploading] = useState(false);
  const [avatarPreview, setAvatarPreview] = useState('');

  // ── change password ──
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [pwdSaving, setPwdSaving] = useState(false);
  const [pwdMsg, setPwdMsg] = useState('');
  const [pwdMsgType, setPwdMsgType] = useState<'success' | 'error'>('success');

  // ── delete account ──
  const [deletePassword, setDeletePassword] = useState('');
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [deleteSaving, setDeleteSaving] = useState(false);
  const [deleteMsg, setDeleteMsg] = useState('');
  const [deleteMsgType, setDeleteMsgType] = useState<'success' | 'error'>('error');

  useEffect(() => {
    if (!id) return;
    apiGet<DetailData>(`/users/${id}`)
      .then(setDetail)
      .catch(() => setError('加载用户详情失败'))
      .finally(() => setLoading(false));
  }, [id]);

  const handleBook = async () => {
    if (!bookingId) return;
    if (!currentUser) {
      navigate('/login');
      return;
    }
    setBookingError('');
    setBookingSuccess('');
    try {
      await apiPost('/appointments', {
        time_slot_id: bookingId,
        message: message || '希望预约一场模拟面试',
      });
      setBookingSuccess('预约成功，请等待对方确认');
      setBookingId(null);
      setMessage('');
    } catch (err: unknown) {
      setBookingError(getApiErrorMessage(err, '预约失败'));
    }
  };

  // ── init edit fields ──
  useEffect(() => {
    if (detail) {
      setEditNickname(detail.user.nickname);
      setEditStudentId(detail.user.student_id);
      setEditDepartment(detail.user.department || '');
      setEditTagsStr((detail.user.tags || []).join(', '));
      setEditAvatar(detail.user.avatar || '');
      setEditContactInfo(detail.user.contact_info || '');
    }
  }, [detail]);

  // ── toast helpers ──
  const showToast = (text: string, type: 'success' | 'error' = 'success') => {
    setToast(text);
    setToastType(type);
    setTimeout(() => setToast(''), 3000);
  };

  const showPwdToast = (text: string, type: 'success' | 'error' = 'success') => {
    setPwdMsg(text);
    setPwdMsgType(type);
    setTimeout(() => setPwdMsg(''), 3000);
  };

  const showDeleteToast = (text: string, type: 'success' | 'error' = 'error') => {
    setDeleteMsg(text);
    setDeleteMsgType(type);
    setTimeout(() => setDeleteMsg(''), 5000);
  };

  // ── save profile ──
  const handleSaveProfile = async () => {
    const tags = editTagsStr.split(/[,，]/).map((t) => t.trim()).filter(Boolean);
    setSaving(true);
    try {
      await apiPut('/profile', {
        nickname: editNickname,
        student_id: editStudentId,
        department: editDepartment,
        tags,
        avatar: editAvatar,
        contact_info: editContactInfo,
      });
      const u = getUser();
      if (u) { setUser({ ...u, nickname: editNickname }); notifyAuthChange(); }
      showToast('保存成功');
      apiGet<DetailData>(`/users/${id}`).then(setDetail);
    } catch (err: unknown) {
      showToast(getApiErrorMessage(err, '保存失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  // ── avatar upload ──
  const handleAvatarUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const previewURL = URL.createObjectURL(file);
    setAvatarPreview(previewURL);

    const allowedTypes = ['image/jpeg', 'image/png', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      showToast('仅支持 JPEG、PNG、WebP 格式', 'error');
      return;
    }
    if (file.size > 2 * 1024 * 1024) {
      showToast('头像文件大小不能超过 2MB', 'error');
      return;
    }

    setAvatarUploading(true);
    try {
      const res = await apiUpload<{ avatar_url: string }>('/profile/avatar', file);
      setEditAvatar(res.avatar_url);
      showToast('头像上传成功');
    } catch (err: unknown) {
      showToast(getApiErrorMessage(err, '上传失败'), 'error');
      URL.revokeObjectURL(previewURL);
      setAvatarPreview('');
    } finally {
      setAvatarUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  // ── change password ──
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPwdMsg('');

    if (!oldPassword) { showPwdToast('请输入旧密码', 'error'); return; }
    if (!newPassword) { showPwdToast('请输入新密码', 'error'); return; }
    if (newPassword.length < 6) { showPwdToast('新密码至少 6 位', 'error'); return; }
    if (newPassword !== confirmPassword) { showPwdToast('两次输入的新密码不一致', 'error'); return; }

    setPwdSaving(true);
    try {
      await apiPut('/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword,
      });
      showPwdToast('密码修改成功');
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: unknown) {
      showPwdToast(getApiErrorMessage(err, '修改密码失败'), 'error');
    } finally {
      setPwdSaving(false);
    }
  };

  // ── delete account ──
  const handleDeleteInitiate = () => {
    if (!deletePassword) {
      showDeleteToast('请输入密码确认身份', 'error');
      return;
    }
    setDeleteConfirm(true);
  };

  const handleDeleteAccount = async () => {
    setDeleteSaving(true);
    setDeleteMsg('');
    try {
      await apiDelete(`/auth/account`, { password: deletePassword });
      removeToken();
      clearUser();
      notifyAuthChange();
      navigate('/', { replace: true });
    } catch (err: unknown) {
      showDeleteToast(getApiErrorMessage(err, '注销失败'));
      setDeleteConfirm(false);
    } finally {
      setDeleteSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto px-6 py-10">
        <div className="skeleton h-48 rounded-2xl mb-4" />
        <div className="skeleton h-64 rounded-2xl" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-3xl mx-auto px-6 py-10">
        <div className="bg-red-50 border border-red-200 rounded-2xl p-8 text-center text-red-600">
          {error}
        </div>
      </div>
    );
  }

  if (!detail) {
    return (
      <div className="max-w-3xl mx-auto px-6 py-10 text-center text-text-muted">
        <p className="text-lg">用户不存在</p>
      </div>
    );
  }

  const { user, availabilities } = detail;
  const isSelf = currentUser?.email === user.email;
  const avatarBg = avatarColors[user.nickname.charCodeAt(0) % avatarColors.length];

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      {/* Toast */}
      {toast && (
        <div
          className={`px-4 py-2.5 rounded-xl text-sm mb-4 border ${
            toastType === 'success'
              ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
              : 'bg-red-50 border-red-200 text-red-600'
          }`}
        >
          {toast}
        </div>
      )}

      {/* user info card */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8 mb-4">
        <div className="flex items-start gap-4">
          {/* avatar area */}
          {isSelf ? (
            <div className="flex-shrink-0">
              {(avatarPreview || editAvatar) ? (
                <img
                  src={avatarPreview || editAvatar}
                  alt="头像预览"
                  className="w-16 h-16 rounded-full object-cover border border-border"
                />
              ) : (
                <div
                  className="w-16 h-16 rounded-full flex items-center justify-center text-white text-2xl font-bold"
                  style={{ background: avatarBg }}
                >
                  {user.nickname.charAt(0)}
                </div>
              )}
            </div>
          ) : (
            <>
              {user.avatar ? (
                <img src={user.avatar} alt={user.nickname} className="w-16 h-16 rounded-full object-cover" />
              ) : (
                <div
                  className="w-16 h-16 rounded-full flex items-center justify-center text-white text-2xl font-bold flex-shrink-0"
                  style={{ background: avatarBg }}
                >
                  {user.nickname.charAt(0)}
                </div>
              )}
            </>
          )}

          <div className="flex-1 min-w-0">
            {isSelf ? (
              /* edit mode */
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <label className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-text-secondary">昵称</span>
                  <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={editNickname} onChange={(e) => setEditNickname(e.target.value)} />
                </label>
                <label className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-text-secondary">学号</span>
                  <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={editStudentId} onChange={(e) => setEditStudentId(e.target.value)} />
                </label>
                <label className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-text-secondary">院系</span>
                  <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={editDepartment} onChange={(e) => setEditDepartment(e.target.value)} placeholder="如：计算机学院" />
                </label>
                <label className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-text-secondary">面试方向标签</span>
                  <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={editTagsStr} onChange={(e) => setEditTagsStr(e.target.value)} placeholder="产品, 前端, 后端" />
                </label>
                <label className="flex flex-col gap-1 sm:col-span-2">
                  <span className="text-xs font-medium text-text-secondary">头像 URL</span>
                  <div className="flex items-center gap-2 flex-wrap">
                    <input className="flex-1 min-w-[150px] px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={editAvatar} onChange={(e) => setEditAvatar(e.target.value)} placeholder="https://...或上传头像" />
                    <input ref={fileInputRef} type="file" accept="image/jpeg,image/png,image/webp" onChange={handleAvatarUpload} className="hidden" />
                    <button type="button" onClick={() => fileInputRef.current?.click()} disabled={avatarUploading} className="px-3 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-xs font-medium transition cursor-pointer border-none disabled:opacity-50 whitespace-nowrap">
                      {avatarUploading ? '上传中...' : '上传头像'}
                    </button>
                  </div>
                </label>
                <label className="flex flex-col gap-1 sm:col-span-2">
                  <span className="text-xs font-medium text-text-secondary">联系方式</span>
                  <input className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt" value={editContactInfo} onChange={(e) => setEditContactInfo(e.target.value)} placeholder="微信 / QQ / 手机" />
                </label>
              </div>
            ) : (
              /* read-only mode */
              <>
                <h1 className="text-2xl font-bold text-text">{user.nickname}</h1>
                <p className="text-text-secondary text-sm mt-1">
                  {user.department || '未设置院系'} · {user.student_id}
                </p>
                {user.tags && user.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1.5 mt-2">
                    {user.tags.map((t) => (
                      <span key={t} className="px-2 py-0.5 rounded-md bg-brand-50 text-brand-700 text-xs font-medium">
                        {t}
                      </span>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>
        </div>

        {/* contact (read-only mode) */}
        {!isSelf && user.contact_info && (
          <div className="mt-4 p-3 rounded-xl bg-gray-50 border border-border text-sm text-text-secondary">
            联系方式：{user.contact_info}
          </div>
        )}

        {/* save button */}
        {isSelf && (
          <button
            onClick={handleSaveProfile}
            disabled={saving}
            className="mt-5 px-5 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50"
          >
            {saving ? '保存中...' : '保存资料'}
          </button>
        )}
      </div>

      {/* change password (self only) */}
      {isSelf && (
        <div className="bg-card rounded-2xl border border-border shadow-sm p-8 mb-4">
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
      )}

      {/* delete account (self only) */}
      {isSelf && (
        <div className="bg-card rounded-2xl border border-danger/30 shadow-sm p-8 mb-4">
          <h2 className="text-lg font-bold text-danger mb-1">注销账号</h2>
          <p className="text-sm text-text-secondary mb-5">此操作不可撤销，所有数据将被永久删除。</p>

          {deleteMsg && (
            <div
              className={`px-4 py-2.5 rounded-xl text-sm mb-4 border ${
                deleteMsgType === 'success'
                  ? 'bg-emerald-50 border-emerald-200 text-emerald-700'
                  : 'bg-red-50 border-red-200 text-red-600'
              }`}
            >
              {deleteMsg}
            </div>
          )}

          {!deleteConfirm ? (
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
              <input type="password" className="px-3 py-2 rounded-lg border border-border text-sm bg-surface-alt w-full sm:w-64" value={deletePassword} onChange={(e) => setDeletePassword(e.target.value)} placeholder="输入密码确认身份" disabled={deleteSaving} />
              <button onClick={handleDeleteInitiate} disabled={deleteSaving} className="px-5 py-2 rounded-lg bg-danger hover:bg-red-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50 whitespace-nowrap">
                {deleteSaving ? '处理中...' : '注销账号'}
              </button>
            </div>
          ) : (
            <div className="rounded-xl border border-danger/20 bg-red-50 p-5">
              <p className="text-sm text-red-700 font-medium mb-1">⚠️ 确认注销账号？</p>
              <p className="text-sm text-red-600 mb-4">此操作不可撤销，你的个人资料、空闲时间和预约记录将被永久删除。</p>
              <div className="flex gap-3">
                <button onClick={handleDeleteAccount} disabled={deleteSaving} className="px-5 py-2 rounded-lg bg-danger hover:bg-red-700 text-white text-sm font-medium transition cursor-pointer border-none disabled:opacity-50">
                  {deleteSaving ? '注销中...' : '确认注销'}
                </button>
                <button onClick={() => { setDeleteConfirm(false); setDeletePassword(''); }} disabled={deleteSaving} className="px-5 py-2 rounded-lg border border-border text-text-secondary hover:text-text text-sm font-medium transition cursor-pointer bg-transparent disabled:opacity-50">
                  取消
                </button>
              </div>
            </div>
          )}
        </div>
      )}

      {/* availabilities */}
      <div className="bg-card rounded-2xl border border-border shadow-sm p-8">
        <h2 className="text-lg font-bold text-text mb-4">空闲时间</h2>

        {availabilities.length === 0 ? (
          <div className="text-center py-12 text-text-muted">
            <p className="font-medium">暂无空闲时间</p>
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {availabilities.map((slot) => {
              const selected = bookingId === slot.id;
              return (
                <div
                  key={slot.id}
                  onClick={() => { if (!isSelf) setBookingId(selected ? null : slot.id); }}
                  className={`rounded-xl p-4 border transition ${
                    selected
                      ? 'border-brand-400 bg-brand-50/50'
                      : 'border-border bg-gray-50/50 hover:border-gray-300'
                  } ${!isSelf ? 'cursor-pointer' : ''}`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-semibold text-text text-sm">
                        {slot.date} <span className="text-text-muted font-normal">{getDayOfWeek(slot.date)}</span>
                      </div>
                      <div className="text-sm text-text-secondary mt-0.5">
                        {slot.start_time} - {slot.end_time}
                      </div>
                    </div>
                    {!isSelf && !selected && (
                      <span className="text-xs text-text-muted">选择</span>
                    )}
                    {selected && (
                      <span className="text-xs text-brand-600 font-medium">已选中</span>
                    )}
                  </div>

                  {selected && (
                    <div className="mt-3 pt-3 border-t border-brand-200">
                      <textarea
                        placeholder="附言：简单介绍一下你想练习的方向..."
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        className="w-full px-3 py-2 rounded-lg border border-border text-sm resize-y min-h-[72px]"
                        rows={3}
                      />
                      <div className="flex items-center gap-2 mt-2">
                        <button
                          onClick={handleBook}
                          className="px-4 py-2 rounded-lg bg-brand-600 hover:bg-brand-700 text-white text-sm font-medium
                                     transition cursor-pointer border-none"
                        >
                          发起预约
                        </button>
                        <button
                          onClick={() => setBookingId(null)}
                          className="px-4 py-2 rounded-lg border border-border bg-white text-text-secondary text-sm
                                     font-medium hover:bg-gray-50 transition cursor-pointer"
                        >
                          取消
                        </button>
                        {bookingError && <span className="text-sm text-danger">{bookingError}</span>}
                        {bookingSuccess && <span className="text-sm text-success">{bookingSuccess}</span>}
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}

        {isSelf && (
          <button
            onClick={() => navigate('/appointments')}
            className="mt-4 w-full py-2.5 rounded-xl border border-dashed border-border text-text-secondary text-sm
                       font-medium hover:border-brand-300 hover:text-brand-600 transition cursor-pointer bg-transparent"
          >
            管理我的空闲时间
          </button>
        )}
      </div>
    </div>
  );
}

export default UserDetail;
