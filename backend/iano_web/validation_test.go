package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// User 测试用的用户结构体
type User struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=0,max=150"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6,max=50"`
}

func TestBindAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		jsonBody   string
		wantError  bool
		fieldCheck string // 要检查的字段
	}{
		{
			name:      "Valid User",
			jsonBody:  `{"name":"John Doe","email":"john@example.com","age":25}`,
			wantError: false,
		},
		{
			name:       "Missing Name",
			jsonBody:   `{"email":"john@example.com","age":25}`,
			wantError:  true,
			fieldCheck: "name",
		},
		{
			name:       "Invalid Email",
			jsonBody:   `{"name":"John Doe","email":"invalid-email","age":25}`,
			wantError:  true,
			fieldCheck: "email",
		},
		{
			name:       "Name Too Short",
			jsonBody:   `{"name":"Jo","email":"john@example.com","age":25}`,
			wantError:  true,
			fieldCheck: "name",
		},
		{
			name:       "Age Too High",
			jsonBody:   `{"name":"John Doe","email":"john@example.com","age":200}`,
			wantError:  true,
			fieldCheck: "age",
		},
		{
			name:       "Negative Age",
			jsonBody:   `{"name":"John Doe","email":"john@example.com","age":-1}`,
			wantError:  true,
			fieldCheck: "age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			engine.POST("/users", func(c *Context) {
				var user User
				if err := c.BindAndValidate(&user); err != nil {
					c.JSON(400, map[string]interface{}{
						"error": err.Error(),
					})
					return
				}
				c.JSON(201, user)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/users", strings.NewReader(tt.jsonBody))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if tt.wantError {
				if w.Code != 400 {
					t.Errorf("Expected status 400 for validation error, got %d", w.Code)
				}
			} else {
				if w.Code != 201 {
					t.Errorf("Expected status 201 for valid request, got %d", w.Code)
				}
			}
		})
	}
}

func TestBindAndValidateLogin(t *testing.T) {
	tests := []struct {
		name      string
		jsonBody  string
		wantError bool
	}{
		{
			name:      "Valid Login",
			jsonBody:  `{"username":"john","password":"123456"}`,
			wantError: false,
		},
		{
			name:      "Password Too Short",
			jsonBody:  `{"username":"john","password":"123"}`,
			wantError: true,
		},
		{
			name:      "Missing Password",
			jsonBody:  `{"username":"john"}`,
			wantError: true,
		},
		{
			name:      "Username Too Long",
			jsonBody:  `{"username":"thisisaverylongusername","password":"123456"}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New()
			engine.POST("/login", func(c *Context) {
				var req LoginRequest
				if err := c.BindAndValidate(&req); err != nil {
					c.JSON(400, map[string]interface{}{
						"error": err.Error(),
					})
					return
				}
				c.JSON(200, map[string]string{
					"message": "Login successful",
				})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/login", strings.NewReader(tt.jsonBody))
			req.Header.Set("Content-Type", "application/json")
			engine.ServeHTTP(w, req)

			if tt.wantError {
				if w.Code != 400 {
					t.Errorf("Expected status 400 for validation error, got %d", w.Code)
				}
			} else {
				if w.Code != 200 {
					t.Errorf("Expected status 200 for valid request, got %d", w.Code)
				}
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	// 直接测试 FormatValidationErrors 函数
	var user User
	err := Validator().Struct(&user)

	if err == nil {
		t.Fatal("Expected validation error for empty user")
	}

	errors := FormatValidationErrors(err)

	if len(errors) == 0 {
		t.Error("Expected validation errors, got none")
	}

	// 检查是否有 name 字段的错误
	hasNameError := false
	for _, e := range errors {
		if e.Field == "Name" { // 注意字段名是大写的
			hasNameError = true
			break
		}
	}
	if !hasNameError {
		t.Logf("Available fields: %v", getErrorFields(errors))
		t.Error("Expected validation error for 'Name' field")
	}
}

func getErrorFields(errors ValidationErrors) []string {
	var fields []string
	for _, e := range errors {
		fields = append(fields, e.Field)
	}
	return fields
}

func TestValidationErrorsError(t *testing.T) {
	errors := ValidationErrors{
		{Field: "name", Message: "该字段是必填项"},
		{Field: "email", Message: "请输入有效的邮箱地址"},
	}

	errMsg := errors.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	if !strings.Contains(errMsg, "name") {
		t.Error("Expected error message to contain 'name'")
	}

	if !strings.Contains(errMsg, "email") {
		t.Error("Expected error message to contain 'email'")
	}
}

func TestValidatorSingleton(t *testing.T) {
	v1 := Validator()
	v2 := Validator()

	if v1 != v2 {
		t.Error("Validator() should return the same instance")
	}
}
