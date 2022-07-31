package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
)

var log *zap.Logger

func init() {
	var err error

	config := zap.NewProductionConfig()

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	config.EncoderConfig = encoderConfig

	log, err = config.Build(zap.AddCallerSkip(1))

	if err != nil {
		panic(err)
	}
}

func LogInfo(message string, fields ...zap.Field) {
	log.Info(message, fields...)
}

func LogFatal(message string, fields ...zap.Field) {
	log.Fatal(message, fields...)
}

func LogDebug(message string, fields ...zap.Field) {
	log.Debug(message, fields...)
}

func LogError(message string, fields ...zap.Field) {
	log.Error(message, fields...)
}

type LocationInfo struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
	Ip  string `json:"ip"`
}

type publicIPAddr struct {
	PublicIP string `json:"public_ip"`
}

func applicationPort() string {
	if port, ok := os.LookupEnv("PORT"); ok {
		return port
	}
	return "80"
}

func readUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

func userPublicIP(w http.ResponseWriter, r *http.Request) {
	ip := readUserIP(r)

	response, _ := json.Marshal(publicIPAddr{
		PublicIP: ip,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Info(ip)
	w.Write(response)
}

func getLocation(w http.ResponseWriter, r *http.Request) {
	ip := readUserIP(r)
	lat := r.URL.Query().Get("lat")
	lon := r.URL.Query().Get("lon")
	data := &LocationInfo{
		lat,
		lon,
		ip,
	}

	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Info(string(b))
	response, _ := json.Marshal(publicIPAddr{
		PublicIP: ip,
	})
	w.Write(response)
}

func main() {

	portNumber := applicationPort()
	log.Info(fmt.Sprintf("Application started at port %s", portNumber))
	http.HandleFunc("/", userPublicIP)
	http.HandleFunc("/location", getLocation)
	http.ListenAndServe(":"+portNumber, nil)
}
