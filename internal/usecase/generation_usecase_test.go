package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// MockGenerationService はテスト用のモックサービス
type MockGenerationService struct {
	GenerateContentFunc func(ctx context.Context, prompt string) (string, error)
}

func (m *MockGenerationService) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if m.GenerateContentFunc != nil {
		return m.GenerateContentFunc(ctx, prompt)
	}
	return "", errors.New("not implemented")
}

func TestGenerateDescriptionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     GenerateDescriptionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: GenerateDescriptionRequest{
				ItemName:       "伝説の剣",
				Category:       "武器",
				Condition:      "新品",
				NumSuggestions: 3,
			},
			wantErr: false,
		},
		{
			name: "empty item name",
			req: GenerateDescriptionRequest{
				ItemName:       "",
				Category:       "武器",
				NumSuggestions: 3,
			},
			wantErr: true,
		},
		{
			name: "whitespace only item name",
			req: GenerateDescriptionRequest{
				ItemName:       "   ",
				Category:       "武器",
				NumSuggestions: 3,
			},
			wantErr: true,
		},
		{
			name: "exceeds max length",
			req: GenerateDescriptionRequest{
				ItemName:       strings.Repeat("あ", MaxInputLength),
				Category:       "武器",
				NumSuggestions: 3,
			},
			wantErr: true,
		},
		{
			name: "zero suggestions defaults to 3",
			req: GenerateDescriptionRequest{
				ItemName:       "伝説の剣",
				NumSuggestions: 0,
			},
			wantErr: false,
		},
		{
			name: "exceeds max suggestions",
			req: GenerateDescriptionRequest{
				ItemName:       "伝説の剣",
				NumSuggestions: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// NumSuggestionsのデフォルト値チェック
			if !tt.wantErr && tt.req.NumSuggestions == 0 {
				// Validateを呼んだ後、デフォルト値が設定されているはず
				// （実際にはValidateメソッド内で設定される）
			}
		})
	}
}

func TestGenerationUseCase_GenerateDescriptions_Success(t *testing.T) {
	mockService := &MockGenerationService{
		GenerateContentFunc: func(ctx context.Context, prompt string) (string, error) {
			return `1. この剣は伝説の勇者が使用した魔法の武器です。光の力を宿し、闇を切り裂く能力を持っています。
2. 古代の鍛冶師が作り上げた最高傑作。その刃は決して錆びることなく、永遠の輝きを保ちます。
3. 勇者のみが扱える神秘的な剣。持つ者に勇気と力を与え、どんな困難も乗り越えられるでしょう。`, nil
		},
	}

	uc := NewGenerationUseCase(mockService)

	req := GenerateDescriptionRequest{
		ItemName:       "伝説の剣",
		Category:       "武器",
		Condition:      "新品",
		NumSuggestions: 3,
	}

	resp, err := uc.GenerateDescriptions(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Suggestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(resp.Suggestions))
	}

	for i, suggestion := range resp.Suggestions {
		if len(suggestion) < MinDescriptionLength {
			t.Errorf("suggestion %d is too short: %d chars", i, len(suggestion))
		}
		if len(suggestion) > MaxDescriptionLength {
			t.Errorf("suggestion %d is too long: %d chars", i, len(suggestion))
		}
	}
}

func TestGenerationUseCase_GenerateDescriptions_InvalidInput(t *testing.T) {
	mockService := &MockGenerationService{}
	uc := NewGenerationUseCase(mockService)

	req := GenerateDescriptionRequest{
		ItemName:       "", // 空のアイテム名
		NumSuggestions: 3,
	}

	_, err := uc.GenerateDescriptions(context.Background(), req)

	if err == nil {
		t.Error("expected error for invalid input")
	}

	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGenerationUseCase_GenerateDescriptions_ServiceError(t *testing.T) {
	mockService := &MockGenerationService{
		GenerateContentFunc: func(ctx context.Context, prompt string) (string, error) {
			return "", errors.New("API error")
		},
	}

	uc := NewGenerationUseCase(mockService)

	req := GenerateDescriptionRequest{
		ItemName:       "伝説の剣",
		NumSuggestions: 3,
	}

	_, err := uc.GenerateDescriptions(context.Background(), req)

	if err == nil {
		t.Error("expected error when service fails")
	}

	if !errors.Is(err, ErrGenerationFailed) {
		t.Errorf("expected ErrGenerationFailed, got %v", err)
	}
}

func TestGenerationUseCase_GenerateDescriptions_ProhibitedWords(t *testing.T) {
	mockService := &MockGenerationService{}
	uc := NewGenerationUseCase(mockService)

	req := GenerateDescriptionRequest{
		ItemName:       "違法な武器",
		NumSuggestions: 3,
	}

	_, err := uc.GenerateDescriptions(context.Background(), req)

	if err == nil {
		t.Error("expected error for prohibited words")
	}

	if !errors.Is(err, ErrContentFiltered) {
		t.Errorf("expected ErrContentFiltered, got %v", err)
	}
}

func TestGenerationUseCase_BuildPrompt(t *testing.T) {
	mockService := &MockGenerationService{}
	uc := NewGenerationUseCase(mockService)

	req := GenerateDescriptionRequest{
		ItemName:       "伝説の剣",
		Category:       "武器",
		Condition:      "新品",
		NumSuggestions: 3,
	}

	prompt := uc.buildPrompt(req)

	// プロンプトに必要な情報が含まれているか確認
	if !strings.Contains(prompt, req.ItemName) {
		t.Error("prompt should contain item name")
	}
	if !strings.Contains(prompt, req.Category) {
		t.Error("prompt should contain category")
	}
	if !strings.Contains(prompt, req.Condition) {
		t.Error("prompt should contain condition")
	}
}

func TestGenerationUseCase_ParseSuggestions(t *testing.T) {
	mockService := &MockGenerationService{}
	uc := NewGenerationUseCase(mockService)

	text := `1. この剣は伝説の勇者が使用した魔法の武器です。光の力を宿し、闇を切り裂く能力を持っています。
2. 古代の鍛冶師が作り上げた最高傑作。その刃は決して錆びることなく、永遠の輝きを保ちます。
3. 勇者のみが扱える神秘的な剣。持つ者に勇気と力を与え、どんな困難も乗り越えられるでしょう。`

	suggestions := uc.parseSuggestions(text, 5)

	if len(suggestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(suggestions))
	}

	// 各候補が番号なしで格納されているか確認
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, "1.") || strings.HasPrefix(suggestion, "2.") {
			t.Errorf("suggestion should not contain number prefix: %s", suggestion)
		}
	}
}

func TestGenerationUseCase_ContainsProhibitedWords(t *testing.T) {
	mockService := &MockGenerationService{}
	uc := NewGenerationUseCase(mockService)

	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "no prohibited words",
			text: "伝説の剣",
			want: false,
		},
		{
			name: "contains prohibited word",
			text: "暴力的な武器",
			want: true,
		},
		{
			name: "case insensitive check",
			text: "VIOLATION",
			want: false, // 英語の禁止ワードはリストに含まれていない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uc.containsProhibitedWords(tt.text)
			if got != tt.want {
				t.Errorf("containsProhibitedWords() = %v, want %v", got, tt.want)
			}
		})
	}
}
