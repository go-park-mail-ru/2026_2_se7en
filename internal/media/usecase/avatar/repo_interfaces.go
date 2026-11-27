package avatar

import (
	"context"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/media/model"
)

type AvatarRepository interface {
	Save(ctx context.Context, file *models.FileUpload) (url string, filename string, err error)
	Remove(ctx context.Context, filename string) error
}
