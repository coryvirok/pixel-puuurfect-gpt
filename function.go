package pp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type FilePayload struct {
	Index  string  `json:"index"`
	Styles string  `json:"styles"`
	Files  []*File `json:"files"`
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload FilePayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	bucketName := "pixel-puuurfect-prod"
	folderID := uuid.New().String()

	indexContent := payload.Index

	if len(indexContent) == 0 {
		http.Error(w, "missing index content", http.StatusBadRequest)
		return
	}

	err = uploadFile(ctx, client, bucketName, folderID+"/index.html", indexContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stylesContent := payload.Styles
	if len(stylesContent) > 0 {
		err = uploadFile(ctx, client, bucketName, folderID+"/styles.css", stylesContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(payload.Files) > 0 {
		for _, file := range payload.Files {
			err := uploadFile(ctx, client, bucketName, folderID+"/"+file.Name, file.Content)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	resp := fmt.Sprintf("{\"link\": \"https://storage.googleapis.com/%s/%s/index.html\"}", bucketName, folderID)
	fmt.Fprintf(w, resp)
}

func uploadFile(ctx context.Context, client *storage.Client, bucketName, objectName, content string) error {
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	w := obj.NewWriter(ctx)
	defer w.Close()

	contentType := getContentType(objectName)
	if len(contentType) != 0 {
		w.ObjectAttrs.ContentType = contentType
	}

	_, err := w.Write([]byte(content))
	return err
}

func getContentType(filename string) string {
	extension := filepath.Ext(filename)
	switch extension {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".pdf":
		return "application/pdf"
	default:
		return ""
	}
}
