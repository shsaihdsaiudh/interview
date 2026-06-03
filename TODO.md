# TODO

待实现的功能想法，按优先级排列。

---

## 💡 想法池

### 智能匹配 Agent

- **核心思路**：用户有个人介绍（标签、院系、技能、bio），发帖/发布链接的人也带介绍，系统自动分析并做双向匹配，形成匹配池——展示「你可能感兴趣的人」和「谁对你感兴趣」。
- **可能的方向**：
  1. 标签智能推荐 — 利用现有 tags + 院系做加权匹配，在「找人」页面增加推荐排序
  2. 帖子广场 + 匹配 — 先实现帖子功能（面试需求/offer），发帖带标签，系统推荐匹配帖子
  3. 双向匹配池 — 新增 bio 字段，计算匹配分，双方互感兴趣后匹配成功
  4. 分阶段迭代 — 标签推荐 → 帖子广场 → 双向匹配池
- **涉及模块**：后端 matching 服务 + endpoint、前端推荐页 / 增强 FindPeople / 实现 Posts、可能需要用户 bio 字段、帖子表、匹配算法

---

## 🚧 进行中

（暂无）

## ✅ 已完成

### 招募卡片 (Recruitment Card) — 2026-06-03

- **核心思路**：用户可以发布自己的面试需求或能力名片（招募卡片），包含技能标签、目标公司、角色、经验年限、个人简介等，供他人搜索和发现。
- **已完成工作**：
  1. **domain/recruitment/** — RecruitmentCard 聚合根（Activate/Deactivate/CanBeManagedBy 行为方法）、RecruitmentCardRepository 接口、领域错误、单元测试
  2. **infrastructure/persistence/** — recruitment_cards 表 migration（docker/init.sql + testutil/pg.go）、PostgresRepo 实现（Upsert/FindByUserID/List 含动态筛选和分页）
  3. **application/** — RecruitmentService（CreateOrUpdateCard / GetCardByUserID / ListCards + 角色校验 + 用户存在性校验）、单元测试
  4. **interfaces/handler/** — RecruitmentHandler（PUT/GET /api/v1/recruitment-card、GET /api/v1/recruitment-cards 含分页和筛选）
  5. **routes.go + main.go** — 注册路由和依赖注入
- **涉及模块**：domain/recruitment、infrastructure/persistence、application/recruitment_service、interfaces/handler/recruitment_handler、routes.go、main.go、docker/init.sql、testutil/pg.go
