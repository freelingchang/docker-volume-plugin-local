package main

// TODO: check that we're running with mount priv

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	pluginId = "mydriver"
)

var (
	socketDir     = "/run/docker/plugins/"
	socketAddress = filepath.Join(socketDir, strings.Join([]string{pluginId, ".sock"}, ""))
	defaultDir    = filepath.Join(volume.DefaultDockerRootDirectory, strings.Join([]string{"_", pluginId}, ""))
	root          = flag.String("root", defaultDir, "NFS volumes root directory")
	verbose       = true
)

type myDriver struct {
	root       string
	Name       string
	Root       string
	DataPath   string
	MountPoint string
	DataFile   string
	Size       int64
}

func NewMyvolume(root string, name string) myDriver {
	m := myDriver{}
	m.Name = name
	m.root = root
	m.Root = filepath.Join(root, "mnt")
	m.DataPath = filepath.Join(root, "metadata")
	m.MountPoint = filepath.Join(m.Root, name)
	m.DataFile = filepath.Join(m.DataPath, name)
	m.Size = 10737418240
	return m
}

func (g myDriver) Create(r volume.Request) volume.Response {
	fmt.Printf("Create %v\n", r)
	//fmt.Println(volume.DefaultDockerRootDirectory)
	m := NewMyvolume(g.root, r.Name)
	_, err := os.Stat(m.Root)
	if err != nil {
		//fmt.Println(m.Root)
		if err := os.MkdirAll(m.Root, os.ModePerm); err != nil {
			return volume.Response{Err: err.Error()}
		}
	}
	_, err = os.Stat(m.DataPath)
	if err != nil {
		//fmt.Println(m.DataPath)
		if err := os.MkdirAll(m.DataPath, os.ModePerm); err != nil {
			return volume.Response{Err: err.Error()}
		}
	}
	_, err = os.Stat(m.MountPoint)
	if err != nil {
		//fmt.Println(m.DataPath)
		if err := os.MkdirAll(m.MountPoint, os.ModePerm); err != nil {
			return volume.Response{Err: err.Error()}
		}
	}
	filename := m.DataFile
	_, err = os.Stat(filename)
	if err == nil {
		log.Printf("file exist")
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Printf(err.Error())
		return volume.Response{Err: err.Error()}
	}
	if err := f.Truncate(m.Size); err != nil {
		log.Printf(err.Error())
		return volume.Response{Err: err.Error()}
	}
	cmd := exec.Command("mkfs.xfs", "-f", filename)
	_, err = cmd.CombinedOutput()
	if err != nil {
		//fmt.Println(err)
		log.Printf(err.Error())
		return volume.Response{Err: err.Error()}
	}
	return volume.Response{}
}

func (g myDriver) Remove(r volume.Request) volume.Response {
	fmt.Printf("Remove %v\n", r)
	return volume.Response{}
}

func (g myDriver) Path(r volume.Request) volume.Response {
	fmt.Printf("Path %v\n", r)
	m := NewMyvolume(g.root, r.Name)
	return volume.Response{Mountpoint: m.MountPoint}
}

func (g myDriver) Get(r volume.Request) volume.Response {
	fmt.Printf("Get %v\n", r)
	m := NewMyvolume(g.root, r.Name)
	f := m.exist(r.Name)
	if f == true {
		return volume.Response{Volume: &volume.Volume{Name: r.Name, Mountpoint: m.MountPoint}}
	}

	return volume.Response{Err: fmt.Sprintf("Unable to find volume mounted on %s", m)}
}

func (g myDriver) List(r volume.Request) volume.Response {
	fmt.Printf("List %v\n", r)
	var vols []*volume.Volume
	m := NewMyvolume(g.root, r.Name)
	files, err := ioutil.ReadDir(m.DataPath)
	if err != nil {
		return volume.Response{Err: err.Error()}
	}
	volumes := []string{}
	l := []string{}

	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			//fileinfo := os.Stat(file)
			l = append(volumes, file.Name())
		}
	}
	for _, v := range l {
		mm := NewMyvolume(g.root, v)
		vols = append(vols, &volume.Volume{Name: v, Mountpoint: mm.MountPoint})
	}
	return volume.Response{Volumes: vols}
}

func (g myDriver) Mount(r volume.MountRequest) volume.Response {
	m := NewMyvolume(g.root, r.Name)
	cmd := exec.Command("mount", m.DataFile, m.MountPoint)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err.Error())
		return volume.Response{Err: err.Error()}
	}
	fmt.Printf("Mount %s at %s\n", r.Name, m.MountPoint)
	return volume.Response{Mountpoint: m.MountPoint}
}

func (g myDriver) Unmount(r volume.UnmountRequest) volume.Response {
	m := NewMyvolume(g.root, r.Name)
	fmt.Printf("Unmount %s\n", m.MountPoint)
	cmd := exec.Command("umount", m.MountPoint)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err.Error())
		return volume.Response{Err: err.Error()}
	}
	return volume.Response{}
}

func (g myDriver) Capabilities(r volume.Request) volume.Response {
	fmt.Printf("Capabilities %v\n", r)
	return volume.Response{Capabilities: volume.Capability{Scope: "global"}}
}

func (g myDriver) exist(name string) bool {
	filename := filepath.Join(g.Root, name)
	_, err := os.Stat(filename)
	if err == nil {
		return true
	} else {
		return false
	}
}

func main() {
	d := myDriver{root: *root}
	h := volume.NewHandler(d)
	fmt.Printf("listening on %s\n", socketAddress)
	fmt.Println(h.ServeUnix(socketAddress, 0))
}
