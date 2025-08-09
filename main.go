package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/syrlramadhan/database-anggota-api/config"
	"github.com/syrlramadhan/database-anggota-api/controller"
	"github.com/syrlramadhan/database-anggota-api/helper"
	"github.com/syrlramadhan/database-anggota-api/repository"
	"github.com/syrlramadhan/database-anggota-api/service"
)

func main() {
	var writer http.ResponseWriter
	errEnv := godotenv.Load()
	if errEnv != nil {
		helper.WriteJSONError(writer, http.StatusInternalServerError, errEnv.Error())
		return
	}

	port := os.Getenv("APP_PORT")
	fmt.Println("api running on port:" + port)

	db, err := config.ConnectToDatabase()
	if err != nil {
		helper.WriteJSONError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	router := httprouter.New()

	memberRepo := repository.NewMemberRepository()
	memberService := service.NewMemberService(memberRepo, db)
	memberController := controller.NewMemberController(memberService)

	router.POST("/api/member", memberController.AddMember)
	router.GET("/api/member", memberController.GetAllMember)
	router.GET("/api/member/:id", memberController.GetMemberById)
	router.PUT("/api/member/:id", memberController.UpdateMember)
	router.DELETE("/api/member/:id", memberController.DeleteMember)
	router.POST("/api/member/login", memberController.Login)
	router.POST("/api/member/token", memberController.LoginToken)

	handler := corsMiddleware(router)

	server := http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	errServer := server.ListenAndServe()
	if errServer != nil {
		helper.WriteJSONError(writer, http.StatusInternalServerError, errServer.Error())
		return
	}
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
