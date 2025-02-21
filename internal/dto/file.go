package dto

import (
	"mime/multipart"
	"time"
)

// FileInfo 文件信息
type FileInfo struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
	IsVirtual bool      `json:"is_virtual"` // 是否是虚拟目录（断层目录）
}

// FileUploadChunk 文件分块上传
type FileUploadChunk struct {
	Path  string                `form:"path" validate:"required"`
	Order int                   `form:"order" validate:"required,min=0"`
	Hash  string                `form:"hash" validate:"required,len=64"`
	File  *multipart.FileHeader `form:"file" validate:"required"`
}

// CreateDirectoryRequest 创建目录请求
type CreateDirectoryRequest struct {
	Path string `json:"path" validate:"required"`
}

// CreateFileRequest 创建文件记录
type CreateFileRequest struct {
	Path     string `json:"path" validate:"required"`
	MimeType string `json:"mimetype"`
	Size     int64  `json:"size"`
}

// FileListRequest 文件列表请求
type FileListRequest struct {
	Path string `json:"path" validate:"required"`
}

// ChunkInfo 分块信息
type ChunkInfo struct {
	Order int    `json:"order"`
	Hash  string `json:"hash"`
	Size  int64  `json:"size"`
}

// InitUploadRequest 初始化文件上传请求
type InitUploadRequest struct {
	SavePath string `json:"save_path" validate:"required"`
	Size     int64  `json:"size" validate:"required,gt=0"`
	Override bool   `json:"override"`
}

// InitUploadResponse 初始化文件上传响应
type InitUploadResponse struct {
	File       *FileInfo `json:"file"`
	ChunkSize  int       `json:"chunk_size"`
	ChunkCount int       `json:"chunk_count"`
}

// MoveFileRequest 移动文件请求
type MoveFileRequest struct {
	SourcePath string `json:"source_path" validate:"required"`
	TargetPath string `json:"target_path" validate:"required"`
	Override   bool   `json:"override"`
}

// CopyFileRequest 复制文件请求
type CopyFileRequest struct {
	SourcePath string `json:"source_path" validate:"required"`
	TargetPath string `json:"target_path" validate:"required"`
	Override   bool   `json:"override"`
}

// RenameFileRequest 重命名文件请求
type RenameFileRequest struct {
	Path     string `json:"path" validate:"required"`
	NewName  string `json:"new_name" validate:"required"`
	Override bool   `json:"override"`
}

// FileDownloadRequest 下载文件请求
type FileDownloadRequest struct {
	Path  string `json:"path" validate:"required" form:"path"`
	Order int    `json:"order" validate:"required,min=0" form:"order"`
}

// FileDownloadResponse 下载文件响应
type FileDownloadResponse struct {
	ChunkSize uint64 `json:"chunk_size"`
	ChunkHash string `json:"chunk_hash"`
	Url       string `json:"url"`
}

// FileChunk 文件分块
type FileChunk struct {
	ID        uint64    `json:"id"`
	Order     int       `json:"order"`
	Hash      string    `json:"hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FileDownloadChunkRequest 下载文件分块请求
type FileDownloadChunkRequest struct {
	Path  string `json:"path" validate:"required" form:"path"`
	Order int    `json:"order" validate:"required,min=0" form:"order"`
}

type FileDownloadLink struct {
	Order int    `json:"order"`
	Size  uint64 `json:"size"`
	Hash  string `json:"hash"`
	Url   string `json:"url"`
}

// UploadInfo 上传信息
type UploadInfo struct {
	WorkspaceId    uint64 `json:"workspace_id"`
	Path           string `json:"path"`
	Name           string `json:"name"`
	Size           int64  `json:"size"`
	ChunkSize      int64  `json:"chunk_size"`
	LastChunkSize  int64  `json:"last_chunk_size"`
	ChunkCount     int    `json:"chunk_count"`
	UploadedChunks []int  `json:"uploaded_chunks"`
	CreatedAt      int64  `json:"created_at"`
	Override       bool   `json:"override"`
}

// RecycleBinItem 回收站项目信息
type RecycleBinItem struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	OriginalPath string    `json:"original_path"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	DeletedAt    time.Time `json:"deleted_at"`
	ExpireAt     time.Time `json:"expire_at"`
}

// ListRecycleBinRequest 列出回收站请求
type ListRecycleBinRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

// RestoreFromRecycleBinRequest 从回收站恢复文件请求
type RestoreFromRecycleBinRequest struct {
	Path       string `json:"path"`
	Override   bool   `json:"override"`
	TargetPath string `json:"target_path"`
}

// RecycleBinSortField 回收站排序字段
type RecycleBinSortField string

const (
	// RecycleBinSortFieldDeletedAt 按删除时间排序
	RecycleBinSortFieldDeletedAt RecycleBinSortField = "deleted_at"
	// RecycleBinSortFieldExpireAt 按过期时间排序
	RecycleBinSortFieldExpireAt RecycleBinSortField = "expire_at"
	// RecycleBinSortFieldName 按名称排序
	RecycleBinSortFieldName RecycleBinSortField = "name"
	// RecycleBinSortFieldSize 按大小排序
	RecycleBinSortFieldSize RecycleBinSortField = "size"
	// RecycleBinSortFieldOriginalPath 按原始路径排序
	RecycleBinSortFieldOriginalPath RecycleBinSortField = "original_path"
)

// GetRecycleBinSortFieldMap 获取回收站排序字段映射
func GetRecycleBinSortFieldMap() map[RecycleBinSortField]string {
	return map[RecycleBinSortField]string{
		RecycleBinSortFieldDeletedAt:    "deleted_at",
		RecycleBinSortFieldExpireAt:     "expire_at",
		RecycleBinSortFieldName:         "name",
		RecycleBinSortFieldSize:         "size",
		RecycleBinSortFieldOriginalPath: "original_path",
	}
}

// IsValidRecycleBinSortField 检查是否为有效的排序字段
func IsValidRecycleBinSortField(field string) bool {
	_, ok := GetRecycleBinSortFieldMap()[RecycleBinSortField(field)]
	return ok
}

// GetDefaultRecycleBinSortField 获取默认的排序字段
func GetDefaultRecycleBinSortField() RecycleBinSortField {
	return RecycleBinSortFieldDeletedAt
}

// RestoreFileRequest 从回收站恢复文件的请求
type RestoreFileRequest struct {
	Path string `json:"path" validate:"required"`
}

// AddChunkByHashRequest 通过 hash 添加分块的请求
type AddChunkByHashRequest struct {
	Path  string `json:"path" validate:"required"`
	Order int    `json:"order" validate:"required,min=0"`
	Size  int64  `json:"size" validate:"required,min=1"`
	Hash  string `json:"hash" validate:"required,len=64"`
}

// OrderedHash 排序后的哈希值列表
type OrderedHash struct {
	Order uint64
	Hash  string
}
