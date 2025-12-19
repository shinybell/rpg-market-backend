package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrGenerationFailed    = errors.New("generation failed")
	ErrContentFiltered     = errors.New("content filtered due to safety concerns")
	ErrExceededMaxLength   = errors.New("input exceeds maximum length")
)

const (
	MaxInputLength      = 500  // 入力の最大文字数
	MaxSuggestions      = 5    // 最大提案数
	DefaultSuggestions  = 3    // デフォルト提案数
	MinDescriptionLength = 50  // 生成される説明文の最小文字数
	MaxDescriptionLength = 500 // 生成される説明文の最大文字数（日本語対応）
)

// GenerationService はコンテンツ生成サービスのインターフェース
type GenerationService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// GenerationUseCase はアイテム説明文生成のユースケース
type GenerationUseCase struct {
	generationService GenerationService
}

// NewGenerationUseCase は新しいGenerationUseCaseを作成
func NewGenerationUseCase(generationService GenerationService) *GenerationUseCase {
	return &GenerationUseCase{
		generationService: generationService,
	}
}

// GenerateDescriptionRequest は説明文生成のリクエスト
type GenerateDescriptionRequest struct {
	ItemName      string
	Category      string
	Condition     string
	NumSuggestions int
}

// GenerateDescriptionResponse は説明文生成のレスポンス
type GenerateDescriptionResponse struct {
	Suggestions []string
}

// Validate はリクエストのバリデーション
func (r *GenerateDescriptionRequest) Validate() error {
	// アイテム名は必須
	if strings.TrimSpace(r.ItemName) == "" {
		return fmt.Errorf("%w: item name is required", ErrInvalidInput)
	}

	// 入力長のチェック
	totalLength := len(r.ItemName) + len(r.Category) + len(r.Condition)
	if totalLength > MaxInputLength {
		return fmt.Errorf("%w: total input length %d exceeds maximum %d", ErrExceededMaxLength, totalLength, MaxInputLength)
	}

	// 提案数のバリデーション
	if r.NumSuggestions <= 0 {
		r.NumSuggestions = DefaultSuggestions
	}
	if r.NumSuggestions > MaxSuggestions {
		r.NumSuggestions = MaxSuggestions
	}

	return nil
}

// GenerateDescriptions はアイテムの説明文候補を生成する
func (uc *GenerationUseCase) GenerateDescriptions(ctx context.Context, req GenerateDescriptionRequest) (*GenerateDescriptionResponse, error) {
	// バリデーション
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// プロンプトを構築
	prompt := uc.buildPrompt(req)

	// 簡易コンテンツフィルタ（禁止ワードチェック）
	if uc.containsProhibitedWords(req.ItemName) {
		return nil, fmt.Errorf("%w: prohibited words detected", ErrContentFiltered)
	}

	// Gemini APIを呼び出し
	generatedText, err := uc.generationService.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// デバッグ用ログ
	log.Printf("[GENERATION] Raw response from Gemini: %s", generatedText)

	// 生成されたテキストを解析して候補に分割
	suggestions := uc.parseSuggestions(generatedText, req.NumSuggestions)

	// デバッグ用ログ
	log.Printf("[GENERATION] Parsed suggestions count: %d", len(suggestions))

	// 候補が空の場合はエラー
	if len(suggestions) == 0 {
		return nil, fmt.Errorf("%w: no valid suggestions generated", ErrGenerationFailed)
	}

	return &GenerateDescriptionResponse{
		Suggestions: suggestions,
	}, nil
}

// buildPrompt はプロンプトを構築
func (uc *GenerationUseCase) buildPrompt(req GenerateDescriptionRequest) string {
	var promptParts []string

	promptParts = append(promptParts, "あなたはRPGゲームのアイテム説明文を生成するアシスタントです。")
	promptParts = append(promptParts, fmt.Sprintf("以下の情報を元に、魅力的なアイテム説明文を%d個生成してください。", req.NumSuggestions))
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, fmt.Sprintf("【アイテム名】%s", req.ItemName))

	if req.Category != "" {
		promptParts = append(promptParts, fmt.Sprintf("【カテゴリ】%s", req.Category))
	}

	if req.Condition != "" {
		promptParts = append(promptParts, fmt.Sprintf("【状態】%s", req.Condition))
	}

	promptParts = append(promptParts, "")
	promptParts = append(promptParts, "【条件】")
	promptParts = append(promptParts, fmt.Sprintf("- 各説明文は%d文字以上%d文字以内（日本語文字数）", MinDescriptionLength, MaxDescriptionLength))
	promptParts = append(promptParts, "- RPGゲームの世界観に合った表現を使用")
	promptParts = append(promptParts, "- 購入者の興味を引く魅力的な内容")
	promptParts = append(promptParts, "- 各説明文は番号付きリストで出力（例: 1. 説明文その1）")
	promptParts = append(promptParts, "- 説明文のみを出力し、他の余計な文章は含めない")

	return strings.Join(promptParts, "\n")
}

// parseSuggestions は生成されたテキストから候補を抽出
func (uc *GenerationUseCase) parseSuggestions(text string, maxSuggestions int) []string {
	var suggestions []string

	// 改行で分割
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 番号付きリストの形式をパース（例: "1. 説明文", "1) 説明文", "1 説明文"）
		// または番号なしの行もそのまま受け入れる
		cleaned := line

		// "1. " や "1) " などのプレフィックスを削除
		for _, prefix := range []string{". ", ") ", " "} {
			for i := 1; i <= 10; i++ {
				pattern := fmt.Sprintf("%d%s", i, prefix)
				if strings.HasPrefix(cleaned, pattern) {
					cleaned = strings.TrimPrefix(cleaned, pattern)
					break
				}
			}
		}

		cleaned = strings.TrimSpace(cleaned)

		// デバッグ用ログ
		log.Printf("[PARSE] Line: %s, Cleaned: %s, Length: %d", line, cleaned, len(cleaned))

		// 長さチェック
		if len(cleaned) < MinDescriptionLength || len(cleaned) > MaxDescriptionLength {
			log.Printf("[PARSE] Skipped due to length: %d (min: %d, max: %d)", len(cleaned), MinDescriptionLength, MaxDescriptionLength)
			continue
		}

		// 簡易フィルタリング
		if uc.containsProhibitedWords(cleaned) {
			continue
		}

		suggestions = append(suggestions, cleaned)

		// 最大数に達したら終了
		if len(suggestions) >= maxSuggestions {
			break
		}
	}

	return suggestions
}

// containsProhibitedWords は禁止ワードが含まれているかチェック
func (uc *GenerationUseCase) containsProhibitedWords(text string) bool {
	// 簡易的な禁止ワードリスト
	prohibitedWords := []string{
		"暴力", "殺人", "麻薬", "違法",
		// 必要に応じて追加
	}

	lowerText := strings.ToLower(text)
	for _, word := range prohibitedWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			return true
		}
	}

	return false
}
