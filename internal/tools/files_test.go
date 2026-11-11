package tools

import (
	"bytes"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_2_se7en/internal/models"
)

func TestSaveAndRemoveAvatar(t *testing.T) {
	t.Chdir(t.TempDir())
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, _, err := SaveAvatar(t.Context(), nil, log); err == nil {
		t.Fatal("nil upload accepted")
	}
	if _, _, err := SaveAvatar(t.Context(), &models.FileUpload{Reader: strings.NewReader("x"), Size: 1, ContentType: "text/plain"}, log); err == nil {
		t.Fatal("invalid content type accepted")
	}
	url, filename, err := SaveAvatar(t.Context(), &models.FileUpload{Reader: strings.NewReader("png"), Filename: "avatar.png", Size: 3, ContentType: "image/png"}, log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "/uploads/avatars/") || !strings.HasSuffix(filename, ".png") {
		t.Fatalf("unexpected avatar location: %q %q", url, filename)
	}
	if err := RemoveAvatar(filename); err != nil {
		t.Fatal(err)
	}
	if err := RemoveAvatar("../escape"); err == nil {
		t.Fatal("path traversal accepted")
	}
	if err := RemoveAvatar(""); err == nil {
		t.Fatal("empty filename accepted")
	}
}

func TestExtractAvatar(t *testing.T) {
	validRequest := func(t *testing.T, name, contentType string, content []byte) *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
		header.Set("Content-Type", contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("POST", "/avatar", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		response := httptest.NewRecorder()
		upload, closeFn, err := ExtractAvatar(response, req)
		if err != nil {
			t.Fatalf("ExtractAvatar() error = %v", err)
		}
		if upload == nil || upload.Filename != name || upload.Size != int64(len(content)) {
			t.Fatalf("unexpected extracted upload: %#v", upload)
		}
		if closeFn != nil {
			closeFn()
		}
		return response
	}
	validRequest(t, "avatar.png", "image/png", []byte("\x89PNG\r\n\x1a\ncontent"))

	var badBody bytes.Buffer
	writer := multipart.NewWriter(&badBody)
	part, err := writer.CreateFormFile("file", "text.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("plain text"))
	_ = writer.Close()
	badReq := httptest.NewRequest("POST", "/avatar", &badBody)
	badReq.Header.Set("Content-Type", writer.FormDataContentType())
	if _, closeFn, err := ExtractAvatar(httptest.NewRecorder(), badReq); err == nil {
		t.Fatal("non-image upload accepted")
	} else if closeFn != nil {
		closeFn()
	}

	missing := httptest.NewRequest("POST", "/avatar", strings.NewReader("not multipart"))
	if _, _, err := ExtractAvatar(httptest.NewRecorder(), missing); err == nil {
		t.Fatal("invalid multipart body accepted")
	}
}
