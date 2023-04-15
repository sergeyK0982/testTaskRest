package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type IdEntry struct {
	RequestId    string
	Code         int
	Number       string
	Count        int
	CreationTime int64
}

var mux = http.NewServeMux()
var ttlSupport = false
var codeLen = 4
var duration = 300
var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

type SendReq struct {
	Number string `json:"number"`
}

type SendResponse struct {
	RequestId string `json:"requestId"`
	Code      int    `json:"code"`
}

func sendHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("sendHandler")
	var sendR SendReq
	json.NewDecoder(req.Body).Decode(&sendR)

	var newEntry = createNewEntry(sendR.Number)
	w.WriteHeader(201)
	w.Write(renderJSON(w, createResponse(newEntry)))

	setToRedis(newEntry)
}

func createResponse(newEntry IdEntry) SendResponse {
	var senRes SendResponse
	senRes.RequestId = newEntry.RequestId
	senRes.Code = newEntry.Code
	return senRes
}

func createNewEntry(number string) IdEntry {
	var newEntry IdEntry
	newEntry.RequestId = GenerateRequestID()
	newEntry.Number = number
	code, _ := strconv.Atoi(GenerateCheckCode())
	newEntry.Code = code
	newEntry.Count = 0
	newEntry.CreationTime = time.Now().UnixMilli()
	return newEntry
}

func renderJSON(w http.ResponseWriter, v interface{}) []byte {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return js
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
const separator = "-"

func GenerateRequestID() string {
	b := make([]byte, 36)
	for i := range b {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			b[i] = separator[0]
		} else {
			b[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
	}
	return string(b)
}

func GenerateCheckCode() string {
	temp := fmt.Sprint(time.Now().Nanosecond())
	return temp[:codeLen]
}

func verifyHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("verifyHandler")
	var sendR SendResponse
	json.NewDecoder(req.Body).Decode(&sendR)
	val := getFromRedis(sendR.RequestId)
	if val.Code == sendR.Code && val.RequestId == sendR.RequestId {
		if ttlSupport {
			currenttime := time.Now().UnixMilli()
			lifeTime := (currenttime - val.CreationTime) / 1000
			if lifeTime > int64(duration) {
				http.Error(w, "code expired", http.StatusBadRequest)
				return
			}
		}
		val.Count++
		if val.Count > 3 {
			http.Error(w, "More than 3 attempts. Code disabled", http.StatusTooManyRequests)
		} else {
			setToRedis(val)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(time.Now().Unix())
		}
	}
}

type Configuration struct {
	TtlSupport string
	CodeLen    string
	Duration   string
}

func readConf() {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		ttlSupport, _ = strconv.ParseBool(configuration.TtlSupport)
		codeLen, _ = strconv.Atoi(configuration.CodeLen)
		duration, _ = strconv.Atoi(configuration.Duration)
	}
}

func setToRedis(newEntry IdEntry) {
	b, err := json.Marshal(newEntry)
	if err != nil {
		fmt.Println(err)
		return
	}
	ctx := context.Background()
	err2 := client.Set(ctx, newEntry.RequestId, b, 0).Err()
	if err != nil {
		fmt.Println(err2)
	}

}

func getFromRedis(key string) IdEntry {
	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	var newEntry IdEntry
	if err == nil {
		json.Unmarshal([]byte(val), &newEntry)
		fmt.Println("entry - ", newEntry)
	}
	return newEntry
}

func main() {
	mux := http.NewServeMux()
	readConf()
	mux.HandleFunc("/api/v1/send", sendHandler)
	mux.HandleFunc("/api/v1/verify/", verifyHandler)
	http.ListenAndServe("localhost:8080", mux)
}
