package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Подключаемся к БД
	if err := InitDB(); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer DB.Close()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "файл не передан. Используй поле 'file'"})
			return
		}

		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
			return
		}
		defer openedFile.Close()

		md5Hash := md5.New()
		sha1Hash := sha1.New()
		sha256Hash := sha256.New()

		multiWriter := io.MultiWriter(md5Hash, sha1Hash, sha256Hash)

		if _, err := io.Copy(multiWriter, openedFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка при чтении файла"})
			return
		}

		md5Sum := hex.EncodeToString(md5Hash.Sum(nil))
		sha1Sum := hex.EncodeToString(sha1Hash.Sum(nil))
		sha256Sum := hex.EncodeToString(sha256Hash.Sum(nil))

		// Сохраняем в БД
		var id int
		err = DB.QueryRow(c.Request.Context(),
			`INSERT INTO files (filename, size, md5, sha1, sha256)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id`,
			file.Filename, file.Size, md5Sum, sha1Sum, sha256Sum,
		).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось сохранить в БД"})
			return
		}

		// Пробуем извлечь EXIF: переоткрываем файл заново
		exifFile, err := file.Open()
		if err == nil {
			defer exifFile.Close()
			if meta := ExtractExif(exifFile); meta != nil {
				_, err = DB.Exec(c.Request.Context(),
					`INSERT INTO metadata
					 (file_id, camera_make, camera_model, datetime, width, height, latitude, longitude, iso)
					 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
					id, meta.CameraMake, meta.CameraModel, meta.DateTime,
					meta.Width, meta.Height, meta.Latitude, meta.Longitude, meta.ISO,
				)
				if err != nil {
					log.Printf("не удалось сохранить metadata: %v", err)
				}
			}
		}

		response := gin.H{
			"id":       id,
			"filename": file.Filename,
			"size":     file.Size,
			"md5":      md5Sum,
			"sha1":     sha1Sum,
			"sha256":   sha256Sum,
		}

		// Перечитываем EXIF ещё раз для ответа
		if exifFile, err := file.Open(); err == nil {
			defer exifFile.Close()
			if meta := ExtractExif(exifFile); meta != nil {
				response["metadata"] = meta
			}
		}

		c.JSON(http.StatusOK, response)
	})

	// История проверок
	r.GET("/files", func(c *gin.Context) {
		rows, err := DB.Query(c.Request.Context(),
			`SELECT id, filename, size, sha256, created_at
			 FROM files
			 ORDER BY created_at DESC
			 LIMIT 50`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка запроса к БД"})
			return
		}
		defer rows.Close()

		type FileRecord struct {
			ID        int    `json:"id"`
			Filename  string `json:"filename"`
			Size      int64  `json:"size"`
			SHA256    string `json:"sha256"`
			CreatedAt string `json:"created_at"`
		}

		var files []FileRecord
		for rows.Next() {
			var f FileRecord
			var createdAt interface{}
			if err := rows.Scan(&f.ID, &f.Filename, &f.Size, &f.SHA256, &createdAt); err != nil {
				continue
			}
			f.CreatedAt = fmt.Sprintf("%v", createdAt)
			files = append(files, f)
		}

		c.JSON(http.StatusOK, gin.H{
			"count": len(files),
			"files": files,
		})
	})

	fmt.Println("Server started on :8080")
	r.Run(":8080")
}