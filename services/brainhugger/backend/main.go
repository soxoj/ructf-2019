package main

import (
	"brainhugger/backend/cbc"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var taskManager TasksManager
var usersManager UsersManager
const configPath = "config"

type NewTask struct {
	Source string
	Stdin string
	Token string
}

type NewUser struct {
	Password string
}

type TaskResponse struct {
	Stdout string
	Status int
	Error string
}

func handleRunTask(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	ok, ownerId, err := usersManager.ValidateCookies(r.Cookies())
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if !ok {
		w.WriteHeader(403)
		return
	}
	var newTask NewTask
	err = json.Unmarshal(data, &newTask)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	taskId := taskManager.AddTask(newTask.Source, newTask.Token, []byte(newTask.Stdin), ownerId)
	if _, err := w.Write([]byte(fmt.Sprintf("{\"taskId\": %v}", taskId))); err != nil {
		panic(err)
	}
}

func handleTaskInfo(w http.ResponseWriter, r *http.Request) {
	rawTaskId := r.URL.Path[len("/task_info/"):]
	taskId, err := strconv.ParseUint(rawTaskId, 10, 64)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if taskId >= uint64(taskManager.TaskStorage.GetItemsCount()) {
		w.WriteHeader(404)
		return
	}
	task, err := taskManager.TaskInfo(uint(taskId))
	if err != nil {
		w.WriteHeader(404)
		return
	}
	ok, ownerId, err := usersManager.ValidateCookies(r.Cookies())
	if err != nil {
		w.WriteHeader(400)
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(403)
		return
	}

	if task.Token != token || task.OwnerId != uint(ownerId) || !ok {
		w.WriteHeader(403)
		return
	}

	taskResponse := TaskResponse{
		Stdout: task.Result.Stdout,
		Error:  task.Result.Error,
		Status: task.Status,
	}
	rawTaskResponse, err := JsonMarshalWithoutEscaping(taskResponse)
	if err != nil {
		log.Printf("Can not marshal task response: %v\n", taskResponse)
		w.WriteHeader(500)
		return
	}
	if _, err := w.Write(rawTaskResponse); err != nil {
		w.WriteHeader(400)
		log.Println("Can not write task response: " + err.Error())
	}
}

func handleRegUser(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	var newUser NewUser
	err = json.Unmarshal(data, &newUser)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	userId, cookie, err := usersManager.AddUser(newUser.Password, cbc.GenerateKey())
	if err != nil {
		w.WriteHeader(400)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "secret",
		Value: cookie,
	})
	http.SetCookie(w, &http.Cookie{
		Name: "uid",
		Value: fmt.Sprint(userId),
	})
}

func handleLoginUser(w http.ResponseWriter, r *http.Request) {
	rawUserId := r.URL.Query().Get("userid")
	password := r.URL.Query().Get("password")
	if rawUserId == "" || password == "" {
		w.WriteHeader(403)
		return
	}
	userId, err := strconv.ParseUint(rawUserId, 10, 64)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	ok, _, err := usersManager.ValidateCookies(r.Cookies())
	if err != nil {
		w.WriteHeader(400)
		return
	}

	if ok {
		secret, err := usersManager.GetForCookie(uint(userId))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name: "secret",
			Value: secret,
		})
		http.SetCookie(w, &http.Cookie{
			Name: "uid",
			Value: fmt.Sprint(userId),
		})
		return
	}

	ok, cookie, err := usersManager.LoginUser(uint(userId), password)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if ok {
		http.SetCookie(w, &http.Cookie{
			Name: "secret",
			Value: cookie,
		})
		http.SetCookie(w, &http.Cookie{
			Name: "uid",
			Value: fmt.Sprint(userId),
		})
	} else {
		w.WriteHeader(403)
		return
	}
	w.WriteHeader(200)
}

func main() {
	config, err := ParseConfig(configPath)
	if err != nil {
		panic("can not parse config: " + err.Error())
	}
	if err := taskManager.Init(config.TasksDir, config.BrainHugExecutorPath, config.MaxItemsCount); err != nil {
		panic(err)
	}
	usersManager.Init(config.UsersDir, config.MaxItemsCount)
	http.HandleFunc("/run_task", handleRunTask)
	http.HandleFunc("/task_info/", handleTaskInfo)
	http.HandleFunc("/register", handleRegUser)
	http.HandleFunc("/check", handleLoginUser)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.ServerHost, config.ServerPort), nil))
}
