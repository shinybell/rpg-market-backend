package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrGenerationFailed  = errors.New("generation failed")
	ErrContentFiltered   = errors.New("content filtered due to safety concerns")
	ErrExceededMaxLength = errors.New("input exceeds maximum length")
)

const (
	MaxInputLength       = 300 // 入力の最大文字数
	MaxSuggestions       = 5   // 最大提案数
	DefaultSuggestions   = 3   // デフォルト提案数
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
	ItemName       string
	Category       string
	Condition      string
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

	promptParts = append(promptParts, "あなたは商品説明文を生成するアシスタントです。")
	promptParts = append(promptParts, fmt.Sprintf("以下の情報を元に、魅力的な商品説明文を%d個生成してください。", req.NumSuggestions))
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, fmt.Sprintf("【商品名】%s", req.ItemName))

	if req.Condition != "" {
		promptParts = append(promptParts, fmt.Sprintf("【状態】%s", req.Condition))
	}

	promptParts = append(promptParts, "")
	promptParts = append(promptParts, "【条件】")
	promptParts = append(promptParts, fmt.Sprintf("- 各説明文は全体で%d文字以上%d文字以内（日本語文字数）", MinDescriptionLength, MaxDescriptionLength))
	promptParts = append(promptParts, "- 購入者の興味を引く魅力的で詳細な内容")
	promptParts = append(promptParts, "- 商品の特徴や魅力を具体的に伝える")
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
			// スキップするのではなく、最大の長さでトリムする
			if len(cleaned) > MaxDescriptionLength {
				cleaned = cleaned[:MaxDescriptionLength]
			} else {
				continue
			}

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

// AppraiseItemRequest はアイテム鑑定のリクエスト
type AppraiseItemRequest struct {
	ItemName    string
	Description string
	Category    string
	Condition   string
}

// AppraiseItemResponse はアイテム鑑定のレスポンス
type AppraiseItemResponse struct {
	RPGName        string
	RPGDescription string
}

// AppraiseItem はアイテムをRPG風に鑑定する
func (uc *GenerationUseCase) AppraiseItem(ctx context.Context, req AppraiseItemRequest) (*AppraiseItemResponse, error) {
	// バリデーション
	if strings.TrimSpace(req.ItemName) == "" {
		return nil, fmt.Errorf("%w: item name is required", ErrInvalidInput)
	}

	// RPG風の商品名を生成
	rpgName, err := uc.generateRPGName(ctx, req)
	if err != nil {
		return nil, err
	}

	// RPG風の説明文を生成
	rpgDescription, err := uc.generateRPGDescription(ctx, req, rpgName)
	if err != nil {
		return nil, err
	}

	return &AppraiseItemResponse{
		RPGName:        rpgName,
		RPGDescription: rpgDescription,
	}, nil
}

// generateRPGName はRPG風の商品名を生成
func (uc *GenerationUseCase) generateRPGName(ctx context.Context, req AppraiseItemRequest) (string, error) {
	var promptParts []string

	promptParts = append(promptParts, "あなたはRPGゲームのアイテム鑑定士です。")
	promptParts = append(promptParts, "以下の現代的なアイテム情報を、RPGゲームの世界観に合った魅力的なアイテム名に変換してください。")
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, fmt.Sprintf("【元の商品名】%s", req.ItemName))

	if req.Category != "" {
		promptParts = append(promptParts, fmt.Sprintf("【カテゴリ】%s", req.Category))
	}

	if req.Condition != "" {
		promptParts = append(promptParts, fmt.Sprintf("【状態】%s", req.Condition))
	}

	promptParts = append(promptParts, "")
	promptParts = append(promptParts, "【条件】")
	promptParts = append(promptParts, "- RPG風のファンタジー要素を取り入れた名前にする")
	promptParts = append(promptParts, "- 元の商品の特徴を残しつつ、魔法や伝説的な要素を加える")
	promptParts = append(promptParts, "- 20文字以内で簡潔に")
	promptParts = append(promptParts, "- アイテム名のみを出力（説明や番号は不要）")

	prompt := strings.Join(promptParts, "\n")

	generatedText, err := uc.generationService.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// 生成されたテキストをクリーンアップ
	rpgName := strings.TrimSpace(generatedText)
	rpgName = strings.Split(rpgName, "\n")[0] // 最初の行のみ取得

	log.Printf("[APPRAISE] Generated RPG name: %s", rpgName)

	return rpgName, nil
}

// generateRPGDescription はRPG風の説明文を生成
func (uc *GenerationUseCase) generateRPGDescription(ctx context.Context, req AppraiseItemRequest, rpgName string) (string, error) {
	var promptParts []string

	promptParts = append(promptParts, "あなたはRPGゲームのアイテム鑑定士です。")
	promptParts = append(promptParts, "以下のアイテムについて、RPGゲームの世界観に合った魅力的な説明文を生成してください。")
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, fmt.Sprintf("【RPG風アイテム名】%s", rpgName))
	promptParts = append(promptParts, fmt.Sprintf("【元の商品名】%s", req.ItemName))

	if req.Description != "" {
		promptParts = append(promptParts, fmt.Sprintf("【元の説明】%s", req.Description))
	}

	if req.Category != "" {
		promptParts = append(promptParts, fmt.Sprintf("【カテゴリ】%s", req.Category))
	}

	promptParts = append(promptParts, "")
	promptParts = append(promptParts, "【条件】")
	promptParts = append(promptParts, "- RPGゲームの世界観に合った表現を使用")
	promptParts = append(promptParts, "- 伝説や神話的な要素を含める")
	promptParts = append(promptParts, "- 冒険者の興味を引く魅力的な内容")
	promptParts = append(promptParts, fmt.Sprintf("- %d文字以上%d文字以内", MinDescriptionLength, MaxDescriptionLength))
	promptParts = append(promptParts, "- 説明文のみを出力（番号や余計な文章は不要）")

	prompt := strings.Join(promptParts, "\n")

	generatedText, err := uc.generationService.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// 生成されたテキストをクリーンアップ
	rpgDescription := strings.TrimSpace(generatedText)

	log.Printf("[APPRAISE] Generated RPG description: %s", rpgDescription)

	return rpgDescription, nil
}

// ConvertSearchQueryRequest は検索クエリ変換のリクエスト
type ConvertSearchQueryRequest struct {
	UserQuery string
}

// ConvertSearchQueryResponse は検索クエリ変換のレスポンス
type ConvertSearchQueryResponse struct {
	Keywords []string
}

// ConvertSearchQuery はユーザーの自然言語クエリを検索キーワードに変換する
func (uc *GenerationUseCase) ConvertSearchQuery(ctx context.Context, req ConvertSearchQueryRequest) (*ConvertSearchQueryResponse, error) {
	// バリデーション
	if strings.TrimSpace(req.UserQuery) == "" {
		return nil, fmt.Errorf("%w: user query is required", ErrInvalidInput)
	}

	// クエリの長さチェック
	if len(req.UserQuery) > MaxInputLength {
		return nil, fmt.Errorf("%w: query too long", ErrExceededMaxLength)
	}

	// プロンプトを構築
	var promptParts []string
	promptParts = append(promptParts, "あなたは検索クエリ抽出の専門家です。")
	promptParts = append(promptParts, "ユーザーが入力した文章から、商品検索に最適なキーワードを抽出してください。")
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, fmt.Sprintf("【ユーザーの入力】%s", req.UserQuery))
	promptParts = append(promptParts, "")
	promptParts = append(promptParts, "【条件】")
	promptParts = append(promptParts, "- 重要なキーワードを1〜5個抽出")
	promptParts = append(promptParts, "- 各キーワードは半角スペース区切りで1行に出力")
	promptParts = append(promptParts, "- 余計な説明や記号は不要")
	promptParts = append(promptParts, "- 例: 「剣 強い 初心者」")

	prompt := strings.Join(promptParts, "\n")

	// Gemini APIを呼び出し
	generatedText, err := uc.generationService.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
	}

	// 生成されたテキストをクリーンアップしてキーワードに分割
	cleanText := strings.TrimSpace(generatedText)
	cleanText = strings.ReplaceAll(cleanText, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "　", " ") // 全角スペースを半角に

	keywords := []string{}
	for _, word := range strings.Fields(cleanText) {
		word = strings.TrimSpace(word)
		if word != "" && len(word) < 50 { // 極端に長い単語は除外
			keywords = append(keywords, word)
		}
		if len(keywords) >= 5 { // 最大5個まで
			break
		}
	}

	log.Printf("[SEARCH QUERY] User query: %s -> Keywords: %v", req.UserQuery, keywords)

	if len(keywords) == 0 {
		return nil, fmt.Errorf("%w: no valid keywords extracted", ErrGenerationFailed)
	}

	return &ConvertSearchQueryResponse{
		Keywords: keywords,
	}, nil
}
