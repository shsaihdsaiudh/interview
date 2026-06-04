import { useEffect, useRef, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiGet, apiPost, apiPut, apiDelete, apiUpload, getApiErrorMessage, removeToken } from '../api/client';
import { getUser, setUser, notifyAuthChange, clearUser } from '../components/Navbar';
import WeekCalendar from '../components/WeekCalendar';

interface UserInfo { email: string; nickname: string; student_id: string; department: string; tags: string[]; avatar: string; contact_info: string; email_verified: boolean; account_status: string; }
interface TimeSlot { id: string; user_id: string; date: string; start_time: string; end_time: string; }
interface DetailData { user: UserInfo; availabilities: TimeSlot[]; }

const avatarColors = ['#e0b868','#78b880','#d4a040','#e07070','#aaaab2','#72727c'];

function UserDetail() {
  const { id } = useParams<{ id: string }>(); const navigate = useNavigate(); const currentUser = getUser();
  const [detail, setDetail] = useState<DetailData | null>(null);
  const [loading, setLoading] = useState(true); const [error, setError] = useState('');
  const [editNickname, setEditNickname] = useState(''); const [editStudentId, setEditStudentId] = useState('');
  const [editDepartment, setEditDepartment] = useState(''); const [editTagsStr, setEditTagsStr] = useState('');
  const [editAvatar, setEditAvatar] = useState(''); const [editContactInfo, setEditContactInfo] = useState('');
  const [saving, setSaving] = useState(false); const [toast, setToast] = useState(''); const [toastType, setToastType] = useState<'success'|'error'>('success');
  const fileRef = useRef<HTMLInputElement>(null); const [avatarUploading, setAvatarUploading] = useState(false);
  const [avatarPreview, setAvatarPreview] = useState('');
  const [oldPwd, setOldPwd] = useState(''); const [newPwd, setNewPwd] = useState(''); const [cfmPwd, setCfmPwd] = useState('');
  const [pwdSaving, setPwdSaving] = useState(false); const [pwdMsg, setPwdMsg] = useState(''); const [pwdType, setPwdType] = useState<'success'|'error'>('success');
  const [delPwd, setDelPwd] = useState(''); const [delConfirm, setDelConfirm] = useState(false);
  const [delSaving, setDelSaving] = useState(false); const [delMsg, setDelMsg] = useState(''); const [delType, setDelType] = useState<'success'|'error'>('error');

  useEffect(() => { if (!id) return; apiGet<DetailData>(`/users/${id}`).then(setDetail).catch(() => setError('加载失败')).finally(() => setLoading(false)); }, [id]);
  const handleBook = async (slotId: string, message: string) => { await apiPost("/appointments", { time_slot_id: slotId, message }); };
  useEffect(() => { if (detail) { setEditNickname(detail.user.nickname); setEditStudentId(detail.user.student_id); setEditDepartment(detail.user.department||''); setEditTagsStr((detail.user.tags||[]).join(', ')); setEditAvatar(detail.user.avatar||''); setEditContactInfo(detail.user.contact_info||''); } }, [detail]);

  const showToast = (t: string, ty: 'success'|'error'='success') => { setToast(t); setToastType(ty); setTimeout(() => setToast(''), 3000); };
  const showPwd = (t: string, ty: 'success'|'error'='success') => { setPwdMsg(t); setPwdType(ty); setTimeout(() => setPwdMsg(''), 3000); };
  const showDel = (t: string, ty: 'success'|'error'='error') => { setDelMsg(t); setDelType(ty); setTimeout(() => setDelMsg(''), 5000); };

  const saveProfile = async () => { const tags = editTagsStr.split(/[,，]/).map((t)=>t.trim()).filter(Boolean); setSaving(true);
    try { await apiPut('/profile', { nickname: editNickname, student_id: editStudentId, department: editDepartment, tags, avatar: editAvatar, contact_info: editContactInfo });
      const u = getUser(); if (u) { setUser({...u, nickname: editNickname}); notifyAuthChange(); }
      showToast('已保存'); apiGet<DetailData>(`/users/${id}`).then(setDetail); }
    catch (err: unknown) { showToast(getApiErrorMessage(err, '失败'), 'error'); } finally { setSaving(false); } };

  const uploadAvatar = async (e: React.ChangeEvent<HTMLInputElement>) => { const file = e.target.files?.[0]; if (!file) return;
    setAvatarPreview(URL.createObjectURL(file)); if (!['image/jpeg','image/png','image/webp'].includes(file.type)) { showToast('格式不支持','error'); return; }
    if (file.size > 2*1024*1024) { showToast('最大 2MB','error'); return; }
    setAvatarUploading(true);
    try { const r = await apiUpload<{avatar_url:string}>('/profile/avatar', file); setEditAvatar(r.avatar_url); showToast('已上传'); }
    catch (err: unknown) { showToast(getApiErrorMessage(err,'失败'),'error'); setAvatarPreview(''); }
    finally { setAvatarUploading(false); if (fileRef.current) fileRef.current.value=''; } };

  const changePwd = async (e: React.FormEvent) => { e.preventDefault(); setPwdMsg('');
    if (!oldPwd) { showPwd('input old password','error'); return; } if (!newPwd||newPwd.length<6) { showPwd('至少 6 位','error'); return; }
    if (newPwd!==cfmPwd) { showPwd('不匹配','error'); return; } setPwdSaving(true);
    try { await apiPut('/auth/change-password', { old_password: oldPwd, new_password: newPwd }); showPwd('成功'); setOldPwd(''); setNewPwd(''); setCfmPwd(''); }
    catch (err: unknown) { showPwd(getApiErrorMessage(err,'失败'),'error'); } finally { setPwdSaving(false); } };

  const delInit = () => { if (!delPwd) { showDel('输入密码','error'); return; } setDelConfirm(true); };
  const delAccount = async () => { setDelSaving(true); setDelMsg('');
    try { await apiDelete('/auth/account', { password: delPwd }); removeToken(); clearUser(); notifyAuthChange(); navigate('/', { replace: true }); }
    catch (err: unknown) { showDel(getApiErrorMessage(err,'失败')); setDelConfirm(false); } finally { setDelSaving(false); } };

  if (loading) return <div className="max-w-3xl mx-auto px-6 py-10"><div className="skeleton h-48 pixel-corners mb-4" /><div className="skeleton h-64 pixel-corners" /></div>;
  if (error) return <div className="max-w-3xl mx-auto px-6 py-10"><div className="p-8 text-center text-danger pixel-corners" style={{ background: 'rgba(224,112,112,0.06)', border: '1px solid rgba(224,112,112,0.2)' }}>{error}</div></div>;
  if (!detail) return <div className="max-w-3xl mx-auto px-6 py-10 text-center text-text-muted"><p className="text-text" style={{ fontSize: 20, fontWeight: 700 }}>用户不存在</p></div>;

  const { user, availabilities } = detail;
  const isSelf = currentUser?.email === user.email;
  const abg = avatarColors[user.nickname.charCodeAt(0) % avatarColors.length];

  return (
    <div className="max-w-3xl mx-auto px-6 py-10">
      {toast && <div className={`px-3 py-2 mb-4 pixel-corners-sm ${toastType==='error'?'text-danger':'text-success'}`} style={{ fontSize: 18, background: toastType==='error'?'rgba(224,112,112,0.08)':'rgba(120,184,128,0.08)', border: `1px solid ${toastType==='error'?'rgba(224,112,112,0.2)':'rgba(120,184,128,0.2)'}` }}>{toast}</div>}

      <div className="bg-card border border-border pixel-corners p-6 mb-4">
        <div className="flex items-start gap-4">
          {isSelf ? (
            <div className="flex-shrink-0">
              {(avatarPreview||editAvatar) ? <img src={avatarPreview||editAvatar} alt="" className="w-14 h-14 object-cover border border-border" /> :
                <div className="w-14 h-14 flex items-center justify-center text-white font-bold" style={{ background: abg, fontSize: 22 }}>{user.nickname.charAt(0)}</div>}
            </div>
          ) : (
            user.avatar ? <img src={user.avatar} alt="" className="w-14 h-14 object-cover" /> :
              <div className="w-14 h-14 flex items-center justify-center text-white font-bold flex-shrink-0" style={{ background: abg, fontSize: 22 }}>{user.nickname.charAt(0)}</div>
          )}
          <div className="flex-1 min-w-0">
            {isSelf ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>昵称</span><input className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={editNickname} onChange={(e) => setEditNickname(e.target.value)} /></label>
                <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>学号</span><input className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={editStudentId} onChange={(e) => setEditStudentId(e.target.value)} /></label>
                <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>院系</span><input className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={editDepartment} onChange={(e) => setEditDepartment(e.target.value)} placeholder="计算机学院" /></label>
                <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>标签</span><input className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={editTagsStr} onChange={(e) => setEditTagsStr(e.target.value)} placeholder="前端, 后端" /></label>
                <label className="flex flex-col gap-1 sm:col-span-2"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>头像</span>
                  <div className="flex items-center gap-2"><input className="flex-1 px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={editAvatar} onChange={(e) => setEditAvatar(e.target.value)} placeholder="头像链接" />
                    <input ref={fileRef} type="file" accept="image/jpeg,image/png,image/webp" onChange={uploadAvatar} className="hidden" />
                    <button type="button" onClick={() => fileRef.current?.click()} disabled={avatarUploading} className="pixel-btn primary whitespace-nowrap" style={{ fontSize: 18 }}>{avatarUploading?'...':'upload'}</button>
                  </div>
                </label>
                <label className="flex flex-col gap-1 sm:col-span-2"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>联系方式</span><input className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={editContactInfo} onChange={(e) => setEditContactInfo(e.target.value)} placeholder="微信 / QQ" /></label>
              </div>
            ) : (
              <>
                <h1 className="text-text" style={{ fontSize: 26, fontWeight: 700 }}>{user.nickname}</h1>
                <p className="text-text-secondary mt-1" style={{ fontSize: 17 }}>{user.department||'--'} · {user.student_id}</p>
                {user.tags && user.tags.length > 0 && <div className="flex flex-wrap gap-1.5 mt-2">{user.tags.map((t) => <span key={t} className="pixel-tag text-brand-600" style={{ borderColor: 'rgba(224,184,104,0.2)' }}>{t}</span>)}</div>}
              </>
            )}
          </div>
        </div>
        {!isSelf && user.contact_info && <div className="mt-4 px-3 py-2 bg-surface-alt border border-border pixel-corners-sm text-text-secondary" style={{ fontSize: 18 }}>contact: {user.contact_info}</div>}
        {isSelf && <button onClick={saveProfile} disabled={saving} className="pixel-btn primary mt-5" style={{ fontSize: 18 }}>{saving?'...':'save'}</button>}
      </div>

      {isSelf && (
        <div className="bg-card border border-border pixel-corners p-6 mb-4">
          <h2 className="text-text mb-4" style={{ fontSize: 20, fontWeight: 700 }}>change password</h2>
          {pwdMsg && <div className={`px-3 py-2 mb-4 pixel-corners-sm ${pwdType==='error'?'text-danger':'text-success'}`} style={{ fontSize: 18, background: pwdType==='error'?'rgba(224,112,112,0.08)':'rgba(120,184,128,0.08)', border: `1px solid ${pwdType==='error'?'rgba(224,112,112,0.2)':'rgba(120,184,128,0.2)'}` }}>{pwdMsg}</div>}
          <form onSubmit={changePwd}>
            <div className="grid grid-cols-1 gap-4 mb-5">
              <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>旧密码</span><input type="password" className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={oldPwd} onChange={(e) => setOldPwd(e.target.value)} disabled={pwdSaving} /></label>
              <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>新密码</span><input type="password" className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={newPwd} onChange={(e) => setNewPwd(e.target.value)} disabled={pwdSaving} /></label>
              <label className="flex flex-col gap-1"><span className="text-text-muted tracking-wider" style={{ fontSize: 17 }}>确认</span><input type="password" className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm" style={{ fontSize: 18 }} value={cfmPwd} onChange={(e) => setCfmPwd(e.target.value)} disabled={pwdSaving} /></label>
            </div>
            <button type="submit" disabled={pwdSaving} className="pixel-btn primary" style={{ fontSize: 18 }}>{pwdSaving?'...':'change'}</button>
          </form>
        </div>
      )}

      {isSelf && (
        <div className="bg-card border border-danger/30 pixel-corners p-6 mb-4">
          <h2 className="text-danger mb-1" style={{ fontSize: 20, fontWeight: 700 }}>delete account</h2>
          <p className="text-text-secondary mb-4" style={{ fontSize: 17 }}>此操作不可撤销</p>
          {delMsg && <div className={`px-3 py-2 mb-4 pixel-corners-sm ${delType==='error'?'text-danger':'text-success'}`} style={{ fontSize: 18, background: delType==='error'?'rgba(224,112,112,0.08)':'rgba(120,184,128,0.08)', border: `1px solid ${delType==='error'?'rgba(224,112,112,0.2)':'rgba(120,184,128,0.2)'}` }}>{delMsg}</div>}
          {!delConfirm ? (
            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
              <input type="password" className="px-3 py-2 bg-surface border border-border text-text pixel-corners-sm w-full sm:w-48" style={{ fontSize: 18 }} value={delPwd} onChange={(e) => setDelPwd(e.target.value)} placeholder="密码" disabled={delSaving} />
              <button onClick={delInit} disabled={delSaving} className="pixel-btn" style={{ fontSize: 18, color: 'var(--color-danger)', borderColor: 'rgba(224,112,112,0.3)' }}>{delSaving?'...':'delete'}</button>
            </div>
          ) : (
            <div className="p-4 pixel-corners-sm" style={{ background: 'rgba(224,112,112,0.06)', border: '1px solid rgba(224,112,112,0.2)' }}>
              <p className="text-danger mb-1" style={{ fontSize: 17, fontWeight: 700 }}>确认注销?</p>
              <p className="text-danger mb-4" style={{ fontSize: 17 }}>此操作不可撤销</p>
              <div className="flex gap-3">
                <button onClick={delAccount} disabled={delSaving} className="pixel-btn primary" style={{ fontSize: 18, background: 'var(--color-danger)', borderColor: 'var(--color-danger)' }}>{delSaving?'...':'confirm'}</button>
                <button onClick={() => { setDelConfirm(false); setDelPwd(''); }} disabled={delSaving} className="pixel-btn" style={{ fontSize: 18 }}>取消</button>
              </div>
            </div>
          )}
        </div>
      )}

      <WeekCalendar availabilities={availabilities} isSelf={isSelf} onBook={handleBook} />
    </div>
  );
}

export default UserDetail;
