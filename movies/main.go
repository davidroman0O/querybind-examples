package main

import (
	"fmt"
	"net/http"
	"os"

	"embed"

	"github.com/davidroman0O/querybind"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	_ "embed"
)

//go:embed public/*.gohtml
var EmbedDirViews embed.FS

type Movie struct {
	Title  string
	Genre  string
	Year   string
	Rating string
}

var allMovies = []Movie{
	{"The Shawshank Redemption", "Drama", "1994", "R"},
	{"The Godfather", "Crime", "1972", "R"},
	{"Inception", "Sci-Fi", "2010", "PG-13"},
	{"Forrest Gump", "Drama", "1994", "PG-13"},
	{"The Dark Knight", "Action", "2008", "PG-13"},
	{"Pulp Fiction", "Crime", "1994", "R"},
	{"Fight Club", "Drama", "1999", "R"},
	{"The Matrix", "Sci-Fi", "1999", "R"},
	{"Goodfellas", "Crime", "1990", "R"},
	{"Interstellar", "Sci-Fi", "2014", "PG-13"},
	{"Whiplash", "Drama", "2014", "R"},
	{"The Prestige", "Mystery", "2006", "PG-13"},
	{"Parasite", "Thriller", "2019", "R"},
	{"1917", "War", "2019", "R"},
	{"Jojo Rabbit", "Comedy", "2019", "PG-13"},
	{"La La Land", "Musical", "2016", "PG-13"},
	{"Avengers: Endgame", "Action", "2019", "PG-13"},
	{"Toy Story 4", "Animation", "2019", "G"},
	{"The Lion King", "Animation", "1994", "G"},
	{"Jurassic Park", "Adventure", "1993", "PG-13"},
}

type QueryParams struct {
	Genres []string `querybind:"genres"`
	Years  []string `querybind:"years"`
}

func main() {
	app := fiber.New()

	var engine *html.Engine

	// embed won't be able to reload the files
	if os.Getenv("ENVIRONMENT") == "production" {
		engine = html.NewFileSystem(http.FS(EmbedDirViews), ".gohtml")
	} else {
		engine = html.New("./movies/public", ".gohtml")
		engine.Debug(true)
		engine.Reload(true)
	}

	// Pass the engine to the Views
	app = fiber.New(fiber.Config{
		Views: engine,
	})

	// Route for the main page
	app.Get("/", func(c *fiber.Ctx) error {

		params, err := querybind.Bind[QueryParams](c)
		if err != nil {
			// Handle error, could return HTTP 400 Bad Request
			return c.Status(fiber.StatusBadRequest).SendString("Invalid query parameters")
		}

		querybind.ResponseBind[QueryParams](c, *params)

		return c.Render("page", fiber.Map{
			"Movies":         applyFilters(params.Genres, params.Years, []string{}),
			"Genres":         uniqueGenres(allMovies),
			"Years":          uniqueYears(allMovies),
			"GenresSelected": params.Genres,
			"YearsSelected":  params.Years,
		}, "layout")
	})

	// Route for filtering movies
	app.Get("/filter", func(c *fiber.Ctx) error {

		genre := c.Query("genre")
		year := c.Query("year")

		params, err := querybind.Bind[QueryParams](c)
		if err != nil {
			// Handle error, could return HTTP 400 Bad Request
			return c.Status(fiber.StatusBadRequest).SendString("Invalid query parameters")
		}

		fmt.Println(genre, year)

		if len(genre) != 0 {
			params.Genres = append(params.Genres, genre)
		}
		if len(year) != 0 {
			params.Years = append(params.Years, year)
		}

		filteredMovies := applyFilters(params.Genres, params.Years, []string{})

		querybind.ResponseBind[QueryParams](c, *params, querybind.WithPath("/"))

		return c.Render("list", fiber.Map{
			"Movies":         filteredMovies,
			"GenresSelected": params.Genres,
			"YearsSelected":  params.Years,
		})
	})

	app.Get("/remove", func(c *fiber.Ctx) error {

		genre := c.Query("genre")
		year := c.Query("year")

		params, err := querybind.Bind[QueryParams](c)
		if err != nil {
			// Handle error, could return HTTP 400 Bad Request
			return c.Status(fiber.StatusBadRequest).SendString("Invalid query parameters")
		}

		fmt.Println(genre, year)

		if len(genre) != 0 {
			params.Genres = removeElement(params.Genres, genre)
		}
		if len(year) != 0 {
			params.Years = removeElement(params.Years, year)
		}

		filteredMovies := applyFilters(params.Genres, params.Years, []string{})

		querybind.ResponseBind[QueryParams](c, *params, querybind.WithPath("/"))

		return c.Render("list", fiber.Map{
			"Movies":         filteredMovies,
			"GenresSelected": params.Genres,
			"YearsSelected":  params.Years,
		})
	})

	// Start server
	panic(app.Listen(":3000"))
}

func getUniqueGenresAndYears(movies []Movie) ([]string, []string) {
	genreMap := make(map[string]bool)
	yearMap := make(map[string]bool)

	for _, m := range movies {
		genreMap[m.Genre] = true
		yearMap[m.Year] = true
	}

	genres := make([]string, 0, len(genreMap))
	years := make([]string, 0, len(yearMap))

	for genre := range genreMap {
		genres = append(genres, genre)
	}
	for year := range yearMap {
		years = append(years, year)
	}

	return genres, years
}

// Add this function outside the main function

func filterMovies(movies []Movie, genre, year string) []Movie {
	var filtered []Movie
	for _, movie := range movies {
		if (genre == "" || movie.Genre == genre) && (year == "" || movie.Year == year) {
			filtered = append(filtered, movie)
		}
	}
	return filtered
}

func applyFilters(genres, years, ratings []string) []Movie {
	var filteredMovies []Movie
	for _, movie := range allMovies {
		if (isInSlice(genres, movie.Genre) || len(genres) == 0) &&
			(isInSlice(years, movie.Year) || len(years) == 0) &&
			(isInSlice(ratings, movie.Rating) || len(ratings) == 0) {
			filteredMovies = append(filteredMovies, movie)
		}
	}
	return filteredMovies
}

// isInSlice checks if a given string is in a slice of strings.
func isInSlice(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func uniqueGenres(movies []Movie) []string {
	genreMap := make(map[string]bool)
	var genres []string
	for _, movie := range movies {
		if _, exists := genreMap[movie.Genre]; !exists {
			genres = append(genres, movie.Genre)
			genreMap[movie.Genre] = true
		}
	}
	return genres
}

func uniqueYears(movies []Movie) []string {
	yearMap := make(map[string]bool)
	var years []string
	for _, movie := range movies {
		if _, exists := yearMap[movie.Year]; !exists {
			years = append(years, movie.Year)
			yearMap[movie.Year] = true
		}
	}
	return years
}

func uniqueRatings(movies []Movie) []string {
	ratingMap := make(map[string]bool)
	var ratings []string
	for _, movie := range movies {
		if _, exists := ratingMap[movie.Rating]; !exists {
			ratings = append(ratings, movie.Rating)
			ratingMap[movie.Rating] = true
		}
	}
	return ratings
}

func removeElement(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // return the original slice if the value is not found
}
