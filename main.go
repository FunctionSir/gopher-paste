/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2024-12-07 22:24:35
 * @LastEditTime: 2024-12-20 23:06:54
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /gopher-paste/main.go
 */

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/ini.v1"
)

const (
	VER               string = "0.0.1"
	CODENAME          string = "SatenRuiko"
	CONTEXT_TYPE_HTML string = "text/html"
	ID_CHARSET        string = "0123456789"
	META_TOKEN_POS    int    = 0
	META_CT_POS       int    = 1
	META_EXP_POS      int    = 2
	META_LM_POS       int    = 3
)

var (
	Addr               string = "127.0.0.1:6450"
	HomePage           string = "index.html"
	PastesDir          string = "pastes"
	IdLen              int    = 8
	DefaultExpiration  int    = 24
	DefaultContentType string = "text/plain"
	BaseURL            string = "http://127.0.0.1:6450/"
	CleanGap           int    = 900
)

func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || stat.IsDir() {
		return false
	}
	return true
}

func LogFatalln(s string) {
	c := color.New(color.FgHiRed, color.Underline)
	log.Fatalln(c.Sprint(s))
}

func LogWarnln(s string) {
	c := color.New(color.FgHiYellow)
	log.Println(c.Sprint(s))
}

func LogInfoln(s string) {
	c := color.New(color.FgHiGreen)
	log.Println(c.Sprint(s))
}

func getHomePage(c *gin.Context) {
	f, _ := os.ReadFile(HomePage)
	c.Data(http.StatusOK, CONTEXT_TYPE_HTML, f)
}

func isValidId(id *string) bool {
	if len(*id) <= 0 {
		return false
	}
	for _, x := range *id {
		if !strings.ContainsRune(ID_CHARSET, x) || x == '.' || x == '/' || x == '-' {
			return false
		}
	}
	return true
}

func getPaste(c *gin.Context) {
	id := c.Params.ByName("id")
	if id == "favicon.ico" {
		c.String(http.StatusOK, "")
		return
	}
	// Judge ID is valid or not.
	if !isValidId(&id) {
		c.String(http.StatusBadRequest, "400 Illegal ID")
		return
	}
	// Is file exists or not
	pastePath := path.Join(PastesDir, id)
	pasteMeta := pastePath + "-metadata"
	if !FileExists(pastePath) || !FileExists(pasteMeta) {
		c.String(http.StatusNotFound, "404 Not Found")
		return
	}
	paste, err := os.ReadFile(pastePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "500 Internal Server Error")
		return
	}
	meta, err := os.ReadFile(pasteMeta)
	if err != nil {
		c.String(http.StatusInternalServerError, "500 Internal Server Error")
		return
	}
	info := strings.Split(string(meta), "\n")
	ct := c.Query("ct")
	if ct == "" {
		ct = info[META_CT_POS]
	}
	c.Data(http.StatusOK, ct, paste)

}

func getMaxSize(exp int) int {
	if exp == 0 || exp >= 2161 {
		return 1 << 16
	}
	if exp >= 721 {
		return 1 << 17
	}
	if exp >= 169 {
		return 1 << 18
	}
	if exp >= 73 {
		return 1 << 19
	}
	if exp >= 25 {
		return 1 << 20
	}
	return 1 << 21
}

func genId() string {
	result := ""
	idRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < IdLen; i++ {
		result = result + string(ID_CHARSET[idRand.Intn(len(ID_CHARSET))])
	}
	return result
}

func createPaste(c *gin.Context) {
	expStr := c.PostForm("expiration")
	if len(expStr) >= 5 {
		c.String(http.StatusBadRequest, "400 Expiration Too Long")
		return
	}
	var err error
	exp := DefaultExpiration
	if len(expStr) > 0 {
		exp, err = strconv.Atoi(expStr)
		if err != nil || exp < 0 {
			c.String(http.StatusBadRequest, "400 Illegal Expiration")
			return
		}
	}
	maxSize := getMaxSize(exp)
	dataStr := c.PostForm("data")
	if len(dataStr) <= 0 {
		c.String(http.StatusBadRequest, "400 Empty Payload")
		return
	}
	if len(dataStr) > maxSize {
		c.String(http.StatusRequestEntityTooLarge, "413 Payload Too Large")
		return
	}
	encodedStr := c.PostForm("encoded")
	data := []byte(dataStr)
	if encodedStr == "true" {
		data, err = base64.StdEncoding.DecodeString(dataStr)
		if err != nil {
			c.String(http.StatusBadRequest, "400 Illegal Base64 String")
			return
		}
	}
	id := genId()
	for FileExists(path.Join(PastesDir, id)) {
		id = genId()
	}
	pasteFile := path.Join(PastesDir, id)
	pasteMeta := pasteFile + "-metadata"
	token := uuid.NewString()
	ct := c.PostForm("content-type")
	if ct == "" {
		ct = DefaultContentType
	}
	lastMod := fmt.Sprintf("%d", time.Now().Unix())
	os.WriteFile(pasteFile, data, os.ModePerm)
	metaStr := token + "\n" + ct + "\n" + strconv.Itoa(exp) + "\n" + lastMod + "\n"
	os.WriteFile(pasteMeta, []byte(metaStr), os.ModePerm)
	c.String(http.StatusOK, BaseURL+id+"\n"+token)
}

func delPaste(c *gin.Context) {
	id := c.Param("id")
	if !isValidId(&id) {
		c.String(http.StatusBadRequest, "400 Illegal ID")
		return
	}
	token := c.Query("token")
	if len(token) <= 0 {
		c.String(http.StatusUnauthorized, "401 No Token Provided")
		return
	}
	pasteFile := path.Join(PastesDir, id)
	pasteMeta := pasteFile + "-metadata"
	if !FileExists(pasteFile) || !FileExists(pasteMeta) {
		c.String(http.StatusNotFound, "404 Not Found")
		return
	}
	metadata, err := os.ReadFile(pasteMeta)
	if err != nil {
		c.String(http.StatusInternalServerError, "500 Internal Server Error")
		return
	}
	correctToken := strings.Split(string(metadata), "\n")[META_TOKEN_POS]
	if correctToken != token {
		c.String(http.StatusForbidden, "403 Wrong Token")
		return
	}
	err = os.Remove(pasteFile)
	if err != nil {
		c.String(http.StatusInternalServerError, "500 Internal Server Error")
		return
	}
	err = os.Remove(pasteMeta)
	if err != nil {
		c.String(http.StatusInternalServerError, "500 Internal Server Error")
		return
	}
	c.String(http.StatusOK, "200 OK")
}

func modifyPaste(c *gin.Context) {
	id := c.Param("id")
	if !isValidId(&id) {
		c.String(http.StatusBadRequest, "400 Illegal ID")
		return
	}
	token := c.PostForm("token")
	if len(token) <= 0 {
		c.String(http.StatusUnauthorized, "401 No Token Provided")
		return
	}
	pasteFile := path.Join(PastesDir, id)
	pasteMeta := pasteFile + "-metadata"
	if !FileExists(pasteFile) || !FileExists(pasteMeta) {
		c.String(http.StatusNotFound, "404 Not Found")
		return
	}
	metadata, err := os.ReadFile(pasteMeta)
	if err != nil {
		c.String(http.StatusInternalServerError, "500 Internal Server Error")
		return
	}
	metaValStrs := strings.Split(string(metadata), "\n")
	correctToken := metaValStrs[META_TOKEN_POS]
	if correctToken != token {
		c.String(http.StatusForbidden, "403 Wrong Token")
		return
	}
	exp, _ := strconv.Atoi(metaValStrs[META_EXP_POS])
	maxSize := getMaxSize(exp)
	dataStr := c.PostForm("data")
	if len(dataStr) <= 0 {
		c.String(http.StatusBadRequest, "400 Empty Payload")
		return
	}
	if len(dataStr) > maxSize {
		c.String(http.StatusRequestEntityTooLarge, "413 Payload Too Large")
		return
	}
	encodedStr := c.PostForm("encoded")
	data := []byte(dataStr)
	if encodedStr == "true" {
		data, err = base64.StdEncoding.DecodeString(dataStr)
		if err != nil {
			c.String(http.StatusBadRequest, "400 Illegal Base64 String")
			return
		}
	}
	ct := c.PostForm("content-type")
	if ct == "" {
		ct = DefaultContentType
	}
	lastMod := fmt.Sprintf("%d", time.Now().Unix())
	os.WriteFile(pasteFile, data, os.ModePerm)
	metaStr := token + "\n" + ct + "\n" + strconv.Itoa(exp) + "\n" + lastMod + "\n"
	os.WriteFile(pasteMeta, []byte(metaStr), os.ModePerm)
	c.String(http.StatusOK, "200 OK")
}

func cleaner() {
	for {
		LogInfoln("Start to clean outdated pastes")
		dir, err := os.ReadDir(PastesDir)
		if err != nil {
			LogFatalln("Can not read the pastes dir: Error: " + err.Error())
		}
		toDel := make([]string, 0)
		for _, x := range dir {
			splited := strings.Split(x.Name(), "-")
			if len(splited) == 1 {
				curTime := time.Now().Unix()
				metapath := path.Join(PastesDir, splited[0]+"-metadata")
				metadata, err := os.ReadFile(metapath)
				if err != nil {
					LogWarnln("Can not read metadata: Error: " + err.Error())
					continue
				}
				metaValStrs := strings.Split(string(metadata), "\n")
				exp, err := strconv.Atoi(metaValStrs[META_EXP_POS])
				if err != nil {
					LogWarnln("Illegal value found: Error: " + err.Error())
					continue
				}
				if exp == 0 {
					continue
				}
				lastMod, err := strconv.Atoi(metaValStrs[META_LM_POS])
				if err != nil {
					LogWarnln("Illegal value found: Error: " + err.Error())
					continue
				}
				if (curTime - int64(lastMod)) > int64(exp)*3600 {
					toDel = append(toDel, path.Join(PastesDir, splited[0]))
					toDel = append(toDel, metapath)
				}
			}
		}
		delCnt := 0
		for _, x := range toDel {
			errCnt := 0
			err := os.Remove(x)
			for err != nil && errCnt < 3 {
				errCnt++
				err = os.Remove(x)
			}
			if err == nil {
				delCnt++
			}
		}
		LogInfoln(fmt.Sprintf("Cleaned %d outdated files", delCnt))
		time.Sleep(time.Duration(CleanGap) * time.Second)
	}
}

func prepare() {
	gin.SetMode(gin.ReleaseMode)
	if len(os.Args) <= 1 {
		return
	}
	conf, err := ini.Load(os.Args[1])
	if err != nil {
		LogFatalln("Can not load config file: Error: " + err.Error())
	}
	if !conf.HasSection("options") {
		LogFatalln("No section \"options\" found in conf file, but it's necessary")
	}
	sec := conf.Section("options")
	if sec.HasKey("Addr") {
		Addr = sec.Key("Addr").String()
	}
	if sec.HasKey("HomePage") {
		HomePage = sec.Key(HomePage).String()
	}
	if sec.HasKey("PastesDir") {
		PastesDir = sec.Key(HomePage).String()
	}
	if sec.HasKey("BaseURL") {
		BaseURL = sec.Key("BaseURL").String()
		if BaseURL[len(BaseURL)-1] != '/' {
			BaseURL = BaseURL + "/"
		}
	}
}

func hello() {
	c := color.New(color.FgHiBlue)
	c.Println("The Gopher Paste Pastebin Server [ Version: " + VER + " (" + CODENAME + ") ]")
}

func main() {
	hello()
	prepare()
	go cleaner()
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("/", getHomePage)
	router.GET("/:id", getPaste)
	router.POST("/", createPaste)
	router.DELETE("/:id", delPaste)
	router.PUT("/:id", modifyPaste)
	router.Run(Addr)
}
