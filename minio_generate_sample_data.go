package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	endpoint        = "localhost:9000"
	accessKeyID     = "minioadmin"
	secretAccessKey = "minioadmin"
	bucketName      = "go-test-bucket"
	location        = "us-east-1"
	useSSL          = false
)

var (
	uploadDir   = "./uploads"
	downloadDir = "./downloads"
)

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("❌ Error creating directory %s: %v", dir, err)
		}
	}
}

func createBucketIfNotExists(minioClient *minio.Client) error {
	ctx := context.Background()
	fmt.Printf("🔍 Checking if bucket '%s' exists...\n", bucketName)

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("❌ Error checking bucket existence: %v", err)
	}

	if !exists {
		fmt.Printf("📦 Creating bucket '%s'...\n", bucketName)
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
		if err != nil {
			return fmt.Errorf("❌ Error creating bucket: %v", err)
		}
		fmt.Printf("✅ Bucket '%s' created successfully\n", bucketName)
	} else {
		fmt.Printf("✅ Bucket '%s' already exists\n", bucketName)
	}

	// Set bucket policy for better logging
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": map[string]interface{}{"AWS": []string{"*"}},
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)},
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		fmt.Printf("⚠️ Could not create policy JSON: %v\n", err)
	} else {
		err = minioClient.SetBucketPolicy(ctx, bucketName, string(policyJSON))
		if err != nil {
			fmt.Printf("⚠️ Could not set bucket policy: %v\n", err)
		} else {
			fmt.Printf("🔒 Set bucket policy for %s\n", bucketName)
		}
	}

	return nil
}

func generateRandomFile() (string, string, int) {
	maxSize, _ := rand.Int(rand.Reader, big.NewInt(49000))
	fileSize := int(maxSize.Int64()) + 1000 // 1KB to 50KB
	// Generate a realistic filename
	fileExtensions := []string{".txt", ".md", ".log", ".csv", ".json"}
	filePrefix := gofakeit.Word()
	fileExt := fileExtensions[gofakeit.Number(0, len(fileExtensions)-1)]
	fileName := filePrefix + fileExt
	// Generate content
	paragraphs := fileSize / 100
	if paragraphs < 1 {
		paragraphs = 1
	}
	var contentBuilder strings.Builder
	for i := 0; i < paragraphs; i++ {
		contentBuilder.WriteString(gofakeit.Paragraph(5, 10, 10, " "))
		contentBuilder.WriteString("\n\n")
	}
	content := contentBuilder.String()

	filePath := filepath.Join(uploadDir, fileName)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Fatalf("❌ Error creating file %s: %v", filePath, err)
	}
	return fileName, filePath, len(content)
}

func uploadFile(minioClient *minio.Client, fileName, filePath string) (int64, error) {
	ctx := context.Background()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("❌ Error getting file stats: %v", err)
	}

	fmt.Printf("📤 Uploading %s (%d bytes)...\n", fileName, fileInfo.Size())

	// Add metadata to trigger more audit log details
	metadata := map[string]string{
		"x-amz-meta-demo-app":    "parseable-minio-demo",
		"x-amz-meta-file-size":   fmt.Sprintf("%d", fileInfo.Size()),
		"x-amz-meta-upload-time": time.Now().Format(time.RFC3339),
	}

	_, err = minioClient.FPutObject(ctx, bucketName, fileName, filePath, minio.PutObjectOptions{
		ContentType:  "text/plain",
		UserMetadata: metadata,
	})
	if err != nil {
		return 0, fmt.Errorf("❌ Error uploading file: %v", err)
	}

	fmt.Printf("✅ Successfully uploaded %s\n", fileName)

	// Delete local file after upload
	err = os.Remove(filePath)
	if err != nil {
		fmt.Printf("⚠️ Could not delete local file %s: %v\n", fileName, err)
	} else {
		fmt.Printf("🗑️ Deleted local file %s\n", fileName)
	}

	return fileInfo.Size(), nil
}

func downloadFile(minioClient *minio.Client, fileName string) (int64, error) {
	ctx := context.Background()
	downloadPath := filepath.Join(downloadDir, "downloaded_"+fileName)

	fmt.Printf("📥 Downloading %s...\n", fileName)

	err := minioClient.FGetObject(ctx, bucketName, fileName, downloadPath, minio.GetObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("❌ Error downloading file: %v", err)
	}

	fmt.Printf("✅ Successfully downloaded %s to %s\n", fileName, downloadPath)

	fileInfo, err := os.Stat(downloadPath)
	if err != nil {
		return 0, fmt.Errorf("❌ Error getting downloaded file stats: %v", err)
	}

	return fileInfo.Size(), nil
}

func listObjects(minioClient *minio.Client) ([]minio.ObjectInfo, error) {
	ctx := context.Background()
	fmt.Printf("📋 Listing objects in bucket %s:\n", bucketName)

	objectsList := []minio.ObjectInfo{}
	objectsCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return nil, fmt.Errorf("❌ Error listing objects: %v", object.Err)
		}
		objectsList = append(objectsList, object)
		fmt.Printf("   - %s (%d bytes, modified: %s)\n", object.Key, object.Size, object.LastModified)
	}

	return objectsList, nil
}

func performAdditionalOperations(minioClient *minio.Client) error {
	ctx := context.Background()
	fmt.Println("🔧 Performing additional MinIO operations for audit logs...")

	// Get bucket info
	fmt.Println("📊 Getting bucket information...")
	policy, err := minioClient.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		fmt.Println("⚠️ No bucket policy found (this is normal)")
	}
	fmt.Println("✅ Retrieved bucket policy: ", policy)

	// List objects with prefix
	fmt.Println("🔍 Listing objects with different parameters...")
	objectsCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    "sample",
		Recursive: false,
	})

	prefixCount := 0
	for range objectsCh {
		prefixCount++
	}
	fmt.Printf("📋 Found %d objects with 'sample' prefix\n", prefixCount)

	// Try to access a non-existent object (will generate 404 audit log)
	_, err = minioClient.GetObject(ctx, bucketName, "non-existent-file.txt", minio.GetObjectOptions{})
	if err != nil {
		fmt.Println("🔍 Attempted to access non-existent file (generates 404 audit log)")
	}

	// Get bucket location
	location, err := minioClient.GetBucketLocation(ctx, bucketName)
	if err != nil {
		fmt.Println("⚠️ Could not get bucket location")
	} else {
		fmt.Printf("🌍 Bucket location: %s\n", location)
	}

	return nil
}

func runDemo() {
	fmt.Println("🚀 Starting Parseable + MinIO Demo Application")
	fmt.Println(strings.Repeat("=", 50))

	// Ensure directories exist
	ensureDirExists(uploadDir)
	ensureDirExists(downloadDir)

	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("❌ Error initializing MinIO client: %v", err)
	}

	// Create bucket if it doesn't exist
	err = createBucketIfNotExists(minioClient)
	if err != nil {
		log.Fatalf("❌ Error with bucket operations: %v", err)
	}

	// Generate random number of files (1-5)
	max, _ := rand.Int(rand.Reader, big.NewInt(5))
	numFiles := int(max.Int64()) + 1
	fmt.Printf("📁 Generating %d random files...\n", numFiles)

	files := make([]string, numFiles)
	for i := 0; i < numFiles; i++ {
		fileName, _, size := generateRandomFile()
		files[i] = fileName
		fmt.Printf("   %d. %s (%d bytes)\n", i+1, fileName, size)
	}

	fmt.Println()

	// Upload all files
	for _, fileName := range files {
		filePath := filepath.Join(uploadDir, fileName)
		_, err := uploadFile(minioClient, fileName, filePath)
		if err != nil {
			log.Printf("❌ Error uploading file: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()

	// List all objects in bucket
	_, err = listObjects(minioClient)
	if err != nil {
		log.Printf("❌ Error listing objects: %v", err)
	}

	fmt.Println()

	// Download some files back
	filesToDownload := min(2, len(files))
	for i := 0; i < filesToDownload; i++ {
		_, err := downloadFile(minioClient, files[i])
		if err != nil {
			log.Printf("❌ Error downloading file: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()

	// Perform additional operations
	err = performAdditionalOperations(minioClient)
	if err != nil {
		log.Printf("❌ Error in additional operations: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Demo completed successfully!")
	fmt.Println("🔍 Check your Parseable dashboard at http://localhost:8000")
	fmt.Println("   - Stream: minio_audit (for MinIO audit logs)")
	fmt.Println("   - Stream: minio_log (for MinIO server logs)")
	fmt.Println("🗂️  Check your MinIO console at http://localhost:9001")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	runDemo()
}
