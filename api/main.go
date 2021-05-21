package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"
	"time"

	"gitea.com/go-chi/binding"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"github.com/the-redback/go-oneliners"
)

var engine *xorm.Engine

type Engineer struct {
	Username string `json:"username" xorm:"pk not null unique"`

	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`

	City     string `json:"city"`
	Division string `json:"division"`

	Position string `json:"position"`

	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
	DeletedAt time.Time `xorm:"deleted"`
	Version   int       `xorm:"version"`
}

func JSON(w http.ResponseWriter, r *http.Request, code int, output interface{}) {
	render.Status(r, code)
	switch data := output.(type) {
	case string:
		render.PlainText(w, r, data)
	case []byte:
		render.Data(w, r, data)
	default:
		render.JSON(w, r, data)
	}
}

func Error(w http.ResponseWriter, r *http.Request, code int, err error) {
	log.Println(err)
	render.Status(r, code)
	render.JSON(w, r, err.Error())
}

// Engineers represents the list of engineers and authenticated users
var Engineers []Engineer
var authUser = make(map[string]string)

var srvr http.Server

func StartXormEngine() {
	var err error
	connStr := "user=postgres password=postgres host=127.0.0.1 port=5432 dbname=apiserver sslmode=disable"

	engine, err = xorm.NewEngine("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := os.Create("apiserver.log")
	if err != nil {
		log.Println(err)
	}
	logger := xorm.NewSimpleLogger(logFile)
	logger.ShowSQL(true)
	engine.SetLogger(logger)

	if engine.TZLocation, err = time.LoadLocation("Asia/Dhaka"); err != nil {
		log.Println(err)
	}
}

// Handler Functions....

func Welcome(w http.ResponseWriter, r *http.Request) {
	JSON(w, r, http.StatusOK, "Welcome")
}

func ShowAllEngineers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello")
	var engineers []Engineer
	if err := engine.Find(&engineers); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	JSON(w, r, http.StatusOK, engineers)
}

func ShowSingleEngineer(w http.ResponseWriter, r *http.Request) {
	engineer, ok := r.Context().Value(getStructName(Engineer{})).(*Engineer)
	if !ok {
		Error(w, r, http.StatusNotFound, errors.New("content not found"))
		return
	}

	JSON(w, r, http.StatusOK, engineer)
}

func AddNewEngineer(w http.ResponseWriter, r *http.Request) {
	engineer := new(Engineer)
	errs := binding.Bind(r, engineer)
	if errs.Len() > 0 {
		Error(w, r, http.StatusBadRequest, errs[0])
		return
	}

	if engineer.Username == "" {
		Error(w, r, http.StatusBadRequest, errors.New("username can't be empty"))
		return
	}

	newEngineer := new(Engineer)
	newEngineer.Username = engineer.Username
	if exist, _ := engine.Get(newEngineer); exist {
		Error(w, r, http.StatusConflict, errors.New("username already exists"))
		return
	}

	// Check if it exists in deleted accounts
	newEngineer = new(Engineer)
	newEngineer.Username = engineer.Username
	if exist, _ := engine.Unscoped().Get(newEngineer); exist {
		Error(w, r, http.StatusConflict, errors.New("username already exists"))
		return
	}

	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	oneliners.PrettyJson(engineer, "Engineer")

	if _, err := session.Insert(engineer); err != nil {
		if err = session.Rollback(); err != nil {
			Error(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := session.Commit(); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	JSON(w, r, http.StatusCreated, engineer)
}

func UpdateEngineerProfile(w http.ResponseWriter, r *http.Request) {
	engineer, ok := r.Context().Value(reflect.TypeOf(Engineer{}).Name()).(*Engineer)
	if !ok {
		Error(w, r, http.StatusNotFound, errors.New("content not found"))
		return
	}

	newEngineer := new(Engineer)
	errs := binding.Bind(r, newEngineer)
	if errs.Len() > 0 {
		Error(w, r, http.StatusBadRequest, errs[0])
		return
	}

	// Updated information assignment
	if newEngineer.FirstName != "" {
		engineer.FirstName = newEngineer.FirstName
	}
	if newEngineer.LastName != "" {
		engineer.LastName = newEngineer.LastName
	}
	if newEngineer.City != "" {
		engineer.City = newEngineer.City
	}
	if newEngineer.Division != "" {
		engineer.Division = newEngineer.Division
	}

	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	if _, err := session.ID(engineer.Username).Update(engineer); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		if err = session.Rollback(); err != nil {
			log.Println(err)
		}
		return
	}

	if err := session.Commit(); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	JSON(w, r, http.StatusOK, engineer)
}

func DeleteEngineer(w http.ResponseWriter, r *http.Request) {
	engineer, ok := r.Context().Value(reflect.TypeOf(Engineer{}).Name()).(*Engineer)
	if !ok {
		Error(w, r, http.StatusNotFound, errors.New("content not found"))
		return
	}

	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	if _, err := session.ID(engineer.Username).Delete(engineer); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		if err = session.Rollback(); err != nil {
			log.Println(err)
		}
		return
	}

	if err := session.Commit(); err != nil {
		Error(w, r, http.StatusInternalServerError, err)
		return
	}

	JSON(w, r, http.StatusOK, "deleted successfully")
}

// CreateInitialEngineerProfile Creating initial engineer profiles
func CreateInitialEngineerProfile() {
	Engineers = make([]Engineer, 0)
	engineer := Engineer{
		Username:  "masud",
		FirstName: "Masudur",
		LastName:  "Rahman",
		City:      "Madaripur",
		Division:  "Dhaka",
		Position:  "Software Engineer",
	}
	Engineers = append(Engineers, engineer)

	engineer = Engineer{
		Username:  "fahim",
		FirstName: "Fahim",
		LastName:  "Abrar",
		City:      "Chittagong",
		Division:  "Chittagong",
		Position:  "Software Engineer",
	}
	Engineers = append(Engineers, engineer)

	engineer = Engineer{
		Username:  "tahsin",
		FirstName: "Tahsin",
		LastName:  "Rahman",
		City:      "Chittagong",
		Division:  "Chittagong",
		Position:  "Software Engineer",
	}
	Engineers = append(Engineers, engineer)

	engineer = Engineer{
		Username:  "jenny",
		FirstName: "Jannatul",
		LastName:  "Ferdows",
		City:      "Chittagong",
		Division:  "Chittagong",
		Position:  "Software Engineer",
	}
	Engineers = append(Engineers, engineer)

	if exist, _ := engine.IsTableExist(new(Engineer)); !exist {
		if err := engine.CreateTables(new(Engineer)); err != nil {
			log.Fatalln(err)
		}
	}

	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		log.Fatalln(err)
	}

	for _, user := range Engineers {
		if _, err := session.Insert(&user); err != nil {
			if err = session.Rollback(); err != nil {
				log.Fatalln(err)
			}
		}
	}
	if err := session.Commit(); err != nil {
		log.Fatalln(err)
	}

	authUser["masud"] = "pass"
	authUser["admin"] = "admin"

}

func AssignValues(host, port string) {
	srvr.Addr = host + ":" + port
}

func getStructName(obj interface{}) string {
	if t := reflect.TypeOf(obj); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		engineer := new(Engineer)
		engineer.Username = chi.URLParam(r, "username")
		fmt.Println("name: ", engineer.Username)
		fmt.Println("path:", r.URL.String())

		exist, err := engine.Get(engineer)
		if err != nil {
			Error(w, r, http.StatusInternalServerError, err)
			return
		} else if !exist {
			Error(w, r, http.StatusNotFound, errors.New(fmt.Sprintf("user:%s not found", engineer.Username)))
			return
		}
		ctx := context.WithValue(r.Context(), getStructName(engineer), engineer)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func StartTheApp() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	srvr.Handler = r

	StartXormEngine()
	defer engine.Close()
	CreateInitialEngineerProfile()

	r.Get("/", Welcome)

	r.Route("/engineers", func(r chi.Router) {
		r.Use(middleware.BasicAuth("", authUser))
		r.Get("/", ShowAllEngineers)
		r.Post("/", AddNewEngineer)
		r.Route("/{username}", func(r chi.Router) {
			r.Use(UserCtx)
			r.Get("/", ShowSingleEngineer)
			r.Put("/", UpdateEngineerProfile)
			r.Delete("/", DeleteEngineer)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println("Just a middleware")
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/nothing", func(w http.ResponseWriter, r *http.Request) {
			JSON(w, r, http.StatusOK, "It's nothing")
		})
		r.Get("/something", func(w http.ResponseWriter, r *http.Request) {
			JSON(w, r, http.StatusOK, "It's something")
		})
	})

	// Mount mounts a completely new router to /admin route
	// or a existing one like
	// r.Mount("/admin", r)

	//r.Mount("/admin", func() chi.Router {
	//	r = chi.NewRouter()
	//	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	//		w.Write(nil)
	//	})
	//	return nil
	//}())

	log.Println("Starting the server")

	if err := srvr.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
	defer srvr.Shutdown(context.Background())

	log.Println("The server has been shut down...!")
}
