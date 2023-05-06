package server

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	_ "net/http/pprof"
	"time"

	"baiyecha/ipvs-manager/server/login_handler"
	"baiyecha/ipvs-manager/server/raft_handler"
	"baiyecha/ipvs-manager/server/store_handler"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// srv struct handling server
type srv struct {
	listenAddress string
	raft          *raft.Raft
	echo          *echo.Echo
}

// Start start the server
func (s srv) Start() error {
	return s.echo.StartServer(&http.Server{
		Addr:         s.listenAddress,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

type Ipvs struct {
	VIP       string     `json:"vip"`
	Backends  []*Backend `json:"backends"`
	Protocol  string     `json:"protocol"`
	SchedName string     `json:"sched_name"`
}

type Backend struct {
	Addr   string `json:"addr"`
	Weight int    `json:"weight"`
	Status int    `json:"status"` // ipvs后端的健康状态，1为不健康，0为健康
}

type IpvsList struct {
	List []*Ipvs `json:"list"`
	Json string  `json:"-"`
}

//go:embed assets/index.html
var index string

//go:embed assets/table.html
var table string

//go:embed assets/jquery.min.js
var jquery string

// New return new server
func NewHttp(listenAddr string, raftHttpListenAddr string, badgerDB *badger.DB, r *raft.Raft, clusterAddress []string) {
	go newWebHttp(listenAddr, badgerDB, r, clusterAddress)
	go newRaftHttp(raftHttpListenAddr, badgerDB, r, clusterAddress)
	signal := make(chan int)
	<-signal
}

func newWebHttp(listenAddr string, badgerDB *badger.DB, r *raft.Raft, clusterAddress []string) {
	e := echo.New()
	t := template.Must(template.New("index.html").Parse(index))
	t = template.Must(t.New("jquery").Parse(jquery))

	renderer := &TemplateRenderer{
		templates: template.Must(t.New("table.html").Parse(table)),
	}

	e.HideBanner = true
	e.HidePort = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	e.Renderer = renderer
	// 登陆
	e.POST("/login", login_handler.Login)
	// 登出
	e.GET("/logout", login_handler.Logout)
	// 登陆表单页面
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})
	storeHandler := store_handler.New(r, badgerDB, clusterAddress)

	// table页面
	e.GET("/table", storeHandler.Table)

	// 更新ipvs
	e.POST("/update", storeHandler.Update)
	fmt.Println("web server start listen on ", listenAddr)
	s := &srv{
		listenAddress: listenAddr,
		echo:          e,
		raft:          r,
	}
	s.Start()
}

func newRaftHttp(listenAddr string, badgerDB *badger.DB, r *raft.Raft, clusterAddress []string) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	// Raft server
	raftHandler := raft_handler.New(r)
	e.POST("/raft/join", raftHandler.JoinRaftHandler)
	e.POST("/raft/remove", raftHandler.RemoveRaftHandler)
	e.GET("/raft/stats", raftHandler.StatsRaftHandler)

	// Store server
	storeHandler := store_handler.New(r, badgerDB, clusterAddress)
	e.POST("/store", storeHandler.Store)
	e.GET("/store/:key", storeHandler.Get)
	e.DELETE("/store/:key", storeHandler.Delete)

	fmt.Println("raft http server start listen on ", listenAddr)
	s := &srv{
		listenAddress: listenAddr,
		echo:          e,
		raft:          r,
	}
	s.Start()
}
