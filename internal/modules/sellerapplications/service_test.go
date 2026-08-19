package sellerapplications

import (
	"context"
	"errors"
	"testing"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

// --- Mocks ---

type mockSellerApplicationRepository struct {
	create          func(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error)
	getByID         func(ctx context.Context, id int64) (db.SellerApplication, error)
	getPendingByUserID func(ctx context.Context, userID int64) (db.SellerApplication, error)
	listByUser      func(ctx context.Context, userID int64) ([]db.SellerApplication, error)
	listPending     func(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error)
	countPending    func(ctx context.Context) (int64, error)
	updateStatus    func(ctx context.Context, id int64, status string) (db.SellerApplication, error)
}

func (m *mockSellerApplicationRepository) Create(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error) {
	return m.create(ctx, arg)
}
func (m *mockSellerApplicationRepository) GetByID(ctx context.Context, id int64) (db.SellerApplication, error) {
	return m.getByID(ctx, id)
}
func (m *mockSellerApplicationRepository) GetPendingByUserID(ctx context.Context, userID int64) (db.SellerApplication, error) {
	return m.getPendingByUserID(ctx, userID)
}
func (m *mockSellerApplicationRepository) ListByUser(ctx context.Context, userID int64) ([]db.SellerApplication, error) {
	return m.listByUser(ctx, userID)
}
func (m *mockSellerApplicationRepository) ListPending(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error) {
	return m.listPending(ctx, limit, offset)
}
func (m *mockSellerApplicationRepository) CountPending(ctx context.Context) (int64, error) {
	return m.countPending(ctx)
}
func (m *mockSellerApplicationRepository) UpdateStatus(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
	return m.updateStatus(ctx, id, status)
}

type mockUserRepository struct {
	updateUserRole    func(ctx context.Context, id int64, role db.UserRole) (db.User, error)
	createStoreWithRole func(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error)
	getUserByID       func(ctx context.Context, id int64) (db.User, error)
}

func (m *mockUserRepository) UpdateUserRole(ctx context.Context, id int64, role db.UserRole) (db.User, error) {
	return m.updateUserRole(ctx, id, role)
}
func (m *mockUserRepository) CreateStoreWithRole(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error) {
	return m.createStoreWithRole(ctx, userID, name, description, logoURL)
}
func (m *mockUserRepository) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return m.getUserByID(ctx, id)
}

type mockEmailSender struct {
	sendApproved          func(to, username, storeName string) error
	sendRejected          func(to, username, reason string) error
	sendNewApplication    func(to, username, storeName, description string) error
}

func (m *mockEmailSender) SendSellerApplicationApproved(to, username, storeName string) error {
	if m.sendApproved != nil {
		return m.sendApproved(to, username, storeName)
	}
	return nil
}
func (m *mockEmailSender) SendSellerApplicationRejected(to, username, reason string) error {
	if m.sendRejected != nil {
		return m.sendRejected(to, username, reason)
	}
	return nil
}
func (m *mockEmailSender) SendNewApplicationNotification(to, username, storeName, description string) error {
	if m.sendNewApplication != nil {
		return m.sendNewApplication(to, username, storeName, description)
	}
	return nil
}

func newTestService(repo SellerApplicationRepository, userRepo UserRepository, email EmailSender) SellerApplicationService {
	return &sellerApplicationService{repo: repo, userRepo: userRepo, email: email}
}

func testApplication() db.SellerApplication {
	return db.SellerApplication{
		ID:        1,
		UserID:    1,
		StoreName: "Test Store",
		Status:    "pending",
		Description: pgtype.Text{String: "Test description", Valid: true},
	}
}

func testUser() db.User {
	return db.User{ID: 1, Email: "test@example.com", Username: "testuser"}
}

// --- Create ---

func TestCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("empty store name", func(t *testing.T) {
		svc := newTestService(&mockSellerApplicationRepository{}, nil, &mockEmailSender{})
		_, err := svc.Create(ctx, 1, "", nil)
		if err == nil {
			t.Fatal("expected error for empty store name")
		}
	})

	t.Run("pending application exists", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			getPendingByUserID: func(ctx context.Context, userID int64) (db.SellerApplication, error) {
				return testApplication(), nil
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Create(ctx, 1, "New Store", nil)
		if err != ErrAlreadyPending {
			t.Fatalf("expected ErrAlreadyPending, got %v", err)
		}
	})

	t.Run("create fails", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			getPendingByUserID: func(ctx context.Context, userID int64) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("not found")
			},
			create: func(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("create error")
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Create(ctx, 1, "New Store", nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		app := testApplication()
		user := testUser()
		repo := &mockSellerApplicationRepository{
			getPendingByUserID: func(ctx context.Context, userID int64) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("not found")
			},
			create: func(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error) {
				return app, nil
			},
		}
		userRepo := &mockUserRepository{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return user, nil
			},
		}
		svc := newTestService(repo, userRepo, &mockEmailSender{})
		result, err := svc.Create(ctx, 1, "Test Store", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != 1 {
			t.Fatalf("expected ID 1, got %d", result.ID)
		}
		if result.StoreName != "Test Store" {
			t.Fatalf("expected store name Test Store, got %s", result.StoreName)
		}
	})

	t.Run("success with description", func(t *testing.T) {
		app := testApplication()
		desc := "My awesome store"
		user := testUser()
		repo := &mockSellerApplicationRepository{
			getPendingByUserID: func(ctx context.Context, userID int64) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("not found")
			},
			create: func(ctx context.Context, arg db.CreateSellerApplicationParams) (db.SellerApplication, error) {
				if arg.StoreName != "Test Store" {
					t.Fatalf("expected store name Test Store, got %s", arg.StoreName)
				}
				if !arg.Description.Valid || arg.Description.String != desc {
					t.Fatalf("expected description '%s'", desc)
				}
				return app, nil
			},
		}
		userRepo := &mockUserRepository{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return user, nil
			},
		}
		svc := newTestService(repo, userRepo, &mockEmailSender{})
		_, err := svc.Create(ctx, 1, "Test Store", &desc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// --- ListMine ---

func TestListMine(t *testing.T) {
	ctx := context.Background()

	t.Run("success with items", func(t *testing.T) {
		app := testApplication()
		repo := &mockSellerApplicationRepository{
			listByUser: func(ctx context.Context, userID int64) ([]db.SellerApplication, error) {
				return []db.SellerApplication{app}, nil
			},
		}
		svc := newTestService(repo, nil, nil)
		result, err := svc.ListMine(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 application, got %d", len(result))
		}
		if result[0].StoreName != "Test Store" {
			t.Fatalf("expected store name Test Store, got %s", result[0].StoreName)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			listByUser: func(ctx context.Context, userID int64) ([]db.SellerApplication, error) {
				return []db.SellerApplication{}, nil
			},
		}
		svc := newTestService(repo, nil, nil)
		result, err := svc.ListMine(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("expected empty list, got %d", len(result))
		}
	})

	t.Run("db error", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			listByUser: func(ctx context.Context, userID int64) ([]db.SellerApplication, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil, nil)
		_, err := svc.ListMine(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- ListPending ---

func TestListPending(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		app := testApplication()
		repo := &mockSellerApplicationRepository{
			listPending: func(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error) {
				return []db.SellerApplication{app}, nil
			},
			countPending: func(ctx context.Context) (int64, error) {
				return 1, nil
			},
		}
		svc := newTestService(repo, nil, nil)
		result, total, err := svc.ListPending(ctx, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 application, got %d", len(result))
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
	})

	t.Run("empty pending", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			listPending: func(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error) {
				return []db.SellerApplication{}, nil
			},
			countPending: func(ctx context.Context) (int64, error) {
				return 0, nil
			},
		}
		svc := newTestService(repo, nil, nil)
		result, total, err := svc.ListPending(ctx, 1, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("expected empty list, got %d", len(result))
		}
		if total != 0 {
			t.Fatalf("expected total 0, got %d", total)
		}
	})

	t.Run("list pending fails", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			listPending: func(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error) {
				return nil, errors.New("db error")
			},
		}
		svc := newTestService(repo, nil, nil)
		_, _, err := svc.ListPending(ctx, 1, 20)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("count pending fails", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			listPending: func(ctx context.Context, limit, offset int32) ([]db.SellerApplication, error) {
				return []db.SellerApplication{}, nil
			},
			countPending: func(ctx context.Context) (int64, error) {
				return 0, errors.New("count error")
			},
		}
		svc := newTestService(repo, nil, nil)
		_, _, err := svc.ListPending(ctx, 1, 20)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

// --- Approve ---

func TestApprove(t *testing.T) {
	ctx := context.Background()

	t.Run("application not found", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("not found")
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Approve(ctx, 999)
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("already processed", func(t *testing.T) {
		app := testApplication()
		app.Status = "approved"
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Approve(ctx, 1)
		if err != ErrAlreadyProcessed {
			t.Fatalf("expected ErrAlreadyProcessed, got %v", err)
		}
	})

	t.Run("update status fails", func(t *testing.T) {
		app := testApplication()
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
			updateStatus: func(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("update error")
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Approve(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("promote user fails", func(t *testing.T) {
		app := testApplication()
		user := testUser()
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
			updateStatus: func(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
				app.Status = "approved"
				return app, nil
			},
		}
		userRepo := &mockUserRepository{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return user, nil
			},
			createStoreWithRole: func(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error) {
				return db.Store{}, errors.New("promote error")
			},
		}
		svc := newTestService(repo, userRepo, &mockEmailSender{})
		_, err := svc.Approve(ctx, 1)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		app := testApplication()
		user := testUser()
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
			updateStatus: func(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
				app.Status = "approved"
				return app, nil
			},
		}
		userRepo := &mockUserRepository{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return user, nil
			},
			createStoreWithRole: func(ctx context.Context, userID int64, name string, description, logoURL *string) (db.Store, error) {
				user.Role = db.UserRoleSeller
				return db.Store{ID: 1, UserID: userID, Name: name}, nil
			},
		}
		email := &mockEmailSender{}
		svc := newTestService(repo, userRepo, email)
		result, err := svc.Approve(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "approved" {
			t.Fatalf("expected status approved, got %s", result.Status)
		}
	})
}

// --- Reject ---

func TestReject(t *testing.T) {
	ctx := context.Background()

	t.Run("application not found", func(t *testing.T) {
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("not found")
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Reject(ctx, 999, "")
		if err != ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("already processed", func(t *testing.T) {
		app := testApplication()
		app.Status = "rejected"
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Reject(ctx, 1, "")
		if err != ErrAlreadyProcessed {
			t.Fatalf("expected ErrAlreadyProcessed, got %v", err)
		}
	})

	t.Run("update status fails", func(t *testing.T) {
		app := testApplication()
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
			updateStatus: func(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
				return db.SellerApplication{}, errors.New("update error")
			},
		}
		svc := newTestService(repo, nil, &mockEmailSender{})
		_, err := svc.Reject(ctx, 1, "")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		app := testApplication()
		user := testUser()
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
			updateStatus: func(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
				app.Status = "rejected"
				return app, nil
			},
		}
		userRepo := &mockUserRepository{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return user, nil
			},
		}
		svc := newTestService(repo, userRepo, &mockEmailSender{})
		result, err := svc.Reject(ctx, 1, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "rejected" {
			t.Fatalf("expected status rejected, got %s", result.Status)
		}
	})

	t.Run("success without user lookup", func(t *testing.T) {
		app := testApplication()
		repo := &mockSellerApplicationRepository{
			getByID: func(ctx context.Context, id int64) (db.SellerApplication, error) {
				return app, nil
			},
			updateStatus: func(ctx context.Context, id int64, status string) (db.SellerApplication, error) {
				app.Status = "rejected"
				return app, nil
			},
		}
		userRepo := &mockUserRepository{
			getUserByID: func(ctx context.Context, id int64) (db.User, error) {
				return db.User{}, errors.New("user not found")
			},
		}
		svc := newTestService(repo, userRepo, &mockEmailSender{})
		result, err := svc.Reject(ctx, 1, "Недостаточно информации")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != "rejected" {
			t.Fatalf("expected status rejected, got %s", result.Status)
		}
	})
}
