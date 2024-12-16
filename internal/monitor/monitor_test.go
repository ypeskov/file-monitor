package monitor_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"ypeskov/file-monitor/internal/monitor"
)

// cleanDir удаляет содержимое указанной директории
func cleanDir(dir string) error {
	return os.RemoveAll(dir)
}

// createTestEnvironment создает тестовую структуру директорий и файлов
func createTestEnvironment(baseDir string) error {
	// Создаем базовую директорию
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}

	// Создаем поддиректории
	subDirs := []string{"subdir1", "subdir2"}
	for _, subDir := range subDirs {
		path := filepath.Join(baseDir, subDir)
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}

		// Создаем по два файла в каждой поддиректории
		for i := 1; i <= 2; i++ {
			filePath := filepath.Join(path, "file"+string(rune('0'+i))+".txt")
			if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestMonitorDirectory(t *testing.T) {
	testDir := "testdirs"

	// 1. Очищаем тестовую директорию перед запуском
	if err := cleanDir(testDir); err != nil {
		t.Fatalf("Failed to clean test directory: %v", err)
	}
	defer cleanDir(testDir) // Убираем тестовую директорию после завершения теста

	// 2. Создаем тестовую структуру файлов и директорий
	if err := createTestEnvironment(testDir); err != nil {
		t.Fatalf("Failed to set up test environment: %v", err)
	}

	// 3. Запускаем мониторинг
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	wg.Add(1) // Увеличиваем счетчик до запуска горутины
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Recovered from panic in MonitorDirectory: %v", r)
			}
			wg.Done() // Гарантированное завершение WaitGroup
		}()
		monitor.MonitorDirectory(testDir, stopChan, &wg)
	}()

	// Даем немного времени на запуск мониторинга
	time.Sleep(500 * time.Millisecond)

	// 4. Генерируем события
	newFilePath := filepath.Join(testDir, "newfile.txt")
	if err := os.WriteFile(newFilePath, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	existingFilePath := filepath.Join(testDir, "subdir1", "file1.txt")
	if err := os.WriteFile(existingFilePath, []byte("updated content"), 0644); err != nil {
		t.Fatalf("Failed to modify existing file: %v", err)
	}

	if err := os.Remove(existingFilePath); err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// Даем немного времени для обработки событий
	time.Sleep(1 * time.Second)

	// 5. Завершаем мониторинг
	close(stopChan)
	wg.Wait()

	t.Log("Test completed.")
}
