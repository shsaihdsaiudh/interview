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

	userSvc := application.NewUserService(repo, repo)
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
			auth.POST("/register", userH.Register)
			auth.GET("/verify-email", userH.VerifyEmail)
			auth.POST("/login", userH.Login)

			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuth())
			{
				authRequired.GET("/me", userH.Me)
			}
		}

		v1.GET("/users", userH.ListUsers)
		v1.GET("/users/:id", userH.GetUser)

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

func registerAndVerify(t *testing.T, email, password, nickname, studentID string) string {
	t.Helper()

	body := `{"email":"` + email + `","password":"` + password + `","nickname":"` + nickname + `","student_id":"` + studentID + `"}`
	w := doJSON("POST", "/api/v1/auth/register", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	u, err := testRepo.FindByEmail(email)
	if err != nil {
		t.Fatalf("find user after register: %v", err)
	}

	verifyURL := "/api/v1/auth/verify-email?token=" + u.VerifyToken
	w = doJSON("GET", verifyURL, "")
	if w.Code != http.StatusOK {
		t.Fatalf("verify email: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	loginBody := `{"email":"` + email + `","password":"` + password + `"}`
	w = doJSON("POST", "/api/v1/auth/login", loginBody)
	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatal("no token in login response")
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

func TestHandler_Register_InvalidEmail(t *testing.T) {
	body := `{"email":"alice@gmail.com","password":"123456","nickname":"Alice","student_id":"S001"}`
	w := doJSON("POST", "/api/v1/auth/register", body)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-edu email, got %d", w.Code)
	}
}

func TestHandler_Register_Duplicate(t *testing.T) {
	email := "dup2@school.edu"
	body := `{"email":"` + email + `","password":"123456","nickname":"Dup","student_id":"D001"}`

	w := doJSON("POST", "/api/v1/auth/register", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("first register: %d: %s", w.Code, w.Body.String())
	}

	w = doJSON("POST", "/api/v1/auth/register", body)
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate: expected 409, got %d", w.Code)
	}
}

func TestHandler_Login_NotVerified(t *testing.T) {
	email := "unverified2@school.edu"
	body := `{"email":"` + email + `","password":"123456","nickname":"UV","student_id":"U001"}`

	w := doJSON("POST", "/api/v1/auth/register", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d", w.Code)
	}

	w = doJSON("POST", "/api/v1/auth/login", `{"email":"`+email+`","password":"123456"}`)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unverified login, got %d", w.Code)
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	email := "wrongpw2@school.edu"
	token := registerAndVerify(t, email, "correct", "WP", "W001")
	_ = token

	w := doJSON("POST", "/api/v1/auth/login", `{"email":"`+email+`","password":"wrong"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", w.Code)
	}
}

func TestHandler_RegisterLoginMe_FullFlow(t *testing.T) {
	token := registerAndVerify(t, "fullflow2@school.edu", "pass123", "FullFlow", "F001")

	w := doJSONWithToken("GET", "/api/v1/auth/me", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("me: %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(w)
	if resp["email"] != "fullflow2@school.edu" {
		t.Errorf("email = %v", resp["email"])
	}
}

// =============================================================================
// 个人信息
// =============================================================================

func TestHandler_GetProfile(t *testing.T) {
	token := registerAndVerify(t, "profile2@school.edu", "pass123", "Profile", "P001")

	w := doJSONWithToken("GET", "/api/v1/profile", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("get profile: %d", w.Code)
	}
}

func TestHandler_UpdateProfile(t *testing.T) {
	token := registerAndVerify(t, "updprof2@school.edu", "pass123", "OldName", "U001")

	body := `{"nickname":"NewName","department":"Math"}`
	w := doJSONWithToken("PUT", "/api/v1/profile", body, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update profile: %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ListUsers(t *testing.T) {
	registerAndVerify(t, "list1a@school.edu", "pass123", "L1", "L001")
	registerAndVerify(t, "list2a@school.edu", "pass123", "L2", "L002")

	w := doJSON("GET", "/api/v1/users", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d", w.Code)
	}
}

// =============================================================================
// 空闲时间
// =============================================================================

func TestHandler_Availability_CRUD(t *testing.T) {
	token := registerAndVerify(t, "avail2@school.edu", "pass123", "Avail", "A001")

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
	mentorToken := registerAndVerify(t, "mentor3@school.edu", "pass123", "MentorFlow", "MF001")
	studentToken := registerAndVerify(t, "student3@school.edu", "pass123", "StudentFlow", "SF001")

	w := doJSONWithToken("POST", "/api/v1/availability", `{"date":"2099-12-31","start_time":"14:00","end_time":"15:00"}`, mentorToken)
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
}

func TestHandler_Appointment_Reject(t *testing.T) {
	mentorToken := registerAndVerify(t, "rej-m2@school.edu", "pass123", "RejMentor", "RM001")
	studentToken := registerAndVerify(t, "rej-s2@school.edu", "pass123", "RejStudent", "RS001")

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
	token := registerAndVerify(t, "selfbook2@school.edu", "pass123", "SelfBook", "SB001")

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
