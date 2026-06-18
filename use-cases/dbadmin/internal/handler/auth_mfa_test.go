package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/security/authn"
	kvstore "github.com/spcent/plumego/store/kv"

	"dbadmin/internal/domain/mfa"
	"dbadmin/internal/domain/session"
)

func newAuthHandlerForTest(t *testing.T) AuthHandler {
	t.Helper()
	sessKV, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create session kv: %v", err)
	}
	t.Cleanup(func() { _ = sessKV.Close() })
	mfaKV, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create mfa kv: %v", err)
	}
	t.Cleanup(func() { _ = mfaKV.Close() })

	return AuthHandler{
		AdminUser:     "admin",
		AdminPassword: "correct-password",
		AdminRole:     "admin",
		SessionTTL:    time.Hour,
		Sessions:      session.NewStore(sessKV, time.Hour),
		MFA:           mfa.NewStore(mfaKV),
		Logger:        plumelog.NewLogger(),
	}
}

func doJSON(t *testing.T, h http.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v (body=%s)", err, rec.Body.String())
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		t.Fatalf("decode data: %v (body=%s)", err, rec.Body.String())
	}
}

// --- Login without MFA enrolled ---

func TestLogin_noMFA_issuesSessionDirectly(t *testing.T) {
	h := newAuthHandlerForTest(t)
	rec := doJSON(t, h.Login, http.MethodPost, "/api/auth/login", loginRequest{
		Username: "admin",
		Password: "correct-password",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("Login status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
	var resp loginResponse
	decodeData(t, rec, &resp)
	if resp.Status != "ok" {
		t.Fatalf("Login response status = %q, want %q", resp.Status, "ok")
	}
	if resp.ChallengeToken != "" {
		t.Fatal("Login without MFA must not return a challenge token")
	}
	found := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "dbadmin_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("Login without MFA must set a session cookie")
	}
}

func TestLogin_wrongPassword_rejected(t *testing.T) {
	h := newAuthHandlerForTest(t)
	rec := doJSON(t, h.Login, http.MethodPost, "/api/auth/login", loginRequest{
		Username: "admin",
		Password: "wrong",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Login status = %d, want 401", rec.Code)
	}
}

// --- Full MFA enroll -> confirm -> login challenge -> verify flow ---

func withPrincipal(r *http.Request, username string) *http.Request {
	ctx := authn.WithPrincipal(r.Context(), &authn.Principal{Subject: username})
	return r.WithContext(ctx)
}

func doJSONAuthenticated(t *testing.T, h http.HandlerFunc, method, path, username string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req = withPrincipal(req, username)
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestMFAEnrollConfirmLoginFlow(t *testing.T) {
	h := newAuthHandlerForTest(t)

	// 1. Enroll: get a secret, MFA must not be enabled yet.
	enrollRec := doJSONAuthenticated(t, h.MFAEnroll, http.MethodPost, "/api/auth/mfa/enroll", "admin", nil)
	if enrollRec.Code != http.StatusOK {
		t.Fatalf("MFAEnroll status = %d, want 200, body=%s", enrollRec.Code, enrollRec.Body.String())
	}
	var enrollResp mfaEnrollResponse
	decodeData(t, enrollRec, &enrollResp)
	if enrollResp.Secret == "" || enrollResp.OTPAuthURI == "" {
		t.Fatal("MFAEnroll must return a non-empty secret and otpauth URI")
	}
	if h.MFA.IsEnabled("admin") {
		t.Fatal("MFA must not be enabled immediately after enroll")
	}

	// Login should still succeed directly since MFA isn't confirmed/enabled.
	preConfirmLogin := doJSON(t, h.Login, http.MethodPost, "/api/auth/login", loginRequest{
		Username: "admin",
		Password: "correct-password",
	})
	var preConfirmResp loginResponse
	decodeData(t, preConfirmLogin, &preConfirmResp)
	if preConfirmResp.Status != "ok" {
		t.Fatalf("Login before MFA confirmation should succeed directly, got status=%q", preConfirmResp.Status)
	}

	// 2. Confirm enrollment with a valid code.
	code, err := mfa.GenerateCode(enrollResp.Secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	confirmRec := doJSONAuthenticated(t, h.MFAConfirm, http.MethodPost, "/api/auth/mfa/confirm", "admin", mfaConfirmRequest{Code: code})
	if confirmRec.Code != http.StatusOK {
		t.Fatalf("MFAConfirm status = %d, want 200, body=%s", confirmRec.Code, confirmRec.Body.String())
	}
	if !h.MFA.IsEnabled("admin") {
		t.Fatal("MFA must be enabled after a successful confirm")
	}

	// 3. Login now requires MFA: no cookie, a challenge token instead.
	loginRec := doJSON(t, h.Login, http.MethodPost, "/api/auth/login", loginRequest{
		Username: "admin",
		Password: "correct-password",
	})
	if loginRec.Code != http.StatusOK {
		t.Fatalf("Login (mfa required) status = %d, want 200, body=%s", loginRec.Code, loginRec.Body.String())
	}
	var loginResp loginResponse
	decodeData(t, loginRec, &loginResp)
	if loginResp.Status != "mfa_required" {
		t.Fatalf("Login response status = %q, want %q", loginResp.Status, "mfa_required")
	}
	if loginResp.ChallengeToken == "" {
		t.Fatal("Login response must include a challenge token when mfa is required")
	}
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == "dbadmin_session" && c.Value != "" {
			t.Fatal("Login must not set a session cookie when mfa is required")
		}
	}

	// 4. Verify with a fresh code completes login and sets a session cookie.
	verifyCode, err := mfa.GenerateCode(enrollResp.Secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	verifyRec := doJSON(t, h.MFAVerify, http.MethodPost, "/api/auth/mfa/verify", mfaVerifyRequest{
		ChallengeToken: loginResp.ChallengeToken,
		Code:           verifyCode,
	})
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("MFAVerify status = %d, want 200, body=%s", verifyRec.Code, verifyRec.Body.String())
	}
	var verifyResp loginResponse
	decodeData(t, verifyRec, &verifyResp)
	if verifyResp.Status != "ok" || verifyResp.User != "admin" {
		t.Fatalf("MFAVerify response = %+v, want status=ok user=admin", verifyResp)
	}
	found := false
	for _, c := range verifyRec.Result().Cookies() {
		if c.Name == "dbadmin_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("MFAVerify must set a session cookie on success")
	}
}

func TestMFAVerify_challengeIsSingleUse(t *testing.T) {
	h := newAuthHandlerForTest(t)
	secret, _, err := h.MFA.StartEnrollment("admin")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	code, _ := mfa.GenerateCode(secret, time.Now())
	if err := h.MFA.ConfirmEnrollment("admin", code); err != nil {
		t.Fatalf("ConfirmEnrollment error = %v", err)
	}

	token, err := h.MFA.CreateChallenge("admin")
	if err != nil {
		t.Fatalf("CreateChallenge error = %v", err)
	}
	verifyCode, _ := mfa.GenerateCode(secret, time.Now())

	first := doJSON(t, h.MFAVerify, http.MethodPost, "/api/auth/mfa/verify", mfaVerifyRequest{ChallengeToken: token, Code: verifyCode})
	if first.Code != http.StatusOK {
		t.Fatalf("first MFAVerify status = %d, want 200, body=%s", first.Code, first.Body.String())
	}

	second := doJSON(t, h.MFAVerify, http.MethodPost, "/api/auth/mfa/verify", mfaVerifyRequest{ChallengeToken: token, Code: verifyCode})
	if second.Code == http.StatusOK {
		t.Fatal("reusing a consumed mfa challenge token must be rejected")
	}
}

func TestMFAVerify_wrongCode_rejected(t *testing.T) {
	h := newAuthHandlerForTest(t)
	secret, _, err := h.MFA.StartEnrollment("admin")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	code, _ := mfa.GenerateCode(secret, time.Now())
	if err := h.MFA.ConfirmEnrollment("admin", code); err != nil {
		t.Fatalf("ConfirmEnrollment error = %v", err)
	}
	token, err := h.MFA.CreateChallenge("admin")
	if err != nil {
		t.Fatalf("CreateChallenge error = %v", err)
	}
	rec := doJSON(t, h.MFAVerify, http.MethodPost, "/api/auth/mfa/verify", mfaVerifyRequest{ChallengeToken: token, Code: "000000"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("MFAVerify with wrong code status = %d, want 401", rec.Code)
	}
}

func TestMFAVerify_unknownToken_rejected(t *testing.T) {
	h := newAuthHandlerForTest(t)
	rec := doJSON(t, h.MFAVerify, http.MethodPost, "/api/auth/mfa/verify", mfaVerifyRequest{ChallengeToken: "does-not-exist", Code: "123456"})
	if rec.Code == http.StatusOK {
		t.Fatal("MFAVerify with an unknown challenge token must be rejected")
	}
}

// --- Disable ---

func TestMFADisable_requiresCorrectPassword(t *testing.T) {
	h := newAuthHandlerForTest(t)
	secret, _, err := h.MFA.StartEnrollment("admin")
	if err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	code, _ := mfa.GenerateCode(secret, time.Now())
	if err := h.MFA.ConfirmEnrollment("admin", code); err != nil {
		t.Fatalf("ConfirmEnrollment error = %v", err)
	}

	wrongRec := doJSONAuthenticated(t, h.MFADisable, http.MethodPost, "/api/auth/mfa/disable", "admin", mfaDisableRequest{Password: "wrong"})
	if wrongRec.Code != http.StatusUnauthorized {
		t.Fatalf("MFADisable with wrong password status = %d, want 401", wrongRec.Code)
	}
	if !h.MFA.IsEnabled("admin") {
		t.Fatal("MFA must remain enabled after a failed disable attempt")
	}

	okRec := doJSONAuthenticated(t, h.MFADisable, http.MethodPost, "/api/auth/mfa/disable", "admin", mfaDisableRequest{Password: "correct-password"})
	if okRec.Code != http.StatusOK {
		t.Fatalf("MFADisable with correct password status = %d, want 200, body=%s", okRec.Code, okRec.Body.String())
	}
	if h.MFA.IsEnabled("admin") {
		t.Fatal("MFA must be disabled after a successful disable")
	}
}

func TestMFAStatus(t *testing.T) {
	h := newAuthHandlerForTest(t)
	rec := doJSONAuthenticated(t, h.MFAStatus, http.MethodGet, "/api/auth/mfa", "admin", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("MFAStatus status = %d, want 200", rec.Code)
	}
	var resp mfaStatusResponse
	decodeData(t, rec, &resp)
	if resp.Enabled {
		t.Fatal("MFAStatus should report disabled before enrollment")
	}
}

func TestMFAConfirm_wrongCode_doesNotEnable(t *testing.T) {
	h := newAuthHandlerForTest(t)
	if _, _, err := h.MFA.StartEnrollment("admin"); err != nil {
		t.Fatalf("StartEnrollment error = %v", err)
	}
	rec := doJSONAuthenticated(t, h.MFAConfirm, http.MethodPost, "/api/auth/mfa/confirm", "admin", mfaConfirmRequest{Code: "000000"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("MFAConfirm with wrong code status = %d, want 401", rec.Code)
	}
	if h.MFA.IsEnabled("admin") {
		t.Fatal("MFA must not be enabled after a failed confirm")
	}
}
