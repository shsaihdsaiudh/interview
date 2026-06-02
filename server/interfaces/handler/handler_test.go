package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"interview-server/application"
	"interview-server/domain/appointment"
	"interview-server/infrastructure/persistence"
	"interview-server/interfaces/middleware"
	"interview-server/testutil"
)

// =============================================================================
// 集成测试夹具
// =============================================================================

var (
	testRouter  *gin.Engine
	testRepo    *persistence.PostgresRepo
	testUserSvc *application.UserService
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	ctx := context.Background()

	pg, err := testutil.StartPostgres(ctx)
	if err != nil {
		panic("启动测试 PostgreSQL 失败: " + err.Error())
	}

	pool := pg.NewPool(ctx)
	repo := persistence.NewPostgresRepo(pool)
	testRepo = repo

	userSvc := application.NewUserService(repo, repo, nil)
	apptSvc := application.NewAppointmentService(repo, repo)
	testUserSvc = userSvc

	userH := NewUserHandler(userSvc)
	apptH := NewAppointmentHandler(apptSvc)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// ── 内联路由（避免 import cycle）──
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", Ping)

		auth := v1.Group("/auth")
		{
			auth.POST("/send-code", userH.SendCode)
			auth.POST("/register", userH.Register)
			auth.POST("/login", userH.Login)

			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuth())
			{
				authRequired.GET("/me", userH.Me)
				authRequired.PUT("/change-password", userH.ChangePassword)
				authRequired.DELETE("/account", userH.DeleteAccount)
			}
		}

		v1.GET("/users", userH.ListUsers)
		v1.GET("/users/:id", middleware.OptionalJWTAuth(), userH.GetUser)

		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		{
			protected.GET("/profile", userH.GetProfile)
			protected.PUT("/profile", userH.UpdateProfile)
			protected.POST("/profile/avatar", userH.UploadAvatar)

			protected.GET("/availability", apptH.GetMyAvailability)
			protected.POST("/availability", apptH.AddAvailability)
			protected.DELETE("/availability/:id", apptH.DeleteAvailability)

			protected.POST("/appointments", apptH.CreateAppointment)
			protected.GET("/appointments", apptH.GetMyAppointments)
			protected.PUT("/appointments/:id/accept", apptH.AcceptAppointment)
			protected.PUT("/appointments/:id/reject", apptH.RejectAppointment)
		}
	}

	return r
}

func getRouter() *gin.Engine {
	if testRouter == nil {
		testRouter = setupRouter()
	}
	return testRouter
}

// =============================================================================
// 辅助函数
// =============================================================================

func doJSON(method, path, body string) *httptest.ResponseRecorder {
	return doJSONWithToken(method, path, body, "")
}

func doJSONWithToken(method, path, body, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	getRouter().ServeHTTP(w, req)
	return w
}

func parseJSON(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	return result
}

// registerHelper 新注册流程：发验证码 → 注册（直接返回 JWT）。
func registerHelper(t *testing.T, email, password, nickname, studentID string) string {
	t.Helper()

	// 1. 发送验证码
	w := doJSON("POST", "/api/v1/auth/send-code", `{"email":"`+email+`"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("send-code: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. 从 service 中获取验证码（集成测试中通过全局 testUserSvc 访问）
	code := testUserSvc.VerificationCodeForTest(email)
	if code == "" {
		t.Fatal("verification code not found")
	}

	// 3. 注册
	regBody := `{"email":"` + email + `","code":"` + code + `","password":"` + password + `","nickname":"` + nickname + `","student_id":"` + studentID + `"}`
	w = doJSON("POST", "/api/v1/auth/register", regBody)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatal("no token in register response")
	}
	return token
}

func TestHandler_UploadAvatar_Success(t *testing.T) {
	token := registerHelper(t, "avatar-ok@std.uestc.edu.cn", "pass123", "AvatarOK", "AV001")

	w := doMultipartWithToken("POST", "/api/v1/profile/avatar", "avatar", "photo.jpg", minimalJPEG(), token)
	if w.Code != http.StatusOK {
		t.Fatalf("upload avatar: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	avatarURL, ok := resp["avatar_url"].(string)
	if !ok || avatarURL == "" {
		t.Fatalf("avatar_url missing in response: %s", w.Body.String())
	}
	if !strings.HasPrefix(avatarURL, "/uploads/avatars/") {
		t.Errorf("avatar_url should start with /uploads/avatars/, got %q", avatarURL)
	}

	w = doJSONWithToken("GET", "/api/v1/profile", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("get profile after upload: %d: %s", w.Code, w.Body.String())
	}
	profileResp := parseJSON(w)
	profileUser, ok := profileResp["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("profile user missing: %s", w.Body.String())
	}
	if profileUser["avatar"] != avatarURL {
		t.Errorf("profile avatar = %v, want %v", profileUser["avatar"], avatarURL)
	}
}

func TestHandler_UploadAvatar_InvalidType(t *testing.T) {
	token := registerHelper(t, "avatar-bt@std.uestc.edu.cn", "pass123", "AvatarBT", "AV002")

	textData := []byte("this is not an image")
	w := doMultipartWithToken("POST", "/api/v1/profile/avatar", "avatar", "test.txt", textData, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-image file, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UploadAvatar_TooLarge(t *testing.T) {
	token := registerHelper(t, "avatar-lg@std.uestc.edu.cn", "pass123", "AvatarLG", "AV003")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("avatar", "big.jpg")
	largeData := make([]byte, 2<<20+1)
	part.Write(largeData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/profile/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	getRouter().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized file, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UploadAvatar_Unauthorized(t *testing.T) {
	w := doMultipartWithToken("POST", "/api/v1/profile/avatar", "avatar", "photo.jpg", minimalJPEG(), "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d: %s", w.Code, w.Body.String())
	}
}

// =============================================================================
// 基础接口
// =============================================================================

func TestHandler_HealthCheck(t *testing.T) {
	w := doJSON("GET", "/api/health", "")
	if w.Code != http.StatusOK {
		t.Errorf("health: expected 200, got %d", w.Code)
	}
}

func TestHandler_Ping(t *testing.T) {
	w := doJSON("GET", "/api/v1/ping", "")
	if w.Code != http.StatusOK {
		t.Errorf("ping: expected 200, got %d", w.Code)
	}
}

// =============================================================================
// 认证流程
// =============================================================================

func TestHandler_SendCode_InvalidEmail(t *testing.T) {
	w := doJSON("POST", "/api/v1/auth/send-code", `{"email":"alice@gmail.com"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-uestc email, got %d", w.Code)
	}
}

func TestHandler_Register_InvalidCode(t *testing.T) {
	email := "badcode@std.uestc.edu.cn"
	doJSON("POST", "/api/v1/auth/send-code", `{"email":"`+email+`"}`)

	w := doJSON("POST", "/api/v1/auth/register", `{"email":"`+email+`","code":"000000","password":"123456"}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong code, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Register_Duplicate(t *testing.T) {
	email := "dup3@std.uestc.edu.cn"
	_ = registerHelper(t, email, "123456", "Dup", "D001")

	// 尝试重复注册
	w := doJSON("POST", "/api/v1/auth/send-code", `{"email":"`+email+`"}`)
	if w.Code != http.StatusConflict {
		t.Errorf("send-code duplicate: expected 409, got %d", w.Code)
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	email := "wrongpw3@std.uestc.edu.cn"
	_ = registerHelper(t, email, "correct", "WP", "W001")

	w := doJSON("POST", "/api/v1/auth/login", `{"email":"`+email+`","password":"wrong"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", w.Code)
	}
}

func TestHandler_RegisterLoginMe_FullFlow(t *testing.T) {
	token := registerHelper(t, "fullflow3@std.uestc.edu.cn", "pass123", "FullFlow", "F001")

	w := doJSONWithToken("GET", "/api/v1/auth/me", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("me: %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	if resp["email"] != "fullflow3@std.uestc.edu.cn" {
		t.Errorf("email = %v", resp["email"])
	}
	if resp["account_status"] != "active" {
		t.Errorf("account_status = %v, want active", resp["account_status"])
	}
}

// =============================================================================
// 个人信息
// =============================================================================

func TestHandler_GetProfile(t *testing.T) {
	token := registerHelper(t, "profile3@std.uestc.edu.cn", "pass123", "Profile", "P001")

	w := doJSONWithToken("PUT", "/api/v1/profile", `{"contact_info":"wechat-profile3"}`, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update profile contact: %d: %s", w.Code, w.Body.String())
	}

	w = doJSONWithToken("GET", "/api/v1/profile", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("get profile: %d", w.Code)
	}
	resp := parseJSON(w)
	profileUser, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("profile user missing: %s", w.Body.String())
	}
	if profileUser["contact_info"] != "wechat-profile3" {
		t.Errorf("contact_info = %v, want own contact", profileUser["contact_info"])
	}
}

func TestHandler_UpdateProfile(t *testing.T) {
	token := registerHelper(t, "updprof3@std.uestc.edu.cn", "pass123", "OldName", "U001")

	body := `{"nickname":"NewName","department":"Math"}`
	w := doJSONWithToken("PUT", "/api/v1/profile", body, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update profile: %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ListUsers(t *testing.T) {
	registerHelper(t, "list1b@std.uestc.edu.cn", "pass123", "L1", "L001")
	registerHelper(t, "list2b@std.uestc.edu.cn", "pass123", "L2", "L002")

	// 默认分页（page=1, page_size=20）
	w := doJSON("GET", "/api/v1/users", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	if users, ok := resp["users"].([]interface{}); ok {
		// 至少有我们刚注册的 2 个用户（可能还有之前测试留下的）
		if len(users) < 2 {
			t.Errorf("expected at least 2 users in list, got %d", len(users))
		}
	} else {
		t.Error("users field missing or wrong type")
	}

	if total, ok := resp["total"].(float64); ok {
		if int(total) < 2 {
			t.Errorf("expected total >= 2, got %v", total)
		}
	} else {
		t.Error("total field missing or wrong type")
	}

	if page, ok := resp["page"].(float64); ok {
		if int(page) != 1 {
			t.Errorf("expected page 1, got %v", page)
		}
	} else {
		t.Error("page field missing or wrong type")
	}

	if pageSize, ok := resp["page_size"].(float64); ok {
		if int(pageSize) != 20 {
			t.Errorf("expected page_size 20, got %v", pageSize)
		}
	} else {
		t.Error("page_size field missing or wrong type")
	}
}

func TestHandler_ListUsers_Pagination(t *testing.T) {
	registerHelper(t, "page1@std.uestc.edu.cn", "pass123", "P1", "P001")
	registerHelper(t, "page2@std.uestc.edu.cn", "pass123", "P2", "P002")
	registerHelper(t, "page3@std.uestc.edu.cn", "pass123", "P3", "P003")

	// 指定分页参数
	w := doJSON("GET", "/api/v1/users?page=1&page_size=2", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list users with pagination: %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	if pageSize, ok := resp["page_size"].(float64); ok {
		if int(pageSize) != 2 {
			t.Errorf("expected page_size 2, got %v", pageSize)
		}
	}

	if users, ok := resp["users"].([]interface{}); ok {
		if len(users) > 2 {
			t.Errorf("expected at most 2 users, got %d", len(users))
		}
	}

	// page_size 超过最大限制 100 时应被截断
	w = doJSON("GET", "/api/v1/users?page=1&page_size=200", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list users with large page_size: %d: %s", w.Code, w.Body.String())
	}
	resp = parseJSON(w)
	if pageSize, ok := resp["page_size"].(float64); ok {
		if int(pageSize) != 100 {
			t.Errorf("expected page_size capped at 100, got %v", pageSize)
		}
	}
}

// =============================================================================
// 空闲时间
// =============================================================================

func TestHandler_Availability_CRUD(t *testing.T) {
	token := registerHelper(t, "avail3@std.uestc.edu.cn", "pass123", "Avail", "A001")

	body := `{"date":"2099-12-31","start_time":"10:00","end_time":"11:00"}`
	w := doJSONWithToken("POST", "/api/v1/availability", body, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("add availability: %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	slotID, _ := resp["id"].(string)

	w = doJSONWithToken("GET", "/api/v1/availability", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("get availability: %d", w.Code)
	}

	w = doJSONWithToken("DELETE", "/api/v1/availability/"+slotID, "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete availability: %d: %s", w.Code, w.Body.String())
	}
}

// =============================================================================
// 预约全流程
// =============================================================================

func TestHandler_Appointment_FullFlow(t *testing.T) {
	mentorToken := registerHelper(t, "mentor4@std.uestc.edu.cn", "pass123", "MentorFlow", "MF001")
	studentToken := registerHelper(t, "student4@std.uestc.edu.cn", "pass123", "StudentFlow", "SF001")

	w := doJSONWithToken("PUT", "/api/v1/profile", `{"contact_info":"wechat-mentor4"}`, mentorToken)
	if w.Code != http.StatusOK {
		t.Fatalf("mentor update contact: %d: %s", w.Code, w.Body.String())
	}

	w = doJSONWithToken("POST", "/api/v1/availability", `{"date":"2099-12-31","start_time":"14:00","end_time":"15:00"}`, mentorToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("mentor add availability: %d: %s", w.Code, w.Body.String())
	}
	resp := parseJSON(w)
	slotID, _ := resp["id"].(string)

	apptBody := `{"time_slot_id":"` + slotID + `","message":"想请教一个问题"}`
	w = doJSONWithToken("POST", "/api/v1/appointments", apptBody, studentToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create appointment: %d: %s", w.Code, w.Body.String())
	}
	apptResp := parseJSON(w)
	apptID, _ := apptResp["id"].(string)

	w = doJSONWithToken("GET", "/api/v1/appointments?role=mentor", "", mentorToken)
	if w.Code != http.StatusOK {
		t.Fatalf("get mentor appointments: %d", w.Code)
	}

	w = doJSONWithToken("GET", "/api/v1/appointments?role=student", "", studentToken)
	if w.Code != http.StatusOK {
		t.Fatalf("get student appointments: %d", w.Code)
	}

	w = doJSONWithToken("PUT", "/api/v1/appointments/"+apptID+"/accept", "", mentorToken)
	if w.Code != http.StatusOK {
		t.Fatalf("accept: %d: %s", w.Code, w.Body.String())
	}
	acceptResp := parseJSON(w)
	if status, _ := acceptResp["status"].(string); status != appointment.StatusAccepted {
		t.Errorf("status = %q, want accepted", status)
	}

	w = doJSONWithToken("GET", "/api/v1/users/mentor4@std.uestc.edu.cn", "", studentToken)
	if w.Code != http.StatusOK {
		t.Fatalf("student get accepted mentor detail: %d: %s", w.Code, w.Body.String())
	}
	userDetail := parseJSON(w)
	detailUser, ok := userDetail["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("detail user missing: %s", w.Body.String())
	}
	if detailUser["contact_info"] != "wechat-mentor4" {
		t.Errorf("accepted mentor contact_info = %v, want wechat-mentor4", detailUser["contact_info"])
	}
}

func TestHandler_Appointment_Reject(t *testing.T) {
	mentorToken := registerHelper(t, "rej-m3@std.uestc.edu.cn", "pass123", "RejMentor", "RM001")
	studentToken := registerHelper(t, "rej-s3@std.uestc.edu.cn", "pass123", "RejStudent", "RS001")

	w := doJSONWithToken("POST", "/api/v1/availability", `{"date":"2099-12-31","start_time":"16:00","end_time":"17:00"}`, mentorToken)
	resp := parseJSON(w)
	slotID, _ := resp["id"].(string)

	w = doJSONWithToken("POST", "/api/v1/appointments", `{"time_slot_id":"`+slotID+`","message":"问个问题"}`, studentToken)
	resp = parseJSON(w)
	apptID, _ := resp["id"].(string)

	w = doJSONWithToken("PUT", "/api/v1/appointments/"+apptID+"/reject", `{"reason":"时间不合适"}`, mentorToken)
	if w.Code != http.StatusOK {
		t.Fatalf("reject: %d: %s", w.Code, w.Body.String())
	}
	rejectResp := parseJSON(w)
	if status, _ := rejectResp["status"].(string); status != appointment.StatusRejected {
		t.Errorf("status = %q, want rejected", status)
	}
}

func TestHandler_Appointment_CannotBookOwnSlot(t *testing.T) {
	token := registerHelper(t, "selfbook3@std.uestc.edu.cn", "pass123", "SelfBook", "SB001")

	w := doJSONWithToken("POST", "/api/v1/availability", `{"date":"2099-12-31","start_time":"09:00","end_time":"10:00"}`, token)
	resp := parseJSON(w)
	slotID, _ := resp["id"].(string)

	w = doJSONWithToken("POST", "/api/v1/appointments", `{"time_slot_id":"`+slotID+`","message":"自己约自己"}`, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for self-booking, got %d", w.Code)
	}
}

// =============================================================================
// 权限
// =============================================================================

func TestHandler_Unauthorized(t *testing.T) {
	w := doJSON("GET", "/api/v1/profile", "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("profile without token: expected 401, got %d", w.Code)
	}

	w = doJSON("POST", "/api/v1/availability", `{"date":"2099-12-31","start_time":"10:00","end_time":"11:00"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("availability without token: expected 401, got %d", w.Code)
	}

	w = doJSON("POST", "/api/v1/appointments", `{"time_slot_id":"xxx","message":"hi"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("appointment without token: expected 401, got %d", w.Code)
	}
}

// =============================================================================
// 修改密码
// =============================================================================

func TestHandler_ChangePassword_FullFlow(t *testing.T) {
	token := registerHelper(t, "chpwd3@std.uestc.edu.cn", "oldpass123", "ChPwd", "CP001")

	// 1. 旧密码错误 → 400
	w := doJSONWithToken("PUT", "/api/v1/auth/change-password",
		`{"old_password":"wrongpassword","new_password":"newpass123"}`, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("wrong old password: expected 400, got %d: %s", w.Code, w.Body.String())
	}

	// 2. 修改密码成功 → 200
	w = doJSONWithToken("PUT", "/api/v1/auth/change-password",
		`{"old_password":"oldpass123","new_password":"newpass123"}`, token)
	if w.Code != http.StatusOK {
		t.Fatalf("change password: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["message"] != "密码修改成功" {
		t.Errorf("message = %v, want 密码修改成功", resp["message"])
	}

	// 3. 旧密码登录应失败 → 401
	w = doJSON("POST", "/api/v1/auth/login",
		`{"email":"chpwd3@std.uestc.edu.cn","password":"oldpass123"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("login with old password: expected 401, got %d", w.Code)
	}

	// 4. 新密码登录应成功 → 200
	w = doJSON("POST", "/api/v1/auth/login",
		`{"email":"chpwd3@std.uestc.edu.cn","password":"newpass123"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("login with new password: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	loginResp := parseJSON(w)
	if _, ok := loginResp["token"].(string); !ok {
		t.Error("login response should include token")
	}
}

func TestHandler_ChangePassword_NoToken(t *testing.T) {
	w := doJSON("PUT", "/api/v1/auth/change-password",
		`{"old_password":"old","new_password":"new"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("change password without token: expected 401, got %d", w.Code)
	}
}

func TestHandler_ChangePassword_Validation(t *testing.T) {
	token := registerHelper(t, "chpwdv3@std.uestc.edu.cn", "pass123456", "ChPwdV", "CV001")

	// 新密码太短
	w := doJSONWithToken("PUT", "/api/v1/auth/change-password",
		`{"old_password":"pass123456","new_password":"12345"}`, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("short password: expected 400, got %d: %s", w.Code, w.Body.String())
	}

	// 缺少 old_password
	w = doJSONWithToken("PUT", "/api/v1/auth/change-password",
		`{"new_password":"newpass123"}`, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing old_password: expected 400, got %d", w.Code)
	}

	// 缺少 new_password
	w = doJSONWithToken("PUT", "/api/v1/auth/change-password",
		`{"old_password":"pass123456"}`, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing new_password: expected 400, got %d", w.Code)
	}
}

// =============================================================================
// 注销账号
// =============================================================================

func TestHandler_DeleteAccount_Success(t *testing.T) {
	token := registerHelper(t, "del-ok@std.uestc.edu.cn", "pass123", "DelOK", "D001")

	w := doJSONWithToken("DELETE", "/api/v1/auth/account",
		`{"password":"pass123"}`, token)
	if w.Code != http.StatusOK {
		t.Fatalf("delete account: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := parseJSON(w)
	if resp["message"] != "账号已注销" {
		t.Errorf("message = %v, want 账号已注销", resp["message"])
	}

	// 注销后无法登录
	w = doJSON("POST", "/api/v1/auth/login",
		`{"email":"del-ok@std.uestc.edu.cn","password":"pass123"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("login after delete: expected 401, got %d", w.Code)
	}
}

func TestHandler_DeleteAccount_WrongPassword(t *testing.T) {
	token := registerHelper(t, "del-wp@std.uestc.edu.cn", "pass123", "DelWP", "D002")

	w := doJSONWithToken("DELETE", "/api/v1/auth/account",
		`{"password":"wrongpass"}`, token)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeleteAccount_NoToken(t *testing.T) {
	w := doJSON("DELETE", "/api/v1/auth/account", `{"password":"pass123"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no token: expected 401, got %d", w.Code)
	}
}

func TestHandler_DeleteAccount_NoPassword(t *testing.T) {
	token := registerHelper(t, "del-np@std.uestc.edu.cn", "pass123", "DelNP", "D003")

	w := doJSONWithToken("DELETE", "/api/v1/auth/account", `{}`, token)
	if w.Code != http.StatusBadRequest {
		t.Errorf("no password: expected 400, got %d", w.Code)
	}
}

func TestHandler_DeleteAccount_WithActiveAppointment(t *testing.T) {
	mentorToken := registerHelper(t, "del-mentor@std.uestc.edu.cn", "pass123", "DelMentor", "DM001")
	studentToken := registerHelper(t, "del-student@std.uestc.edu.cn", "pass123", "DelStudent", "DS001")

	// 创建空闲时间
	w := doJSONWithToken("POST", "/api/v1/availability",
		`{"date":"2099-12-31","start_time":"10:00","end_time":"11:00"}`, mentorToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("add availability: %d: %s", w.Code, w.Body.String())
	}
	resp := parseJSON(w)
	slotID, _ := resp["id"].(string)

	// 创建预约
	w = doJSONWithToken("POST", "/api/v1/appointments",
		`{"time_slot_id":"`+slotID+`","message":"test"}`, studentToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create appointment: %d: %s", w.Code, w.Body.String())
	}

	// mentor 有活跃预约，无法注销
	w = doJSONWithToken("DELETE", "/api/v1/auth/account",
		`{"password":"pass123"}`, mentorToken)
	if w.Code != http.StatusConflict {
		t.Errorf("mentor with active appointment: expected 409, got %d: %s", w.Code, w.Body.String())
	}

	// student 有活跃预约，无法注销
	w = doJSONWithToken("DELETE", "/api/v1/auth/account",
		`{"password":"pass123"}`, studentToken)
	if w.Code != http.StatusConflict {
		t.Errorf("student with active appointment: expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

// =============================================================================
// 头像上传
// =============================================================================

// minimalJPEG 返回一个最小的合法 JPEG 字节（1×1 像素）。
func minimalJPEG() []byte {
	return []byte{
		0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43,
		0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
		0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
		0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20,
		0x24, 0x2e, 0x27, 0x20, 0x22, 0x2c, 0x23, 0x1c, 0x1c, 0x28, 0x37, 0x29,
		0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27, 0x39, 0x3d, 0x38, 0x32,
		0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x1f, 0x00, 0x00,
		0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0xff, 0xc4, 0x00, 0xb5, 0x10, 0x00, 0x02, 0x01, 0x03,
		0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d,
		0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
		0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xa1, 0x08,
		0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62, 0x72,
		0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28,
		0x29, 0x2a, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45,
		0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
		0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74, 0x75,
		0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
		0x8a, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3,
		0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6,
		0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9,
		0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2,
		0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4,
		0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xff, 0xda, 0x00, 0x08, 0x01, 0x01,
		0x00, 0x00, 0x3f, 0x00, 0x7b, 0x94, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xd9,
	}
}

// doMultipartWithToken 构造 multipart/form-data 请求并发送。
func doMultipartWithToken(method, path, fieldName, fileName string, fileData []byte, token string) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		panic("failed to create form file: " + err.Error())
	}
	if _, err := io.Copy(part, bytes.NewReader(fileData)); err != nil {
		panic("failed to copy file data: " + err.Error())
	}
	writer.Close()

	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	getRouter().ServeHTTP(w, req)
	return w
}
