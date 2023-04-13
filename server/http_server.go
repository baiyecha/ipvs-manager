package server

import (
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
	IpvsList []*Ipvs `json:"ipvs_list"`
	Json     string  `json:"-"`
}

// New return new server
func NewHttp(listenAddr string, badgerDB *badger.DB, r *raft.Raft, clusterAddress []string) *srv {
	e := echo.New()
	t := template.Must(template.New("index.html").Parse(`
<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>index</title>
</head>
<body>
<div>
<form action="/login" method="post">
	用户名: <input type="text" name="name"><br/>
	密&nbsp;&nbsp;&nbsp;码: <input type="password" name="password"><br/>
	<input type="submit">
</form>
</div>
</body>
</html>
`))
	renderer := &TemplateRenderer{
		templates: template.Must(t.New("table.html").Parse(`
<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
    <script src="https://apps.bdimg.com/libs/jquery/2.1.4/jquery.min.js"></script>
    <style>
        table, table tr th, table tr td {
            border: 1px solid #0094ff;
        }

        button {
            margin: 2px;
        }

        table {
            width: 800px;
            min-height: 25px;
            line-height: 25px;
            text-align: center;
            border-collapse: collapse;
            padding: 2px;
        }
    </style>
</head>
<script>
    let ipvsList = JSON.parse({{ .Json }})
    let operate = null
    let backends = 1
	let updateIndex = -1
    function add() {
        $("#form").show()
        operate = "add"
        $("#table").hide()
    }

    function cancel() {
        $("#form").hide()
        operate = null
		backends = 1
        updateIndex = -1
        $("#table").show()
        $("#backends").empty()
        $("#vip").val("")
        $("#protocol").val("")
        $("#sched_name").val("")
        $("#addr").val("")
        $("#weight").val("")
    }

    function deleteIpvs(id) {
        //console.log(id)
		ipvsList.splice(id, 1)
		$.ajax({
			url         : 'update',
			type        : 'POST',
			dataType    : 'json',
			contentType : 'application/json;charset=UTF-8',
			data        : JSON.stringify({ipvs_list:ipvsList}),
			complete    : function (re) {
				console.log("add re:",re)
				window.location.reload()
			}
		})
    }

    function update(index) {
        console.log(index)
        $("#form").show()
        operate = "update"
        $("#table").hide()
        updateIndex = index
        $("#vip").val(ipvsList[index].vip)
        $("#protocol").val(ipvsList[index].protocol)
        $("#sched_name").val(ipvsList[index].sched_name)
		ipvsList[index].backends.forEach((item,i)=>{
            if (i === 0) {
                $("#addr").val(item.addr)
                $("#weight").val(item.weight)
            } else {
                backends=i+1
                $("#backends").append("<div id=\"backend" + backends + "\">\n" +
                    "            <label>\n" +
                    "                Addr:\n" +
                    "                <input type=\"text\" name=\"addr\" value=\""+item.addr+"\">\n" +
                    "            </label>\n" +
                    "            <label>\n" +
                    "                Weight:\n" +
                    "                <input type=\"text\" name=\"weight\" value=\""+item.weight+"\">\n" +
                    "            </label>\n" +
                    "            <button onclick=\"deleteBackend('backend" + backends + "')\">-</button>\n" +
                    "        </div>")
            }
        })
    }

    function addBackend() {
        backends++
        $("#backends").append("<div id=\"backend" + backends + "\">\n" +
            "            <label>\n" +
            "                Addr:\n" +
            "                <input type=\"text\" name=\"addr\">\n" +
            "            </label>\n" +
            "            <label>\n" +
            "                Weight:\n" +
            "                <input type=\"text\" name=\"weight\">\n" +
            "            </label>\n" +
            "            <button onclick=\"deleteBackend('backend" + backends + "')\">-</button>\n" +
            "        </div>")
    }

    function deleteBackend(id) {
        $("#" + id).remove()
    }

    function addOrUpdate() {
        let values = $("#form").serializeArray()
        let ipvs = {
            backends: [],
        }
        values.forEach((item) => {
            if (item.name === "vip") {
                ipvs.vip = item.value
            } else if (item.name === "protocol") {
                ipvs.protocol = item.value
            } else if (item.name === "sched_name") {
                ipvs.sched_name = item.value
            } else if (item.name === "addr") {
                ipvs.backends.push({addr: item.value})
            } else if (item.name === "weight") {
                ipvs.backends[ipvs.backends.length-1].weight = Number(item.value)
            }
        })
        if (operate === "add") {
            ipvsList.push(ipvs)
        } else if (operate === "update") {
            ipvsList.splice(updateIndex, 1, ipvs)
        }
		$.ajax({
			url         : 'update',
			type        : 'POST',
			dataType    : 'json',
			contentType : 'application/json;charset=UTF-8',
			data        : JSON.stringify({ipvs_list:ipvsList}),
			complete    : function (re) {
				console.log("add re:",re)
				window.location.reload()
			}
		})

    }

    $(function () {
        console.log(ipvsList)
    })
</script>
<body>
<button type="button" id="add" onclick="add()">新增</button>
<table id="table">

    <thead>
    <tr>
        <th>VIP</th>
        <th>Protocol</th>
        <th>SchedName</th>
        <th>Addr</th>
        <th>Weight</th>
        <th>Status</th>
        <th>操作</th>
    </tr>
    </thead>
    <tbody>
	{{- range $index1, $ipvs := .IpvsList}}
    <tr>
        <td rowspan="{{ len $ipvs.Backends}}">{{ $ipvs.VIP }}</td>
		<td rowspan="{{ len $ipvs.Backends}}">{{ $ipvs.Protocol }}</td>
		<td rowspan="{{ len $ipvs.Backends}}">{{ $ipvs.SchedName }}</td>
		{{- range $index, $backend := $ipvs.Backends}}
			{{- if gt $index 0}}
		</tr><tr>
			{{- end}}
        <td>{{ $backend.Addr }}</td>
		<td>{{ $backend.Weight }}</td>
		<td>{{ $backend.Status }}</td>
			{{- if eq $index 0}}
        <td rowspan="{{ len $ipvs.Backends}}">
            <button type="button" onclick="update({{ $index1 }})">编辑</button>
            <button type="button" onclick="deleteIpvs({{ $index1 }})">删除</button>
        </td>
			{{- end}}
		{{- end}}
    </tr>
    {{- end}}
    </tbody>
</table>
<form id="form" style="display: none;" action="/update" method="post" onsubmit="return false">
    <label>
        VIP:
        <input id="vip" type="text" name="vip">
    </label><br/>
    <label>
        Protocol:
        <input id="protocol" type="text" name="protocol">
    </label><br/>
    <label>
        SchedName:
        <input id="sched_name" type="text" name="sched_name">
    </label><br/>
    Backends:
    <div>
        <label>
            Addr:
            <input id="addr" type="text" name="addr">
        </label>
        <label>
            Weight:
            <input id="weight" type="text" name="weight">
        </label>
        <button onclick="addBackend()">+</button>
    </div>
    <div id="backends">

    </div>
    <br/>
    <button onclick="addOrUpdate()">提交</button>
    <button onclick="cancel()">取消</button>
</form>
</body>
</html>
`)),
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

	// table页面
	e.GET("/table", storeHandler.Table)

	// 更新ipvs
	e.POST("/update", storeHandler.Update)
	return &srv{
		listenAddress: listenAddr,
		echo:          e,
		raft:          r,
	}
}
