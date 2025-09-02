package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/syrlramadhan/database-anggota-api/config"
	"github.com/syrlramadhan/database-anggota-api/controller"
	"github.com/syrlramadhan/database-anggota-api/repository"
	"github.com/syrlramadhan/database-anggota-api/service"
)

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		fmt.Printf("Error loading .env file: %v\n", errEnv)
		return
	}

	port := os.Getenv("APP_PORT")
	fmt.Println("api running on port:" + port)

	db, err := config.ConnectToDatabase()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	router := httprouter.New()

	memberRepo := repository.NewMemberRepository()
	memberService := service.NewMemberService(memberRepo, db)
	memberController := controller.NewMemberController(memberService)

	// Notification system
	notificationRepo := repository.NewNotificationRepository()
	notificationService := service.NewNotificationService(notificationRepo, memberRepo, db)
	notificationController := controller.NewNotificationController(notificationService)

	// Member CRUD operations
	router.POST("/api/member", memberController.AddMember)
	router.GET("/api/member", memberController.GetAllMember)
	router.PUT("/api/member/:id", memberController.UpdateMember)
	router.PUT("/api/member/:id/notify", memberController.UpdateMemberWithNotification)
	router.DELETE("/api/member/:id", memberController.DeleteMember)

	// Authentication
	router.POST("/api/auth/login", memberController.Login)
	router.POST("/api/auth/token", memberController.LoginToken)

	// Profile management
	router.GET("/api/profile", memberController.GetProfile)
	router.PUT("/api/profile/password", memberController.SetPassword)
	router.PUT("/api/profile/complete", memberController.CompleteProfile)

	// Notification routes
	router.GET("/api/notifications", notificationController.GetNotifications)
	router.PUT("/api/notifications/:id/read", notificationController.MarkNotificationAsRead)
	router.GET("/api/notifications/unread/count", notificationController.GetUnreadNotificationCount)

	// Status change routes
	router.POST("/api/status-change/request", notificationController.CreateStatusChangeRequest)
	router.PUT("/api/status-change/:id/accept", notificationController.AcceptStatusChangeRequest)
	router.PUT("/api/status-change/:id/reject", notificationController.RejectStatusChangeRequest)

	// Route untuk mengakses file upload (gambar)
	router.GET("/uploads/:filename", serveUploadedFile)

	handler := corsMiddleware(router)

	server := http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	errServer := server.ListenAndServe()
	if errServer != nil {
		fmt.Printf("Error starting server: %v\n", errServer)
		return
	}
}

// serveUploadedFile handles serving uploaded files from the uploads directory
func serveUploadedFile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filename := ps.ByName("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Construct the file path
	filePath := filepath.Join("uploads", filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
