import { useNavigate } from 'react-router-dom';
import { getUser } from './Navbar';

export interface RecruitmentCardData {
  id: string;
  user_id: string;
  nickname: string;
  avatar: string;
  role: '面试者' | '面试官' | '两者皆可';
  skills: string[];
  target_companies: string[];
  experience_years: number;
  bio: string;
}

const avatarColors = ['#e0b868','#78b880','#d4a040','#e07070','#aaaab2','#72727c'];

const roleConfig: Record<string, { label: string; cls: string }> = {
  interviewee: { label: '面试者', cls: 'text-brand-600' },
  interviewer: { label: '面试官', cls: 'text-success' },
  both: { label: '两者皆可', cls: 'text-warning' },
};

function RecruitmentCard({ card }: { card: RecruitmentCardData }) {
  const navigate = useNavigate();
  const currentUser = getUser();

  const displayName = card.nickname || (card.user_id ? card.user_id.substring(0, 8) : '?');
  const firstChar = displayName.charAt(0);
  const avatarBg = avatarColors[firstChar.charCodeAt(0) % avatarColors.length];
  const role = roleConfig[card.role] || roleConfig.interviewee;

  const visibleSkills = card.skills?.slice(0, 3) || [];
  const hiddenSkillCount = (card.skills?.length || 0) - 3;
  const visibleCompanies = card.target_companies?.slice(0, 2) || [];
  const hiddenCompanyCount = (card.target_companies?.length || 0) - 2;

  const truncateBio = (bio: string, maxLen = 100): string =>
    !bio ? '' : bio.length > maxLen ? bio.slice(0, maxLen) + '...' : bio;

  const handleBook = () => {
    const target = `/user/${card.user_id}`;
    if (!currentUser) navigate(`/login?redirect=${encodeURIComponent(target)}`);
    else navigate(target);
  };

  return (
    <div className="bg-card border border-border pixel-corners p-5 card-hover flex flex-col">
      <div className="flex items-center gap-3">
        {card.avatar ? (
          <img src={card.avatar} alt={displayName} className="w-10 h-10 object-cover flex-shrink-0" />
        ) : (
          <div className="w-10 h-10 flex items-center justify-center text-white font-bold flex-shrink-0"
               style={{ background: avatarBg, fontSize: 18 }}>
            {firstChar}
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="text-text truncate" style={{ fontSize: 17, fontWeight: 700 }}>{displayName}</div>
          <span className={role.cls} style={{ fontSize: 17 }}>[{role.label}]</span>
        </div>
      </div>

      {card.experience_years != null && (
        <div className="mt-3 text-text-muted" style={{ fontSize: 18 }}>
          exp: {card.experience_years}y
        </div>
      )}

      {visibleSkills.length > 0 && (
        <div className="mt-3">
          <div className="text-text-muted mb-1.5 tracking-wider" style={{ fontSize: 17 }}>技能</div>
          <div className="flex flex-wrap gap-1.5">
            {visibleSkills.map((s) => (
              <span key={s} className="pixel-tag text-brand-600" style={{ borderColor: 'rgba(224,184,104,0.2)' }}>{s}</span>
            ))}
            {hiddenSkillCount > 0 && <span className="pixel-tag">+{hiddenSkillCount}</span>}
          </div>
        </div>
      )}

      {visibleCompanies.length > 0 && (
        <div className="mt-3">
          <div className="text-text-muted mb-1.5 tracking-wider" style={{ fontSize: 17 }}>目标公司</div>
          <div className="flex flex-wrap gap-1.5">
            {visibleCompanies.map((c) => (
              <span key={c} className="pixel-tag text-success" style={{ borderColor: 'rgba(120,184,128,0.2)' }}>{c}</span>
            ))}
            {hiddenCompanyCount > 0 && <span className="pixel-tag">+{hiddenCompanyCount}</span>}
          </div>
        </div>
      )}

      {card.bio && (
        <p className="mt-3 text-text-secondary flex-1" style={{ fontSize: 17, lineHeight: 1.6 }}>
          {truncateBio(card.bio)}
        </p>
      )}

      <button onClick={handleBook}
        className="pixel-btn primary w-full justify-center mt-4"
        style={{ fontSize: 18 }}>
        预约
      </button>
    </div>
  );
}

export default RecruitmentCard;
