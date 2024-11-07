package main

import (
	"log"
	"net/http"
	"time"

	"bajetapp/config"
	"bajetapp/mwr"
	"bajetapp/routes"
	"bajetapp/services"
	"bajetapp/utils"
	"bajetapp/views/pages"

	"bajetapp/db"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

type GoPlaygorundValidator struct {
	validator *validator.Validate
}

func (cv *GoPlaygorundValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("error loading config: %w", err)
	}

	// Initialize Database
	_, mongo, err := db.ConnectMongoDB(config.MongoURI, config.MongoDatabase)
	if err != nil {
		log.Fatal("error connecting to MongoDB: %w", err)
	}

	e := echo.New()
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())
	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.CSRFWithConfig(
		middleware.CSRFConfig{
			TokenLookup: "form:_csrf",
		},
	))
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.ContextTimeout(time.Second * time.Duration(config.ContextTimeoutSecond)))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(config.SessionSecret))))

	e.Debug = true
	e.Validator = &GoPlaygorundValidator{validator: validator.New()}
	// e.HTTPErrorHandler = func(err error, c echo.Context) {
	// 	code := http.StatusInternalServerError
	// 	if he, ok := err.(*echo.HTTPError); ok {
	// 		code = he.Code
	// 	}
	// 	c.Logger().Error(err)
	// 	// c.HTML(code, "error", nil)
	// }

	staticRoot := e.Group("/static")
	staticRoot.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "./views/static",
		Browse: false,
	}))

	// Routes
	e.GET("/", handleLogin)
	e.GET("/main", handleMain, mwr.AuthMiddleware)
	e.GET("/add", handleAddPage, mwr.AuthMiddleware)

	authService := services.NewAuthService(mongo, config.GoogleClientID, config.GoogleClientSecret, config.RedirectURL)
	trxService := services.NewTransactionService(mongo)

	routes.NewAuthRoutes(authService, e)
	routes.NewTransactionRoutes(trxService, e)

	e.Logger.Fatal(e.Start(":8080"))
}

func handleMain(c echo.Context) error {
	if !utils.IsAuthenticated(c) {
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	cpmnt := pages.MainPage()
	Render(c, &cpmnt)
	return nil
}

func handleAddPage(c echo.Context) error {
	cpmnt := pages.AddTransaction(
		c.Get(middleware.DefaultCSRFConfig.ContextKey).(string),
		[]string{"🍔 Food", "🚗 Transport", "🎉 Entertainment", "🛍️ Shopping", "🔧 Others"},
		[]string{"💼 Salary", "📈 Business", "📊 Investment", "🎁 Gift", "🏥 Other"},
	)

	Render(c, &cpmnt)
	return nil
}

// Login page
func handleLogin(c echo.Context) error {
	if utils.IsAuthenticated(c) {
		c.Redirect(http.StatusTemporaryRedirect, "/main")
	}

	cpmnt := pages.Login()
	Render(c, &cpmnt)
	return nil
}

func Render(echoCtx echo.Context, component *templ.Component) {
	(*component).Render(echoCtx.Request().Context(), echoCtx.Response().Writer)
}
