package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const destination = "test/download/"
const endpoint = "127.0.0.1:8000"
const accessKeyID = "minioadmin"
const secretAccessKey = "minioadmin"
const useSSL = false

func main() {
	http.HandleFunc("/uploads/", uploadFileHandler())
	http.HandleFunc("/downloadfiles/", downloadFileHandler())
	log.Print("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			renderError(w, "METHOD_NOT_ALLOWED", http.StatusInternalServerError)
			return
		}
		pathFile := r.URL.Path
		minioFile(pathFile)
	})
}
func downloadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			renderError(w, "METHOD_NOT_ALLOWED", http.StatusInternalServerError)
			return
		}
		//  filePath := r.URL.Path
		segments := strings.Split(r.URL.Path, "/")
		fileName := segments[len(segments)-1]
		minioFiledownload(fileName)
	})
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func minioFile(filePath string) {
	ctx := context.Background()
	// endpoint := "127.0.0.1:8000"
	// accessKeyID := "minioadmin"
	// secretAccessKey := "minioadmin"
	// useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Make a new bucket.
	bucketName := "fileup"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}
	segments := strings.Split(filePath, "/")
	fileName := segments[len(segments)-1]
	fullURLFile := strings.Replace(filePath, "/", "", 1)
	// Upload the zip file
	objectName := fileName
	//	filePath := "D:/minio-uploads-download/down/productionand uat (3).txt"
	extensions := strings.Split(fileName, ".")
	extension := extensions[len(extensions)-1]
	contentType := "application/" + extension

	// Upload the zip file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, fullURLFile, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
}

// //////To download files////////////////
func minioFiledownload(filename string) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "fulup"

	log.Println(bucketName)

	log.Println(filename)

	found, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Println("Error checking if minio bucket exists")
		log.Println(err)
		return
	}
	if found {
		log.Println("Bucket found")
	}

	objInfo, err := minioClient.StatObject(context.Background(), bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		log.Println("Error checking minio object metadata")
		log.Println(err)
		return
	} else {
		log.Println(objInfo)
	}

	downloadInfo, err := minioClient.GetObject(context.Background(), bucketName, filename, minio.GetObjectOptions{})

	if err != nil {
		log.Println("Error downloading minio object")
		log.Println(err)
		return
	}

	log.Println("Successfully downloaded minio bytes: ", downloadInfo)
	var (
		filePath string
	)
	filePath = destination + filename
	// Create blank file
	file, err := os.Create(filePath)
	if err != nil {
		log.Println(err)
		// renderError(w, "Unable_To_Create_File", http.StatusInternalServerError)
		file.Close()
		return
	}

	if _, err = io.Copy(file, downloadInfo); err != nil {
		fmt.Println(err)
		return
	}

}
