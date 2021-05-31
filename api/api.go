package api

import (
	"crypto/sha256"
	"net/http"

	"strings"
	"fmt"

	"time"

	"github.com/gin-gonic/gin"
)

type User struct {
	username	string
	password	string
}

type Workload struct {
	workload_id string
	workload_name string
	filter string
	status string
	//var filered_images = make([]Image,5,5)
}

type Image struct{
	image_file string
	//workload_id string
	image_id string
	image_type string
}

var USERS = make(map[string]string)
var ACTIVE_WORKLOADS = make([]Workload,5,5)
var WORKLOADS = 0
var RUNNING_JOBS int
var FILTERED_IMAGES = make([]Image,5,5)

func Start() {
	router := gin.Default()
	router.POST("/login", Login)
	router.DELETE("/logout", Logout) // change to DELETE
	router.POST("/images", Upload)
	router.GET("/status", Status)
	router.POST("/workloads", Workloads)
	router.Run(":8080")
}

func Login(context *gin.Context) {
	username, password, hasAuth := context.Request.BasicAuth()
	if hasAuth { // user validation not yet implemented
		user := User{username: username, password: password}
		token := Hash(user)
		USERS[token] = username
		context.JSON(http.StatusOK, gin.H{"message": "Hi " + username + ", welcome to the DPIP System", "token": token})
	} else {
		context.AbortWithStatusJSON(http.StatusOK, gin.H{"status": false, "message": "could not verify your authorization"})
	}
}

func Logout(context *gin.Context) {
	token := getToken(context.Request.Header.Get("Authorization"))
	if username, flag := USERS[token]; flag {
		delete(USERS, token)
		context.JSON(http.StatusOK, gin.H{"message": "Bye " + username + ", your token has been revoked"})
	} else {
		abort(context)
	}
}

func Upload(context *gin.Context) {
	token := getToken(context.Request.Header.Get("Authorization"))
	if _, flag := USERS[token]; flag {
		file, err := context.FormFile("data")
		/*workload_id context.FormValue("workload")
		if workload_id != "" {
			context.String(http.StatusBadRequest, fmt.Sprintf("get form error: %s", err.Error()))
			return
		}*/
		if err != nil {
			context.String(http.StatusBadRequest, fmt.Sprintf("get form error: %s", err.Error()))
			return
		}
		if err := context.SaveUploadedFile(file, "images/" + file.Filename); err != nil {
			context.String(http.StatusBadRequest, fmt.Sprintf("upload file error: %s", err.Error()))
			return
		}
		image := Image{image_file:file.Filename, image_id:Hash(file.Filename), image_type:"original"}
		FILTERED_IMAGES = append(FILTERED_IMAGES, image)
		context.JSON(http.StatusOK, gin.H{"message": "An image has been successfully uploaded", "filename": image.image_file, "size": file.Size})
	} else {
		abort(context)
	}
}

func Status(context *gin.Context) {
	token := getToken(context.Request.Header.Get("Authorization"))
	if username, flag := USERS[token]; flag {
		context.JSON(http.StatusOK, gin.H{"message": "Hi " + username + ", the DPIP System is Up and Running", "time": time.Now().Format("01-02-2006 15:04:05"), "workloads" : WORKLOADS})
	} else {
		
	}
}

func Workloads(context *gin.Context){	
	token := getToken(context.Request.Header.Get("Authorization"))
	workload_name := strings.TrimRight(strings.Split(context.Request.Header.Get("Authorization"), " ")[2], "[GIN]")
	workload_id := Hash(workload_name)
	exists := false
	for i := 0; i < len(ACTIVE_WORKLOADS); i++{
		if ACTIVE_WORKLOADS[i].workload_id == workload_id{
			exists = true
		}
	}
	if exists == false{
		if _,flag := USERS[token]; flag{
			workload := addWorkload(context.Request.Header.Get("Authorization"))
			ACTIVE_WORKLOADS = append(ACTIVE_WORKLOADS, workload)
			WORKLOADS = WORKLOADS + 1
			context.JSON(http.StatusOK, gin.H{"Message": "Workload submitted", "workload id": workload.workload_id, "workload name": workload.workload_name, "status": workload.status, "filter": workload.filter})
		}else{
			abort(context)
		}
	}
}

func addWorkload(header string) Workload{
	workload_name := strings.TrimRight(strings.Split(header, " ")[2], "[GIN]")
	workload_id := Hash(workload_name)
	filter := strings.TrimRight(strings.Split(header, " ")[3], "[GIN]")
	workload := Workload{workload_id: workload_id, workload_name: workload_name, filter: filter, status: "scheduling"}
	
	return workload
}


func getToken(header string) string {
	token := strings.TrimRight(strings.Split(header, " ")[1], "[GIN]")

	return token
}

func Hash(object interface{}) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v", object)))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func abort(context *gin.Context) {
	context.AbortWithStatusJSON(http.StatusOK, gin.H{"status": false, "message": "No username registered with the given token. Please check your token and try again or log in"})
}
