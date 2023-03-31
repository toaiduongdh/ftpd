package web

import (
	"time"

	"gitea.com/lunny/tango"
	"gitea.com/tango/binding"
	"gitea.com/tango/flash"
	"gitea.com/tango/renders"
	"gitea.com/tango/session"
	"gitea.com/tango/xsrf"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/toaiduongdh/ftpd/modules/public"
	"github.com/toaiduongdh/ftpd/modules/templates"
	"goftp.io/server/v2"
)

const (
	USER_MODULE = iota + 1
	GROUP_MODULE
	PERM_MODULE
	CHGPASS_MODULE
)

var (
	DB        UserDB
	Perm      server.Perm
	Driver    server.Driver
	adminUser string
)

type author interface {
	AskLogin() bool
	IsLogin() bool
	LoginUserId() string
}

func auth() tango.HandlerFunc {
	return func(ctx *tango.Context) {
		if a, ok := ctx.Action().(author); ok {
			if a.AskLogin() {
				if !a.IsLogin() {
					ctx.Redirect("/login")
					return
				}
			}
		}
		ctx.Next()
	}
}

const (
	timeout = time.Minute * 20
)

func Web(listen, static, templatesDir, admin, pass string, tls bool, certFile, keyFile string) {
	_, err := DB.GetUser(admin)
	if err != nil {
		if err == leveldb.ErrNotFound {
			err = DB.AddUser(admin, pass)
		}
	}
	if err != nil {
		panic(err)
	}
	adminUser = admin

	t := tango.Classic()
	sess := session.New(session.Options{
		MaxAge: timeout,
	})
	t.Use(
		public.Static(static),
		renders.New(renders.Options{
			Reload:     true, // if reload when template is changed
			Directory:  templatesDir,
			FileSystem: templates.FileSystem(templatesDir),
		}),
		sess,
		auth(),
		binding.Bind(),
		xsrf.New(timeout),
		flash.Flashes(sess),
	)

	t.Get("/", new(MainAction))
	t.Any("/login", new(LoginAction))
	t.Get("/logout", new(LogoutAction))
	t.Get("/down", new(DownAction))
	t.Group("/user", func(g *tango.Group) {
		g.Get("/", new(UserAction))
		g.Any("/chgpass", new(ChgPassAction))
		g.Any("/add", new(UserAddAction))
		g.Any("/edit", new(UserEditAction))
		g.Any("/del", new(UserDelAction))
	})

	t.Group("/group", func(g *tango.Group) {
		g.Get("/", new(GroupAction))
		g.Get("/add", new(GroupAddAction))
		g.Get("/edit", new(GroupEditAction))
		g.Get("/del", new(GroupDelAction))
	})
	t.Group("/perm", func(g *tango.Group) {
		g.Get("/", new(PermAction))
		g.Any("/add", new(PermAddAction))
		g.Any("/edit", new(PermEditAction))
		g.Any("/del", new(PermDelAction))
		g.Any("/updateOwner", new(PermUpdateOwner))
		g.Any("/updateGroup", new(PermUpdateGroup))
		g.Any("/updatePerm", new(PermUpdatePerm))
	})

	if tls {
		t.RunTLS(certFile, keyFile, listen)
		return
	}

	t.Run(listen)
}
