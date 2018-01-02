package service

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/dearcode/crab/http/server"
	"github.com/juju/errors"
	"github.com/zssky/log"

	"github.com/dearcode/sapper/service/debug"
)

//RequestHeader 默认请求头.
type RequestHeader struct {
	Session string
	Request http.Request
}

//ResponseHeader 默认返回头.
type ResponseHeader struct {
	Status  int
	Message string `json:",omitempty"`
}

//Service 一个服务对象.
type Service struct {
	doc       document
	docView   docView
	router    router
	keepalive *keepalive
}

var (
	host        = flag.String("h", ":8080", "listen address.")
	version     = flag.Bool("v", false, "version info.")
	logLevel    = flag.String("logLevel", "debug", "log level: fatal, error, warning, debug, info.")
	logFile     = flag.String("logFile", "", "log file name.")
	maxWaitTime = time.Hour * 24 * 7
)

//New 返回service对象.
func New() *Service {
	return &Service{
		doc:       newDocument(),
		router:    newRouter(),
		keepalive: newKeepalive(),
	}
}

//Init 解析flag参数, 初始化基本信息.
func (s *Service) Init() {
	flag.Parse()

	if *version {
		debug.Print()
		os.Exit(0)
	}

	if *logFile != "" {
		log.SetHighlighting(false)
		log.SetRotateByDay()
		log.SetOutputByName(*logFile)
	}

	log.SetLevelByString(*logLevel)

	server.RegisterPrefix(&debug.Debug{}, "/debug/pprof/")
	server.RegisterPrefix(&debug.Version{}, "/debug/version/")
	server.RegisterPrefix(&s.doc, "/document/")

}

//Register 注册接口.
func (s *Service) Register(obj interface{}) error {
	t := reflect.TypeOf(obj)
	path := t.PkgPath()
	name := t.Name()

	//不能脱壳，脱壳后取不到method.
	if t.Kind() == reflect.Ptr {
		path = t.Elem().PkgPath()
		name = t.Elem().Name()
	}

	if idx := strings.Index(path, "/"); idx > 0 {
		path = path[strings.Index(path, "/"):]
	} else {
		path = "/" + path
	}

	url := path + "/" + name

	//log.Debugf("%s url:%v, method:%d", name, url, t.NumMethod())

	for _, k := range []string{"Get", "Post", "Put", "Delete"} {
		if m, ok := t.MethodByName(k); ok {
			if m.Type.NumIn() == 3 {
				if err := server.RegisterHandler(s.handler, strings.ToUpper(k), url); err != nil {
					log.Errorf("RegisterPrefix %v error:%v", url, err)
					return err
				}
				s.router.add(strings.ToUpper(k), url, m)
				s.doc.add(name, url, m)
			}
		}
	}

	return nil
}

//Start 如果本地测试只需要localMode为true，则不会去etcd中注册.
func (s *Service) Start(localMode bool) {
	s.docView = newDocView(s.doc)
	server.RegisterPath(&s.docView, "/doc/")
	//第一步，启动服务
	ln, err := server.Start(*host)
	if err != nil {
		log.Errorf("%v", errors.ErrorStack(err))
		panic(err)
	}

	//第二步，注册到接口平台API接口队列中.
	if !localMode {
		if err := s.keepalive.start(ln, s.doc); err != nil {
			log.Errorf("apiRegister error:%v", errors.ErrorStack(err))
			panic(err)
		}
	}

	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGUSR1)

	sig := <-shutdown
	s.keepalive.stop()
	log.Warningf("%v recv signal %v, close:%v", os.Getpid(), sig, ln.Close())

	log.Warningf("%v wait timeout:%v.", os.Getpid(), maxWaitTime)
	<-time.After(maxWaitTime)
	log.Warningf("%v exit", os.Getpid())
}
