package uploader

import (
	"io"
	"time"
)

// ArtifactSource represents a source for artifact upload
type ArtifactSource interface {
	Name() string      // アーティファクト名
	Reader() io.Reader // データソース
	Size() int64       // サイズ（分かる場合、-1で不明）
	Close() error      // リソースクリーンアップ
}

// UploadOptions contains options for artifact upload
type UploadOptions struct {
	Overwrite  bool          // 既存ファイルを上書きするか
	Timeout    time.Duration // アップロードタイムアウト
	RetryCount int           // リトライ回数
	ChunkSize  int64         // チャンクサイズ
}

// CredentialsForWriteRequest represents the request for credentials-for-write API
type CredentialsForWriteRequest struct {
	RunID string   `json:"run_id"`
	Path  []string `json:"path"`
}

// CredentialsForWriteResponse represents the response from credentials-for-write API
type CredentialsForWriteResponse struct {
	CredentialInfos []ArtifactCredentialInfo `json:"credential_infos"`
}

// ArtifactCredentialInfo represents artifact credential information
type ArtifactCredentialInfo struct {
	RunID     string       `json:"run_id"`
	Path      string       `json:"path"`
	SignedURI string       `json:"signed_uri"`
	Headers   []HTTPHeader `json:"headers"`
	Type      string       `json:"type"`
}

// HTTPHeader represents HTTP header
type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
