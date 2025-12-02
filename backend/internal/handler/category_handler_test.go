package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/handler/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Note: CategoryHandler uses the same MockTaskService defined in task_handler_test.go

// =============================================================================
// CategoryHandler Constructor Tests
// =============================================================================

func TestNewCategoryHandler(t *testing.T) {
	mockService := new(MockTaskService)
	handler := NewCategoryHandler(mockService)

	assert.NotNil(t, handler)
}

// =============================================================================
// Rename Tests
// =============================================================================

func TestCategoryHandler_Rename_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	mockService.On("RenameCategory", mock.Anything, "user-123", "Old Category", "New Category").
		Return(5, nil)

	reqBody := map[string]string{
		"old_name": "Old Category",
		"new_name": "New Category",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Category renamed successfully", response["message"])
	assert.Equal(t, float64(5), response["updated_count"])
}

func TestCategoryHandler_Rename_Unauthenticated(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", handler.Rename)

	reqBody := map[string]string{
		"old_name": "Old Category",
		"new_name": "New Category",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "RenameCategory")
}

func TestCategoryHandler_Rename_InvalidJSON(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "RenameCategory")
}

func TestCategoryHandler_Rename_MissingOldName(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	reqBody := map[string]string{
		"new_name": "New Category",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "RenameCategory")
}

func TestCategoryHandler_Rename_MissingNewName(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	reqBody := map[string]string{
		"old_name": "Old Category",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "RenameCategory")
}

func TestCategoryHandler_Rename_BindingValidation_EmptyNewName(t *testing.T) {
	// Tests that Gin's binding validation rejects empty new_name before reaching service
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	reqBody := map[string]string{
		"old_name": "Old",
		"new_name": "",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Gin binding validation fails because new_name is required (binding:"required")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "RenameCategory")
}

func TestCategoryHandler_Rename_ServiceError(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	mockService.On("RenameCategory", mock.Anything, "user-123", "Old", "New").
		Return(0, domain.NewInternalError("database error", nil))

	reqBody := map[string]string{
		"old_name": "Old",
		"new_name": "New",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Rename_ZeroUpdated(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.PUT("/categories/rename", testutil.WithAuthContext(router, "user-123", handler.Rename))

	mockService.On("RenameCategory", mock.Anything, "user-123", "NonExistent", "New").
		Return(0, nil)

	reqBody := map[string]string{
		"old_name": "NonExistent",
		"new_name": "New",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/categories/rename", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["updated_count"])
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestCategoryHandler_Delete_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.DELETE("/categories/:name", testutil.WithAuthContext(router, "user-123", handler.Delete))

	mockService.On("DeleteCategory", mock.Anything, "user-123", "Work").
		Return(3, nil)

	req := httptest.NewRequest("DELETE", "/categories/Work", nil)

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Category deleted successfully", response["message"])
	assert.Equal(t, float64(3), response["updated_count"])
}

func TestCategoryHandler_Delete_Unauthenticated(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.DELETE("/categories/:name", handler.Delete)

	req := httptest.NewRequest("DELETE", "/categories/Work", nil)

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "DeleteCategory")
}

func TestCategoryHandler_Delete_ServiceError(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.DELETE("/categories/:name", testutil.WithAuthContext(router, "user-123", handler.Delete))

	mockService.On("DeleteCategory", mock.Anything, "user-123", "Work").
		Return(0, domain.NewInternalError("database error", nil))

	req := httptest.NewRequest("DELETE", "/categories/Work", nil)

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Delete_NotFoundCategory(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.DELETE("/categories/:name", testutil.WithAuthContext(router, "user-123", handler.Delete))

	mockService.On("DeleteCategory", mock.Anything, "user-123", "NonExistent").
		Return(0, nil)

	req := httptest.NewRequest("DELETE", "/categories/NonExistent", nil)

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["updated_count"])
}

func TestCategoryHandler_Delete_URLEncodedName(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.DELETE("/categories/:name", testutil.WithAuthContext(router, "user-123", handler.Delete))

	// URL-encoded "Work Tasks" = "Work%20Tasks"
	mockService.On("DeleteCategory", mock.Anything, "user-123", "Work Tasks").
		Return(2, nil)

	req := httptest.NewRequest("DELETE", "/categories/Work%20Tasks", nil)

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCategoryHandler_Delete_EmptyPathParam_Returns404(t *testing.T) {
	// Tests that Gin router returns 404 for empty path parameter
	// The handler is never reached because the route doesn't match
	router, mockService := setupTaskTest()
	handler := NewCategoryHandler(mockService)

	router.DELETE("/categories/:name", testutil.WithAuthContext(router, "user-123", handler.Delete))

	req := httptest.NewRequest("DELETE", "/categories/", nil)

	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Router returns 404 because "/categories/" doesn't match "/categories/:name"
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertNotCalled(t, "DeleteCategory")
}

func TestCategoryHandler_Delete_SpecialCharacters(t *testing.T) {
	testCases := []struct {
		name         string
		categoryName string
		encodedPath  string
	}{
		{"ampersand", "Work & Personal", "/categories/Work%20%26%20Personal"},
		{"hash", "Project#1", "/categories/Project%231"},
		{"plus", "C++", "/categories/C%2B%2B"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, mockService := setupTaskTest()
			handler := NewCategoryHandler(mockService)

			router.DELETE("/categories/:name", testutil.WithAuthContext(router, "user-123", handler.Delete))

			mockService.On("DeleteCategory", mock.Anything, "user-123", tc.categoryName).
				Return(1, nil)

			req := httptest.NewRequest("DELETE", tc.encodedPath, nil)

			w := testutil.NewResponseRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
