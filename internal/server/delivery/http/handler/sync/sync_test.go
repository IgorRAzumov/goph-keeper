package sync

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"goph-keeper/internal/contract"
	"goph-keeper/internal/logging"
	appsync "goph-keeper/internal/server/application/sync"
	"goph-keeper/internal/server/delivery/http/middleware"
	"goph-keeper/internal/server/domain/common"
	"goph-keeper/internal/server/domain/record/model"
	repomocks "goph-keeper/internal/server/domain/record/repository/mocks"
	recordsvc "goph-keeper/internal/server/domain/record/service"
	"log/slog"

	"go.uber.org/mock/gomock"
)

func testLogger() logging.Logger {
	return slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestSyncPullReturnsRecords(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().ListSince(gomock.Any(), "owner-1", int64(0)).Return([]*model.Record{
		{ID: "r1", OwnerID: "owner-1", Type: model.RecordTypeText, Ciphertext: []byte("x"), Version: 1},
	}, nil)

	usecase := appsync.NewUsecase(recordsvc.NewRecordService(repo))
	handler := Pull(testLogger(), usecase)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/sync/?since=0", nil)
	ctx := middleware.WithUserID(request.Context(), "owner-1")
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}

	var body struct {
		Records []contract.Record `json:"records"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Records) != 1 || body.Records[0].ID != "r1" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestSyncPullRejectsInvalidSince(t *testing.T) {
	t.Parallel()

	usecase := appsync.NewUsecase(recordsvc.NewRecordService(nil))
	handler := Pull(testLogger(), usecase)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/sync/?since=bad", nil)
	ctx := middleware.WithUserID(request.Context(), "owner-1")
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestSyncPushReturnsConflictStatus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	serverRecord := &model.Record{
		ID: "r1", OwnerID: "owner-1", Type: model.RecordTypeText,
		Ciphertext: []byte("server"), Version: 10,
	}

	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().UpdateBatch(gomock.Any(), "owner-1", gomock.Any()).Return([]*model.Record{serverRecord}, nil)

	usecase := appsync.NewUsecase(recordsvc.NewRecordService(repo))
	handler := Push(testLogger(), usecase)

	payload := `{"records":[{"id":"r1","type":"text","meta":"","ciphertext":"Y2xpZW50","version":3,"deleted":false}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sync/", bytes.NewBufferString(payload))
	ctx := middleware.WithUserID(request.Context(), "owner-1")
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", response.Code, response.Body.String())
	}
}

func TestSyncPushSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	repo.EXPECT().UpdateBatch(gomock.Any(), "owner-1", gomock.Any()).Return([]*model.Record{}, nil)

	usecase := appsync.NewUsecase(recordsvc.NewRecordService(repo))
	handler := Push(testLogger(), usecase)

	payload := `{"records":[{"id":"r1","type":"text","meta":"","ciphertext":"Y2xpZW50","version":3,"deleted":false}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sync/", bytes.NewBufferString(payload))
	ctx := middleware.WithUserID(request.Context(), "owner-1")
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", response.Code, response.Body.String())
	}

	var body struct {
		Conflicts []contract.Record `json:"conflicts"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Conflicts) != 0 {
		t.Fatalf("expected empty conflicts, got %+v", body.Conflicts)
	}
}

func TestSyncPushRejectsInvalidRecord(t *testing.T) {
	t.Parallel()

	usecase := appsync.NewUsecase(recordsvc.NewRecordService(nil))
	handler := Push(testLogger(), usecase)

	payload := `{"records":[{"id":"r1","type":"text","meta":"","ciphertext":"!!!","version":1,"deleted":false}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sync/", bytes.NewBufferString(payload))
	ctx := middleware.WithUserID(request.Context(), "owner-1")
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestSyncPushRejectsInvalidRecordType(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	repo := repomocks.NewMockRecordRepository(ctrl)
	usecase := appsync.NewUsecase(recordsvc.NewRecordService(repo))
	handler := Push(testLogger(), usecase)

	payload := `{"records":[{"id":"r1","type":"unknown","meta":"","ciphertext":"YQ==","version":1,"deleted":false}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sync/", bytes.NewBufferString(payload))
	ctx := middleware.WithUserID(request.Context(), "owner-1")
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte(common.ErrInvalidInput.Error())) {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
}
