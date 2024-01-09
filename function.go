package pp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type FilePayload struct {
	Index string            `json:"index"`
	Files map[string]string `json:"files"`
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
	for filename, content := range payload.Files {
		err := uploadFile(ctx, client, bucketName, folderID+"/"+filename, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s/index.html", bucketName, folderID)
	fmt.Fprintf(w, url)
}

func uploadFile(ctx context.Context, client *storage.Client, bucketName, objectName, content string) error {
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	w := obj.NewWriter(ctx)
	defer w.Close()

	_, err := w.Write([]byte(content))
	return err
}
