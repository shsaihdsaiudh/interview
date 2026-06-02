package handler

import (
	"context"
	"encoding/json"
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

	w := doJSON("GET", "/api/v1/users", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d", w.Code)
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
