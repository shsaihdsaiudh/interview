// Seed 脚本：向空数据库注入测试数据，方便手动测试。
// 运行方式：go run ./cmd/seed/
// 幂等：重复运行不会产生重复数据。
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"interview-server/domain/appointment"
	"interview-server/domain/recruitment"
	"interview-server/domain/user"
	"interview-server/infrastructure/persistence"
)

// ── 测试用户定义 ──

type testUser struct {
	Nickname        string
	Email           string
	Password        string
	Skills          []string
	TargetCompanies []string
	Role            string
	ExperienceYears int
}


var testUsers = []testUser{
	{
		Nickname:        "张三",
		Email:           "zhangsan@test.com",
		Password:        "test123",
		Skills:          []string{"React", "TypeScript", "Node.js"},
		TargetCompanies: []string{"字节", "腾讯"},
		Role:            recruitment.RoleBoth,
		ExperienceYears: 3,
	},
	{
		Nickname:        "李四",
		Email:           "lisi@test.com",
		Password:        "test123",
		Skills:          []string{"Go", "Python", "Docker"},
		TargetCompanies: []string{"Google", "Amazon"},
		Role:            recruitment.RoleInterviewer,
		ExperienceYears: 5,
	},
	{
		Nickname:        "王五",
		Email:           "wangwu@test.com",
		Password:        "test123",
		Skills:          []string{"Java", "Spring", "MySQL"},
		TargetCompanies: []string{"阿里", "美团"},
		Role:            recruitment.RoleInterviewee,
		ExperienceYears: 2,
	},
	{
		Nickname:        "赵六",
		Email:           "zhaoliu@test.com",
		Password:        "test123",
		Skills:          []string{"Vue", "JavaScript", "CSS"},
		TargetCompanies: []string{"字节", "腾讯", "阿里"},
		Role:            recruitment.RoleBoth,
		ExperienceYears: 4,
	},
	{
		Nickname:        "钱七",
		Email:           "qianqi@test.com",
		Password:        "test123",
		Skills:          []string{"Rust", "C++", "系统设计"},
		TargetCompanies: []string{"Microsoft", "Apple"},
		Role:            recruitment.RoleInterviewer,
		ExperienceYears: 7,
	},
}

// ── 入口 ──

func main() {
	ctx := context.Background()

	// 优先使用环境变量 DATABASE_URL，否则使用默认值
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://interview:interview123@localhost:5432/interview_platform?sslmode=disable"
	}

	fmt.Println("🌱 Seed: 连接数据库...")
	pool := persistence.NewPool(ctx, dsn)
	defer pool.Close()

	repo := persistence.NewPostgresRepo(pool)

	// ── 幂等检查：如果测试用户已存在则跳过 ──
	existing, _ := repo.FindByEmail(testUsers[0].Email)
	if existing != nil {
		fmt.Println("✅ 检测到已有测试数据（张三已存在），跳过 seed。")
		fmt.Println()
		printAccountSummary()
		return
	}

	fmt.Println("📝 Seed: 开始注入测试数据...")
	fmt.Println()

	// ── 1. 创建用户 ──
	fmt.Println("👤 创建测试用户...")
	createdUsers := make([]*user.User, len(testUsers))
	for i, tu := range testUsers {
		u := createUser(repo, tu)
		createdUsers[i] = u
		fmt.Printf("  ✓ %s <%s>\n", tu.Nickname, tu.Email)
	}
	fmt.Println()

	// ── 2. 创建招募卡片 ──
	fmt.Println("📇 创建招募卡片...")
	for _, tu := range testUsers {
		createRecruitmentCard(repo, tu)
		fmt.Printf("  ✓ %s — %s | 技能: %v | 目标: %v\n",
			tu.Nickname, tu.Role, tu.Skills, tu.TargetCompanies)
	}
	fmt.Println()

	// ── 3. 生成空闲时间 ──
	fmt.Println("⏰ 生成空闲时间...")
	allAvailabilities := generateAllAvailabilities(repo, createdUsers)
	fmt.Printf("  总计生成 %d 个空闲时间段\n", len(allAvailabilities))
	fmt.Println()

	// ── 4. 创建预约记录 ──
	fmt.Println("📅 创建预约记录...")
	createMockAppointments(repo, allAvailabilities)
	fmt.Println()

	// ── 打印账号汇总 ──
	fmt.Println("══════════════════════════════════════════════")
	fmt.Println("  ✅ Seed 数据注入完成！")
	fmt.Println("══════════════════════════════════════════════")
	fmt.Println()
	printAccountSummary()
}


// ── 用户创建 ──

func createUser(repo *persistence.PostgresRepo, tu testUser) *user.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(tu.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码加密失败 (%s): %v", tu.Email, err)
	}

	u := &user.User{
		Email:         tu.Email,
		PasswordHash:  string(hash),
		Nickname:      tu.Nickname,
		EmailVerified: true,
		CreatedAt:     time.Now(),
	}
	if err := repo.Create(u); err != nil {
		log.Fatalf("创建用户失败 (%s): %v", tu.Email, err)
	}
	return u
}

// ── 招募卡片创建 ──

func createRecruitmentCard(repo *persistence.PostgresRepo, tu testUser) *recruitment.RecruitmentCard {
	now := time.Now()
	card := &recruitment.RecruitmentCard{
		ID:              randomHex(16),
		UserID:          tu.Email,
		Skills:          tu.Skills,
		TargetCompanies: tu.TargetCompanies,
		Role:            tu.Role,
		ExperienceYears: tu.ExperienceYears,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := repo.Upsert(card); err != nil {
		log.Fatalf("创建招募卡片失败 (%s): %v", tu.Email, err)
	}
	return card
}

// ── 空闲时间生成 ──

func generateAllAvailabilities(repo *persistence.PostgresRepo, users []*user.User) []*appointment.Availability {
	var all []*appointment.Availability

	today := time.Now().Truncate(24 * time.Hour)

	for _, u := range users {
		for dayOffset := 0; dayOffset < 7; dayOffset++ {
			dayDate := today.AddDate(0, 0, dayOffset)
			dateStr := dayDate.Format("2006-01-02")
			slots := generateDaySlots(repo, u.Email, dateStr)
			all = append(all, slots...)
		}
	}

	return all
}


// generateDaySlots 为某个用户在指定日期生成 3-6 个不重叠的空闲时间段。
// 时间段在 9:00-21:00 之间，随机 30 分钟或 1 小时。
func generateDaySlots(repo *persistence.PostgresRepo, userEmail, dateStr string) []*appointment.Availability {
	// 所有可能的 30min 起始时间：09:00, 09:30, ..., 20:30（共 24 个）
	type slotInfo struct {
		start    string
		startMin int // 距离 00:00 的分钟数
	}
	var allSlots []slotInfo
	for h := 9; h < 21; h++ {
		allSlots = append(allSlots, slotInfo{
			start:    fmt.Sprintf("%02d:00", h),
			startMin: h * 60,
		})
		allSlots = append(allSlots, slotInfo{
			start:    fmt.Sprintf("%02d:30", h),
			startMin: h*60 + 30,
		})
	}

	// 随机选择 3-6 个不同的索引
	count := randomInt(3, 6)
	indices := randomPickN(len(allSlots), count)

	var occupied [48]bool // 以 30min 为单位，0 = 00:00-00:30

	var slots []*appointment.Availability
	for _, idx := range indices {
		if occupied[idx] {
			continue // 已被更长的前序 slot 覆盖
		}

		si := allSlots[idx]

		// 随机决定时长：30min 或 1hr（默认 1hr）
		durationMin := 60
		if randomBool() {
			durationMin = 30
		}

		// 检查 1hr slot 是否超出 21:00 或与已有 slot 重叠
		if durationMin == 60 {
			if si.startMin+60 > 21*60 {
				durationMin = 30
			} else if idx+1 < len(allSlots) && occupied[idx+1] {
				durationMin = 30
			}
		}

		// 计算结束时间
		endMin := si.startMin + durationMin
		endStr := fmt.Sprintf("%02d:%02d", endMin/60, endMin%60)

		// 标记已占用
		for m := si.startMin / 30; m < endMin/30; m++ {
			occupied[m] = true
		}

		avail := &appointment.Availability{
			ID:        randomHex(16),
			UserID:    userEmail,
			Date:      dateStr,
			StartTime: si.start,
			EndTime:   endStr,
		}

		if err := repo.CreateAvailability(avail); err != nil {
			log.Printf("   ⚠️ 创建空闲时间失败 (%s %s %s-%s): %v",
				userEmail, dateStr, si.start, endStr, err)
			continue
		}
		slots = append(slots, avail)
	}

	return slots
}


// ── 预约记录 ──

func createMockAppointments(repo *persistence.PostgresRepo, allAvailabilities []*appointment.Availability) {
	// 获取各用户的空闲时间段
	userSlots := make(map[string][]*appointment.Availability)
	for _, a := range allAvailabilities {
		userSlots[a.UserID] = append(userSlots[a.UserID], a)
	}

	// 预约 1: 王五 (student) → 李四 (mentor) 第一个空闲时间 (pending)
	userWangWu := testUsers[2].Email
	userLiSi := testUsers[1].Email
	if slots := userSlots[userLiSi]; len(slots) > 0 {
		createAppointment(repo, userLiSi, userWangWu, slots[0],
			"李四你好，想和你模拟一次 Go 后端面试，方便吗？", appointment.StatusPending)
		fmt.Printf("  ✓ 王五 → 李四 [pending]\n")
	}

	// 预约 2: 张三 (student) → 钱七 (mentor) 第一个空闲时间 (accepted)
	userZhangSan := testUsers[0].Email
	userQianQi := testUsers[4].Email
	if slots := userSlots[userQianQi]; len(slots) > 0 {
		createAppointment(repo, userQianQi, userZhangSan, slots[0],
			"钱七你好，想请教一些 Rust 方向的问题，希望能预约你的时间。", appointment.StatusAccepted)
		fmt.Printf("  ✓ 张三 → 钱七 [accepted]\n")
	}

	// 预约 3: 赵六 (student) → 李四 (mentor) 第二个空闲时间 (pending)
	userZhaoLiu := testUsers[3].Email
	if slots := userSlots[userLiSi]; len(slots) > 1 {
		createAppointment(repo, userLiSi, userZhaoLiu, slots[1],
			"李四大佬，求带飞！想模拟 Docker 相关的面试。", appointment.StatusPending)
		fmt.Printf("  ✓ 赵六 → 李四 [pending]\n")
	} else if slots := userSlots[userQianQi]; len(slots) > 1 {
		createAppointment(repo, userQianQi, userZhaoLiu, slots[1],
			"钱七你好，对系统设计面试很感兴趣，希望能约个时间聊聊。", appointment.StatusPending)
		fmt.Printf("  ✓ 赵六 → 钱七 [pending]\n")
	}
}

func createAppointment(repo *persistence.PostgresRepo, mentorID, studentID string,
	slot *appointment.Availability, message, status string) {
	appt := &appointment.Appointment{
		ID:         randomHex(16),
		MentorID:   mentorID,
		StudentID:  studentID,
		TimeSlotID: slot.ID,
		Message:    message,
		Status:     status,
		CreatedAt:  time.Now(),
	}
	if err := repo.CreateAppointment(appt); err != nil {
		log.Printf("   ⚠️ 创建预约失败: %v", err)
	}
}


// ── 随机工具函数 ──

// randomHex 生成指定字节数的随机 hex 字符串。
func randomHex(bytes int) string {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("生成随机 ID 失败: %v", err))
	}
	return hex.EncodeToString(b)
}

// randomInt 返回 [min, max] 范围内的随机整数。
func randomInt(min, max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return min + int(n.Int64())
}

// randomBool 返回随机布尔值（50% 概率）。
func randomBool() bool {
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return false
	}
	return n.Int64() == 1
}

// randomPickN 从 0..total-1 中随机选取 n 个不重复的整数。
func randomPickN(total, n int) []int {
	if n > total {
		n = total
	}
	pool := make([]int, total)
	for i := range pool {
		pool[i] = i
	}
	for i := 0; i < n; i++ {
		j := randomInt(i, total-1)
		pool[i], pool[j] = pool[j], pool[i]
	}
	return pool[:n]
}

// ── 打印账号汇总 ──

func printAccountSummary() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  📋 测试账号列表（密码统一为: test123）")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  %-6s  %-25s  %-15s  %-12s  %s\n", "昵称", "邮箱", "角色", "年限", "技能")
	fmt.Println("  ─────────────────────────────────────────────────────────────────")
	for _, tu := range testUsers {
		fmt.Printf("  %-6s  %-25s  %-15s  %-12d  %v\n",
			tu.Nickname, tu.Email, tu.Role, tu.ExperienceYears, tu.Skills)
	}
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  使用以上任一邮箱 + 密码 test123 即可登录")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

