// rpg-market-backend/internal/usecase/user_usecase_test.go
package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/shinybell/rpg-market-backend/internal/domain/entity"
)

// MockUserRepository はテスト用のモックリポジトリ
type MockUserRepository struct {
	users              map[string]*entity.User // firebaseUID -> User
	createError        error
	findError          error
	updateError        error
	deleteError        error
	lastLoginError     error
	updateProfileError error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*entity.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	if m.createError != nil {
		return m.createError
	}
	user.ID = int64(len(m.users) + 1)
	m.users[user.FirebaseUID] = user
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*entity.User, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) FindByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	user, exists := m.users[firebaseUID]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	if m.updateError != nil {
		return m.updateError
	}
	m.users[user.FirebaseUID] = user
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	if m.lastLoginError != nil {
		return m.lastLoginError
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, userID int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	for uid, user := range m.users {
		if user.ID == userID {
			delete(m.users, uid)
			return nil
		}
	}
	return nil
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, profile *entity.UserProfile) error {
	if m.updateProfileError != nil {
		return m.updateProfileError
	}
	return nil
}

func (m *MockUserRepository) UpdateWallet(ctx context.Context, wallet *entity.Wallet) error {
	return nil
}

// TestLoginOrRegister_NewUser は新規ユーザー登録のテスト
func TestLoginOrRegister_NewUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	user, err := uc.LoginOrRegister(ctx, "firebase123", "test@example.com", "testuser")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user to be created")
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.Profile == nil || user.Profile.Nickname != "testuser" {
		t.Errorf("expected nickname 'testuser', got '%s'", user.Profile.Nickname)
	}
	if user.Status != entity.UserStatusActive {
		t.Errorf("expected status 'active', got '%s'", user.Status)
	}
}

// TestLoginOrRegister_ExistingUser は既存ユーザーログインのテスト
func TestLoginOrRegister_ExistingUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	// 既存ユーザーを作成
	existingUser := &entity.User{
		ID:          1,
		FirebaseUID: "firebase123",
		Email:       "existing@example.com",
		Status:      entity.UserStatusActive,
		Profile:     &entity.UserProfile{Nickname: "existing"},
	}
	mockRepo.users["firebase123"] = existingUser

	user, err := uc.LoginOrRegister(ctx, "firebase123", "existing@example.com", "newname")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected existing user with ID 1, got %d", user.ID)
	}
	// 既存ユーザーなのでニックネームは変わらない
	if user.Profile.Nickname != "existing" {
		t.Errorf("expected nickname 'existing', got '%s'", user.Profile.Nickname)
	}
}

// TestLoginOrRegister_InvalidNickname はニックネームバリデーションのテスト
func TestLoginOrRegister_InvalidNickname(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name     string
		nickname string
	}{
		{"too short", "a"},
		{"too long", "this_is_a_very_long_nickname_that_exceeds_fifty_characters_limit"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.LoginOrRegister(ctx, "firebase123", "test@example.com", tt.nickname)
			if err != ErrInvalidNickname {
				t.Errorf("expected ErrInvalidNickname, got %v", err)
			}
		})
	}
}

// TestGetUserByFirebaseUID はFirebase UIDでユーザー取得のテスト
func TestGetUserByFirebaseUID(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	// ユーザーを作成
	existingUser := &entity.User{
		ID:          1,
		FirebaseUID: "firebase123",
		Email:       "test@example.com",
		Status:      entity.UserStatusActive,
	}
	mockRepo.users["firebase123"] = existingUser

	t.Run("existing user", func(t *testing.T) {
		user, err := uc.GetUserByFirebaseUID(ctx, "firebase123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.ID != 1 {
			t.Errorf("expected user ID 1, got %d", user.ID)
		}
	})

	t.Run("non-existing user", func(t *testing.T) {
		_, err := uc.GetUserByFirebaseUID(ctx, "nonexistent")
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// TestUpdateProfile はプロフィール更新のテスト
func TestUpdateProfile(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	// ユーザーを作成
	existingUser := &entity.User{
		ID:          1,
		FirebaseUID: "firebase123",
		Email:       "test@example.com",
		Status:      entity.UserStatusActive,
		Profile:     &entity.UserProfile{UserID: 1, Nickname: "oldname"},
	}
	mockRepo.users["firebase123"] = existingUser

	newNickname := "newname"
	newBio := "New bio"

	user, err := uc.UpdateProfile(ctx, "firebase123", &newNickname, &newBio, nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Profile.Nickname != "newname" {
		t.Errorf("expected nickname 'newname', got '%s'", user.Profile.Nickname)
	}
	if user.Profile.Bio == nil || *user.Profile.Bio != "New bio" {
		t.Errorf("expected bio 'New bio', got '%v'", user.Profile.Bio)
	}
}

// TestUpdateProfile_InvalidNickname はプロフィール更新時のバリデーションテスト
func TestUpdateProfile_InvalidNickname(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	existingUser := &entity.User{
		ID:          1,
		FirebaseUID: "firebase123",
		Email:       "test@example.com",
		Profile:     &entity.UserProfile{UserID: 1, Nickname: "oldname"},
	}
	mockRepo.users["firebase123"] = existingUser

	invalidNickname := "a"
	_, err := uc.UpdateProfile(ctx, "firebase123", &invalidNickname, nil, nil)

	if err != ErrInvalidNickname {
		t.Errorf("expected ErrInvalidNickname, got %v", err)
	}
}

// TestDeleteUser はユーザー削除のテスト
func TestDeleteUser(t *testing.T) {
	mockRepo := NewMockUserRepository()
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	existingUser := &entity.User{
		ID:          1,
		FirebaseUID: "firebase123",
		Email:       "test@example.com",
	}
	mockRepo.users["firebase123"] = existingUser

	t.Run("delete existing user", func(t *testing.T) {
		err := uc.DeleteUser(ctx, "firebase123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, exists := mockRepo.users["firebase123"]; exists {
			t.Error("expected user to be deleted")
		}
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		err := uc.DeleteUser(ctx, "nonexistent")
		if err != ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// TestRepositoryError はリポジトリエラーのハンドリングテスト
func TestRepositoryError(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockRepo.findError = errors.New("database error")
	uc := NewUserUseCase(mockRepo)
	ctx := context.Background()

	_, err := uc.GetUserByFirebaseUID(ctx, "firebase123")
	if err == nil {
		t.Error("expected error from repository")
	}
}
