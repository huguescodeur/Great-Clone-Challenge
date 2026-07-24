package upload

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type SignedURLResponse struct {
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	APIKey    string `json:"apiKey"`
	CloudName string `json:"cloudName"`
	UploadURL string `json:"uploadUrl"`
	Folder    string `json:"folder"`
}

func HandleSignedURL(w http.ResponseWriter, r *http.Request) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	folder := os.Getenv("CLOUDINARY_FOLDER")
	if folder == "" {
		folder = "insta-lite"
	}

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		http.Error(w, "storage not configured", http.StatusServiceUnavailable)
		return
	}

	timestamp := time.Now().Unix()
	paramStr := fmt.Sprintf("folder=%s&timestamp=%d%s", folder, timestamp, apiSecret)
	h := sha1.Sum([]byte(paramStr))
	signature := fmt.Sprintf("%x", h)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignedURLResponse{
		Signature: signature,
		Timestamp: timestamp,
		APIKey:    apiKey,
		CloudName: cloudName,
		UploadURL: fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/auto/upload", cloudName),
		Folder:    folder,
	})
}
