package namingx

import (
	"context"
	"flag"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sanity-io/litter"

	"github.com/tsingson/discovery/discovery"

	"github.com/tsingson/discovery-etcdv3/conf"

	xhttp "github.com/tsingson/discovery-etcdv3/lib/http"
	xtime "github.com/tsingson/discovery-etcdv3/lib/time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	flag.Parse()
	// go mockDiscoverySvr()
	time.Sleep(time.Second)
	os.Exit(m.Run())
}

func mockDiscoverySvr() {
	c := &conf.Config{
		Env: &conf.Env{
			Region:    "dev",
			Zone:      "dev",
			DeployEnv: "dev",
			Host:      "test_server",
		},
		Nodes: []string{"127.0.0.1:7171"},
		HTTPServer: &conf.ServerConfig{
			Addr: "127.0.0.1:7171",
		},
		HTTPClient: &xhttp.ClientConfig{
			Dial:      xtime.Duration(time.Second),
			KeepAlive: xtime.Duration(time.Second * 30),
		},
	}
	_ = c.Fix()
	dis, _ := discovery.New(c)
	http.Init(c, dis)
}

func TestDiscovery(t *testing.T) {
	conf := &NamingConfig{
		Nodes:  []string{"127.0.0.1:7171"},
		Region: "china",
		Zone:   "gd",
		Env:    "dev",
		Host:   "discovery",
	}

	dis := NewClient(conf)
	println("new")
	appid := "test1"
	Convey("test discovery register", t, func() {
		instance := &Instance{
			Region:   "china",
			Zone:     "gd",
			Env:      "dev",
			AppID:    appid,
			Hostname: "test-host",
		}
		_, err := dis.Register(instance)
		So(err, ShouldBeNil)
		dis.node.Store([]string{"127.0.0.1:7172"})
		instance.AppID = "test2"
		// instance.Metadata = map[string]string{"meta": "meta"}
		_, err = dis.Register(instance)
		So(err, ShouldNotBeNil)
		_ = dis.renew(context.TODO(), instance)
		dis.node.Store([]string{"127.0.0.1:7171"})
		_ = dis.renew(context.TODO(), instance)
		Convey("test discovery set", func() {
			rs := dis.Build(appid)
			inSet := &Instance{
				Region:   "china",
				Zone:     "gd",
				Env:      "dev",
				AppID:    appid,
				Hostname: "test-host",
				Addrs: []string{
					"grpc://127.0.0.1:8080",
				},
				Metadata: map[string]string{
					"dev":    "1",
					"weight": "111",
					"color":  "blue",
				},
			}
			err = dis.Set(inSet)
			So(err, ShouldBeNil)
			ch := rs.Watch()
			<-ch
			ins, _ := rs.Fetch()
			litter.Dump(ins.Instances)
			So(ins.Instances["gd"][0].Metadata["weight"], ShouldResemble, "111")
		})
	})
	Convey("test discovery watch", t, func() {
		rsl := dis.Build(appid)
		ch := rsl.Watch()
		<-ch
		ins, ok := rsl.Fetch()
		So(ok, ShouldBeTrue)
		So(len(ins.Instances["gd"]), ShouldEqual, 1)
		So(ins.Instances["gd"][0].AppID, ShouldEqual, appid)
		instance2 := &Instance{
			Region:   "china",
			Zone:     "gd",
			Env:      "dev",
			AppID:    appid,
			Hostname: "test-host2",
		}
		err := addNewInstance(instance2)
		So(err, ShouldBeNil)
		// watch for next update
		<-ch
		ins, ok = rsl.Fetch()
		So(ok, ShouldBeTrue)
		So(len(ins.Instances["gd"]), ShouldEqual, 2)
		So(ins.Instances["gd"][0].AppID, ShouldEqual, appid)
		rsl.Close()
		conf.Nodes = []string{"127.0.0.1:7171"}
		dis.Reload(conf)
		So(dis.Scheme(), ShouldEqual, "discovery")
		dis.Close()
	})

}

func addNewInstance(ins *Instance) error {
	cli := xhttp.NewClient(&xhttp.ClientConfig{
		Dial:      xtime.Duration(time.Second),
		KeepAlive: xtime.Duration(time.Second * 30),
	})
	params := url.Values{}
	params.Set("env", ins.Env)
	params.Set("zone", ins.Zone)
	params.Set("hostname", ins.Hostname)
	params.Set("appid", ins.AppID)
	params.Set("addrs", strings.Join(ins.Addrs, ","))
	params.Set("version", ins.Version)
	params.Set("status", "1")
	params.Set("latest_timestamp", strconv.FormatInt(time.Now().UnixNano(), 10))
	res := new(struct {
		Code int `json:"code"`
	})
	return cli.Post(context.TODO(), "http://127.0.0.1:7171/discovery/register", "", params, &res)
}

func TestUseScheduler(t *testing.T) {
	newIns := func() *InstancesInfo {
		insInfo := &InstancesInfo{}
		insInfo.Instances = make(map[string][]*Instance)
		insInfo.Instances["sh001"] = []*Instance{
			{Zone: "sh001", Metadata: map[string]string{
				"weight": "10",
			}},
			{Zone: "sh001", Metadata: map[string]string{
				"weight": "10",
			}},
		}
		insInfo.Instances["sh002"] = []*Instance{
			{Zone: "sh002", Metadata: map[string]string{
				"weight": "5",
			}},
			{Zone: "sh002", Metadata: map[string]string{
				"weight": "2",
			}},
		}
		insInfo.Instances["sh003"] = []*Instance{
			{Zone: "sh003", Metadata: map[string]string{
				"weight": "5",
			}},
			{Zone: "sh003", Metadata: map[string]string{
				"weight": "3",
			}},
		}
		insInfo.Scheduler = []Zone{
			{
				Src: "sh001",
				Dst: map[string]int64{
					"sh001": 2,
					"sh002": 1,
				},
			},
			{
				Src: "sh002",
				Dst: map[string]int64{
					"sh001": 1,
					"sh002": 2,
				},
			},
		}
		return insInfo
	}
	Convey("use scheduler for sh001", t, func() {
		insInfo := newIns()
		inss := insInfo.UseScheduler("sh001")
		So(len(inss), ShouldEqual, 4)
		logInss(t, "sh001", inss)
	})
	Convey("use scheduler for sh002", t, func() {
		insInfo := newIns()
		inss := insInfo.UseScheduler("sh002")
		So(len(inss), ShouldEqual, 4)
		logInss(t, "sh002", inss)
	})
	Convey("use scheduler for sh003 without scheduler", t, func() {
		insInfo := newIns()
		inss := insInfo.UseScheduler("sh003")
		So(len(inss), ShouldEqual, 2)
		logInss(t, "sh003", inss)
	})
	Convey("zone not exit", t, func() {
		insInfo := newIns()
		inss := insInfo.UseScheduler("sh004")
		So(len(inss), ShouldEqual, 6)
		logInss(t, "sh004", inss)
	})
}

func logInss(t *testing.T, msg string, inss []*Instance) {
	t.Log("instance of", msg)
	for _, in := range inss {
		t.Logf("%+v", in)
	}
}
