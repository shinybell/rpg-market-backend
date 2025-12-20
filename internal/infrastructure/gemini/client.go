package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client はGemini APIクライアント
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient は新しいGemini APIクライアントを作成
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateContentRequest はコンテンツ生成リクエスト
type GenerateContentRequest struct {
	Contents []Content `json:"contents"`
}

// Content はプロンプトの内容
type Content struct {
	Parts []Part `json:"parts"`
}

// Part はプロンプトの各パート
type Part struct {
	Text string `json:"text"`
}

// GenerateContentResponse はGemini APIからのレスポンス
type GenerateContentResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate は生成された候補
type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

// SafetyRating は安全性評価
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GenerateContent はGemini APIを使ってコンテンツを生成
func (c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY is not set")
	}

	// リクエストボディを作成
	reqBody := GenerateContentRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// APIエンドポイント
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey)

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエスト実行
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// ステータスコードチェック
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	// レスポンスをパース
	var genResp GenerateContentResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 候補が存在するかチェック
	if len(genResp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned from Gemini API")
	}

	// 最初の候補のテキストを取得
	candidate := genResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content parts in candidate")
	}

	return candidate.Content.Parts[0].Text, nil
}
