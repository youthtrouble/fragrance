package controllers

import (
	"github.com/ichtrojan/thoth"
	"html/template"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/ichtrojan/fragrance/database"
	"github.com/ichtrojan/fragrance/models"
	"github.com/ichtrojan/fragrance/views"
)

var view *views.View

type CategoriesData struct {
	Categories []models.Category
}

type ScentsData struct {
	models.Scent
	Category string
}

type BottleData struct {
	models.Bottle
	Category string
	Scent    string
}

type BottleSize struct {
	models.BottleSize
}

type Fragrance struct {
	Image    string
	Category string
	Scent    string
	Bottle   string
	Sizes    []models.BottleSize
	Price    float64
}

func Home(w http.ResponseWriter, r *http.Request) {
	view = views.NewView("app", "home")
	must(view.Render(w, nil))
}

func Category(w http.ResponseWriter, r *http.Request) {
	view = views.NewView("app", "fragrance")

	db := database.Init()

	var categories []models.Category

	query := db.Find(&categories)

	defer query.Close()

	data := CategoriesData{
		Categories: categories,
	}

	must(view.Render(w, data))
}

func Scent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	category := vars["category"]

	view = views.NewView("app", "scent")

	db := database.Init()

	var categories []models.Category

	query := db.Find(&categories)

	defer query.Close()

	var slugs []string

	for _, category := range categories {
		slugs = append(slugs, category.Slug)
	}

	if !slugExists(slugs, category) {
		notFound(w, r)
		return
	}

	var scents []models.Scent

	var data []ScentsData

	query = db.Find(&scents)

	defer query.Close()

	for _, scent := range scents {
		data = append(data, ScentsData{
			Scent:    scent,
			Category: category,
		})
	}

	must(view.Render(w, data))
}

func Perfume(w http.ResponseWriter, r *http.Request) {
	view = views.NewView("app", "perfume")

	vars := mux.Vars(r)

	category := vars["category"]

	scent := vars["scent"]

	db := database.Init()

	var scents []models.Scent

	query := db.Find(&scents)

	defer query.Close()

	var slugs []string

	for _, scent := range scents {
		slugs = append(slugs, scent.Slug)
	}

	if !slugExists(slugs, scent) {
		notFound(w, r)
		return
	}

	var data []BottleData

	var bottles []models.Bottle

	query = db.Find(&bottles)

	defer query.Close()

	for _, bottle := range bottles {
		data = append(data, BottleData{
			Bottle:   bottle,
			Category: category,
			Scent:    scent,
		})
	}

	must(view.Render(w, data))
}

func Checkout(w http.ResponseWriter, r *http.Request) {
	view = views.NewView("app", "checkout")

	vars := mux.Vars(r)

	category := vars["category"]

	scent := vars["scent"]

	bottle := vars["bottle"]

	var fragrance models.Bottle

	var fragraceScent models.Scent

	db := database.Init()

	var bottles []models.Bottle

	query := db.Find(&bottles)

	defer query.Close()

	var slugs []string

	for _, bottle := range bottles {
		slugs = append(slugs, bottle.Slug)
	}

	if !slugExists(slugs, bottle) {
		notFound(w, r)
		return
	}

	query = db.Where("slug = ?", bottle).Preload("BottleSizes").First(&fragrance)

	query = db.Where("slug = ?", scent).First(&fragraceScent)

	defer query.Close()

	data := Fragrance{
		Image:    fragrance.Image,
		Category: category,
		Scent:    scent,
		Bottle:   bottle,
		Sizes:    fragrance.BottleSizes,
		Price:    fragraceScent.Price,
	}

	must(view.Render(w, data))
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	view := views.NewView("app", "testauth")

	adminId := GetSession(w, r)

	var admin models.Admin

	db := database.Init()

	query := db.Where("id = ?", adminId).First(&admin)

	defer query.Close()

	type User struct {
		Name string
	}

	var data = User{Name: admin.Name}

	must(view.Render(w, data))
}

func must(err error) {
	var logger, _ = thoth.Init("log")

	if err != nil {
		logger.Log(err)
		panic(err)
	}
}

func slugExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	logger, _ := thoth.Init("log")

	view, err := template.ParseFiles("views/errors/404.html")

	if err != nil {
		logger.Log(err)
		log.Fatal(err)
	}

	_ = view.Execute(w, nil)
}
