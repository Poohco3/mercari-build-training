package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

	//"encoding/json"

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
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir                   = "images"
	DB_PATH                  = "/Users/poohco/mercari-build-training/db/mercari.splite3"
	getCategoryFromNameQuery = "SELECT id FROM categories WHERE name = $1"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

type Items struct {
	Items []Item `json:"items"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

/*func addItemToFile(name string, category string, image string) error {
	f, err := os.OpenFile("items.json", os.O_RDWR|os.O_CREATE, 0755) // os.O_RDWR：読み書き両用オープン,os.O_CREATE：正常の終了した時アクセス時刻などを記録
	if err != nil {
		return err
	}
	defer f.Close() //ファイルを閉じる

	itemData := Items{} // 今あるデータで挿入
	err = json.NewDecoder(f).Decode(&itemData)
	if err != nil && err != io.EOF {
		return err
	}

	// ハッシュ化
	h := sha256.New()
	h.Write([]byte(strings.Split(image, ".")[0]))
	hash := h.Sum(nil)
	image = fmt.Sprintf("%x", hash) + ".jpg"

	newItem := Item{Name: name, Category: category, Image: image} // 新しいアイテムを加える
	itemData.Items = append(itemData.Items, newItem)

	f.Truncate(0) // ファイルサイズ０に→ファイルを空にしている
	f.Seek(0, 0)  // 組み込まれたファイルの初めにいく
	//err = json.NewEncoder(f).Encode(itemData) // 新しいデータを書く //ストリームから読み込んだ時にjson.NewDecoderを使う。
	if err != nil {
		return err
	}

	return nil
}*/

/*func addItem(c echo.Context) error {
	var categoryID int
	name := c.FormValue("name")
	//c.Logger().Infof("Receive item: %s", name)
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}
	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	// Create a new SHA256 hash
	hash := sha256.New()
	hashInBytes := hash.Sum(nil)
	// Convert hash bytes to hex string
	hashString := hex.EncodeToString(hashInBytes)
	image_jpg := hashString + ".jpg"
	new_image, err := os.Create("images/" + image_jpg)
	if err != nil {
		return err
	}
	if _, err := io.Copy(new_image, src); err != nil {
		return err
	}

	err := addItemToFile(name, category, image) // ファイルに新しいものを追加する。
	if err != nil {
		c.Logger().Errorf("Error appending item to file: %s", err)
		res := Response{Message: "Error appending item to file"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	item := Item{Name: name, Category: category, Image: image_jpg}

	c.Logger().Infof("Receive item: %s. Category: %s. Image_Name: %s", name, category, image)

	message := fmt.Sprintf("item received: %s", name)
	db, err := sql.Open("sqlite3", DB_PATH) //データベースに接続
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer db.Close()
	row := db.QueryRow(getCategoryFromNameQuery, item.Category)
	err = row.Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("INSERT INTO categories (name) VALUES ($1)", item.Category)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}
			row := db.QueryRow(getCategoryFromNameQuery, item.Category)
			err = row.Scan(&categoryID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			} else {
				return err
			}
		}
	}

	_, err = db.Exec(cmd2, item.Name, categoryID, item.Image) //データベースに値を入れる
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}*/

func addItem(c echo.Context) error {
	var items Items
	var categoryID int

	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}
	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	// Create a new SHA256 hash
	hash := sha256.New()
	hashInBytes := hash.Sum(nil)
	// Convert hash bytes to hex string
	hashString := hex.EncodeToString(hashInBytes)
	image_jpg := hashString + ".jpg"
	new_image, err := os.Create("images/" + image_jpg)
	if err != nil {
		return err
	}
	// Copy the file content to the hash
	if _, err := io.Copy(new_image, src); err != nil {
		return err
	}
	item := Item{Name: name, Category: category, Image: image_jpg}
	c.Logger().Infof("Receive item: %s, %s", item.Name, item.Category, item.Image)
	message := fmt.Sprintf("item received: %s, %s, %s", item.Name, item.Category, item.Image)
	res := Response{Message: message}

	items.Items = append(items.Items, item)

	//f, err := os.OpenFile("items.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	db, err := sql.Open("sqlite3", DB_PATH)

	if err != nil {
		return err
	}
	//defer f.Close()
	defer db.Close()

	//output, err := json.Marshal(&items)
	row := db.QueryRow(getCategoryFromNameQuery, item.Category)
	err = row.Scan(&categoryID)
	if err != nil {
		//return err
		if err == sql.ErrNoRows {
			_, err = db.Exec("INSERT INTO categories (name) VALUES ($1)", item.Category)
			if err != nil {
				return err
			}
			row := db.QueryRow(getCategoryFromNameQuery, item.Category)
			err = row.Scan(&categoryID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	//_, err = f.Write(output)
	cmd2 := "INSERT INTO items (name, category_id, image_name) VALUES ($1, $2, $3)"
	_, err = db.Exec(cmd2, item.Name, categoryID, item.Image)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
func Error() {
	panic("unimplemented")
}

func getItem(c echo.Context) error {
	var items Items
	/*f, err := os.Open("items.json") // ファイルを開く
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
	}*/
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT items.name, categories.name, image_name FROM items JOIN categories ON items.category_id = categories.id") //itemsTableから値を取ってくる。
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, category, image string
		err := rows.Scan(&name, &category, &image)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}
		item := Item{Name: name, Category: category, Image: image}
		items.Items = append(items.Items, item)
	} //整列して出力
	return c.JSON(http.StatusOK, items)
}

func getItemId(c echo.Context) error {
	var items Items
	id, _ := strconv.Atoi(c.Param("id")) // アイテムのidを取得
	//f, err := os.Open("items.json")      // ファイルを開く
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer db.Close()

	cmd := "SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?"
	rows, err := db.Query(cmd, id)
	if err != nil {
		//c.Logger().Errorf("Error decoding file: %s", err)
		//res := Response{Message: "Error decoding file"}
		return c.JSON(http.StatusInternalServerError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, category, image string
		if err := rows.Scan(&name, &category, &image); err != nil {
			return err
		}
		item := Item{Name: name, Category: category, Image: image}
		items.Items = append(items.Items, item)
	}
	/*itemData := ItemData{}
	err = json.NewDecoder(f).Decode(&itemData)
	if err != nil && err != io.EOF {
		//c.Logger().Errorf("Error decoding file: %s", err)
		//res := Response{Message: "Error decoding file"}
		return c.JSON(http.StatusInternalServerError, err)
	}
	if id <= 0 {
		return c.JSON(http.StatusOK, itemData.Items[id])
	} //エラー*/
	return c.JSON(http.StatusOK, items.Items[id-1])
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

func searchItem(c echo.Context) error {
	var items Items
	keyword := c.FormValue("keyword")
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		return err
	}
	defer db.Close()

	cmd := "SELECT items.name, categories.name, items.image_name FROM items JOIN categories ON items.category_id = categories.id WHERE items.name LIKE ?"
	rows, err := db.Query(cmd, "%"+keyword+"%")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, category, image string
		if err := rows.Scan(&name, &category, &image); err != nil {
			return err
		}
		item := Item{Name: name, Category: category, Image: image}
		items.Items = append(items.Items, item)
	}
	return c.JSON(http.StatusOK, items)
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
	e.GET("/search", searchItem)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
