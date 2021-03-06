package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/drone/drone-plugin-go/plugin"
)

type Itamae struct {
	Recipes           []string `json:["nginx.rb", "jenkins.rb"]`
	ItamaeTargetImage string   `json:"itamae_target_image"`
	Storage           string   `json:"storage_driver"`
	Registry          string   `json:"registry"`
	Username          string   `json:"username"`
	Password          string   `json:"password"`
	Email             string   `json:"email"`
	Auth              string   `json:"auth"`
	Repo              string   `json:"repo"`
	Tag               string   `json:"tag"`
	File              string   `json:"file"`
}

func main() {
	clone := plugin.Clone{}
	vargs := Itamae{}

	plugin.Param("clone", &clone)
	plugin.Param("vargs", &vargs)
	if err := plugin.Parse(); err != nil {
		println(err.Error())
		os.Exit(1)
	}

	// Set the storage driver
	if len(vargs.Storage) == 0 {
		vargs.Storage = "aufs"
	}

	stop := func() {
		cmd := exec.Command("start-stop-daemon", "--stop", "--pidfile", "/var/run/docker.pid")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		cmd.Run()
	}
	defer stop()

	// Starts the Docker daemon
	go func() {
		cmd := exec.Command("/bin/bash", "/bin/wrapdocker")
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		cmd.Run()

		cmd = exec.Command("/etc/init.d/docker", "start")
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		trace(cmd)
		cmd.Run()
	}()

	// Sleep for a few seconds
	time.Sleep(5 * time.Second)

	// Set the Registry value
	if len(vargs.Registry) == 0 {
		vargs.Registry = "https://index.docker.io/v1/"
	}
	// Set the Dockerfile path
	if len(vargs.File) == 0 {
		vargs.File = "."
	}
	// Set the Tag value
	switch vargs.Tag {
	case "$DRONE_BRANCH":
		vargs.Tag = clone.Branch
	case "$DRONE_COMMIT":
		vargs.Tag = clone.Sha
	case "":
		vargs.Tag = "latest"
	}
	vargs.Repo = fmt.Sprintf("%s:%s", vargs.Repo, vargs.Tag)

	// Login to Docker
	if len(vargs.Username) > 0 {
		cmd := exec.Command("docker", "login", "-u", vargs.Username, "-p", vargs.Password, "-e", vargs.Email, vargs.Registry)
		cmd.Dir = clone.Dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			stop()
			os.Exit(1)
		}
	}

	// Docker environment info
	cmd := exec.Command("docker", "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()
	cmd = exec.Command("docker", "info")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()

	// Build the container
	//TODO custom json
	image_id := ""
	for index, recipe := range vargs.Recipes {
		from_image := ""
		if index == 0 {
			from_image = vargs.ItamaeTargetImage
		} else {
			from_image = image_id
		}
		cmd = exec.Command("itamae", "docker", "--image="+from_image, "--node-json=attribute.json", recipe)
		cmd.Dir = clone.Dir
		trace(cmd)
		out, err := cmd.Output()
		println(string(out))
		if err != nil {
			stop()
			os.Exit(1)
		}
		stdoutArray := strings.Split(string(out), "\n")
		rs, _ := regexp.Compile("INFO : Image created: ([(a-z0-9]+)")
		ri, _ := regexp.Compile("([(a-z0-9]+)$")
		for _, stdout := range stdoutArray {
			if rs.MatchString(stdout) {
				image_id = ri.FindString(stdout)
			}
		}
	}

	cmd = exec.Command("docker", "tag", image_id, vargs.Repo)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	cmd.Run()

	// Push the container
	cmd = exec.Command("docker", "push", vargs.Repo)
	cmd.Dir = clone.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err := cmd.Run()
	if err != nil {
		stop()
		os.Exit(1)
	}
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}

// authorize is a helper function that authorizes the Docker client
// by manually creating the Docker authentication file.
func authorize(d *Itamae) error {
	var path = "/root/.dockercfg" // TODO should probably use user.Home() for good measure
	var data = fmt.Sprintf(dockerconf, d.Registry, d.Auth, d.Email)
	return ioutil.WriteFile(path, []byte(data), 0644)
}

var dockerconf = `
{
	"%s": {
		"auth": "%s",
		"email": "%s"
	}
}
`
