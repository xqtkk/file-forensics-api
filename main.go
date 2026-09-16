package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Загрузка файла + хеширование
	r.POST("/upload", func(c *gin.Context) {
		// 1. Получаем файл из формы
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "файл не передан. Используй поле 'file'",
			})
			return
		}

		// 2. Открываем файл
		openedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "не удалось открыть файл",
			})
			return
		}
		defer openedFile.Close()

		// 3. Создаём хешеры
		md5Hash := md5.New()
		sha1Hash := sha1.New()
		sha256Hash := sha256.New()

		// 4. Пишем файл одновременно во все три хешера
		// io.MultiWriter позволяет писать в несколько мест сразу
		multiWriter := io.MultiWriter(md5Hash, sha1Hash, sha256Hash)

		// 5. Копируем содержимое файла в хешеры
		// io.Copy читает файл потоково — не грузит весь файл в память
		_, err = io.Copy(multiWriter, openedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "ошибка при чтении файла",
			})
			return
		}

		// 6. Получаем хеши в виде hex-строк
		md5Sum := hex.EncodeToString(md5Hash.Sum(nil))
		sha1Sum := hex.EncodeToString(sha1Hash.Sum(nil))
		sha256Sum := hex.EncodeToString(sha256Hash.Sum(nil))

		// 7. Возвращаем результат
		c.JSON(http.StatusOK, gin.H{
			"filename": file.Filename,
			"size":     file.Size,
			"md5":      md5Sum,
			"sha1":     sha1Sum,
			"sha256":   sha256Sum,
		})
	})

	fmt.Println("Server started on :8080")
	r.Run(":8080")
}
