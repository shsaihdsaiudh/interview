import { Link } from 'react-router-dom';

export interface RecruitmentCardData {
  id: string;
  user_id: string;
  nickname: string;
  avatar: string;
  role: 'interviewee' | 'interviewer' | 'both';
  skills: string[];
  target_companies: string[];
  experience_years: number;
  bio: string;
}

const avatarColors = ['#6366f1', '#10b981', '#f59e0b', '#ec4899', '#8b5cf6', '#06b6d4', '#ef4444', '#f97316'];

const roleConfig: Record<string, { label: string; bg: string; text: string }> = {
  interviewee: { label: '找Mock伙伴', bg: 'bg-brand-50', text: 'text-brand-700' },
  interviewer: { label: '可做面试官', bg: 'bg-emerald-50', text: 'text-emerald-700' },
  both: { label: '互相模拟', bg: 'bg-amber-50', text: 'text-amber-700' },
};

function RecruitmentCard({ card }: { card: RecruitmentCardData }) {
  const avatarBg = avatarColors[card.nickname.charCodeAt(0) % avatarColors.length];
  const role = roleConfig[card.role] || roleConfig.interviewee;

  const visibleSkills = card.skills?.slice(0, 3) || [];
  const hiddenSkillCount = (card.skills?.length || 0) - 3;

  const visibleCompanies = card.target_companies?.slice(0, 2) || [];
  const hiddenCompanyCount = (card.target_companies?.length || 0) - 2;

  const truncateBio = (bio: string, maxLen = 100): string => {
    if (!bio) return '';
    return bio.length > maxLen ? bio.slice(0, maxLen) + '...' : bio;
  };

  return (
    <div
      className="group bg-card rounded-2xl border border-border p-5
                 hover:-translate-y-1 hover:shadow-md hover:border-brand-200
                 transition-all duration-200 flex flex-col"
    >
      {/* ── Header: Avatar + Name + Role ── */}
      <div className="flex items-center gap-3">
        {card.avatar ? (
          <img
            src={card.avatar}
            alt={card.nickname}
            className="w-11 h-11 rounded-full object-cover flex-shrink-0"
          />
        ) : (
          <div
            className="w-11 h-11 rounded-full flex items-center justify-center text-white text-base font-bold flex-shrink-0"
            style={{ background: avatarBg }}
          >
            {card.nickname.charAt(0)}
          </div>
        )}
        <div className="min-w-0 flex-1">
          <div className="font-semibold text-text text-sm truncate">
            {card.nickname}
          </div>
          <span
            className={`inline-block mt-0.5 px-2 py-0.5 rounded-md text-xs font-medium ${role.bg} ${role.text}`}
          >
            {role.label}
          </span>
        </div>
      </div>

      {/* ── Experience ── */}
      {card.experience_years != null && (
        <div className="mt-3 text-xs text-text-muted">
          经验年限：{card.experience_years} 年
        </div>
      )}

      {/* ── Skills ── */}
      {visibleSkills.length > 0 && (
        <div className="mt-3">
          <div className="text-xs text-text-muted mb-1.5">技能</div>
          <div className="flex flex-wrap gap-1.5">
            {visibleSkills.map((s) => (
              <span
                key={s}
                className="px-2 py-0.5 rounded-md bg-brand-50 text-brand-700 text-xs font-medium"
              >
                {s}
              </span>
            ))}
            {hiddenSkillCount > 0 && (
              <span className="px-2 py-0.5 rounded-md bg-gray-100 text-text-muted text-xs font-medium">
                +{hiddenSkillCount}
              </span>
            )}
          </div>
        </div>
      )}

      {/* ── Target Companies ── */}
      {visibleCompanies.length > 0 && (
        <div className="mt-3">
          <div className="text-xs text-text-muted mb-1.5">目标公司</div>
          <div className="flex flex-wrap gap-1.5">
            {visibleCompanies.map((c) => (
              <span
                key={c}
                className="px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-700 text-xs font-medium"
              >
                {c}
              </span>
            ))}
            {hiddenCompanyCount > 0 && (
              <span className="px-2 py-0.5 rounded-md bg-gray-100 text-text-muted text-xs font-medium">
                +{hiddenCompanyCount}
              </span>
            )}
          </div>
        </div>
      )}

      {/* ── Bio ── */}
      {card.bio && (
        <p className="mt-3 text-sm text-text-secondary leading-relaxed flex-1">
          {truncateBio(card.bio)}
        </p>
      )}

      {/* ── CTA Button ── */}
      <Link
        to={`/user/${card.user_id}`}
        className="mt-4 inline-flex items-center justify-center w-full px-4 py-2
                   rounded-lg bg-brand-600 text-white text-sm font-medium
                   hover:bg-brand-700 transition no-underline cursor-pointer"
      >
        预约ta
      </Link>
    </div>
  );
}

export default RecruitmentCard;
