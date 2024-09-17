package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Image_Name string `json:"image"`
}

type ItemData struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItemToFile(name string, category string, image string) error {
	f, err := os.OpenFile("items.json", os.O_RDWR|os.O_CREATE, 0755) // os.O_RDWR：読み書き両用オープン,os.O_CREATE：正常の終了した時アクセス時刻などを記録
	if err != nil {
		return err
	}
	defer f.Close() //ファイルを閉じる

	itemData := ItemData{} // 今あるデータで挿入
	err = json.NewDecoder(f).Decode(&itemData)
	if err != nil && err != io.EOF {
		return err
	}

	// ハッシュ化
	h := sha256.New()
	h.Write([]byte(strings.Split(image, ".")[0]))
	hash := h.Sum(nil)
	image = fmt.Sprintf("%x", hash) + ".jpg"

	newItem := Item{Name: name, Category: category, Image_Name: image} // 新しいアイテムを加える
	itemData.Items = append(itemData.Items, newItem)

	f.Truncate(0)                             // ファイルサイズ０に→ファイルを空にしている
	f.Seek(0, 0)                              // 組み込まれたファイルの初めにいく
	err = json.NewEncoder(f).Encode(itemData) // 新しいデータを書く //ストリームから読み込んだ時にjson.NewDecoderを使う。
	if err != nil {
		return err
	}

	return nil
}
func addItem(c echo.Context) error {
	name := c.FormValue("name")
	//c.Logger().Infof("Receive item: %s", name)
	category := c.FormValue("category")
	image := c.FormValue("image")

	err := addItemToFile(name, category, image) // ファイルに新しいものを追加する。
	if err != nil {
		c.Logger().Errorf("Error appending item to file: %s", err)
		res := Response{Message: "Error appending item to file"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	c.Logger().Infof("Receive item: %s. Category: %s. Image_Name: %s", name, category, image)

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItem(c echo.Context) error {
	f, err := os.Open("items.json") // ファイルを開く
	if err != nil {
		c.Logger().Errorf("Error opening file: %s", err)
		res := Response{Message: "Error opening file"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer f.Close()

	itemData := ItemData{}
	err = json.NewDecoder(f).Decode(&itemData)
	if err != nil && err != io.EOF {
		c.Logger().Errorf("Error decoding file: %s", err)
		res := Response{Message: "Error decoding file"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, itemData)
}

func getItemId(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id")) // アイテムのidを取得
	f, err := os.Open("items.json")      // ファイルを開く
	if err != nil {
		c.Logger().Errorf("Error opening file: %s", err)
		res := Response{Message: "Error opening file"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer f.Close()

	itemData := ItemData{}
	err = json.NewDecoder(f).Decode(&itemData)
	if err != nil && err != io.EOF {
		c.Logger().Errorf("Error decoding file: %s", err)
		res := Response{Message: "Error decoding file"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	if id <= 0 {
		return c.JSON(http.StatusOK, itemData.Items[id])
	} //エラー
	return c.JSON(http.StatusOK, itemData.Items[id-1])
}

func getImgs(c echo.Context) error {
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItem)
	e.GET("/items/:id", getItemId)
	e.GET("/image/:imageFilename", getImgs)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
