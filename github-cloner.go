package main

import "fmt"
import "github.com/google/go-github/github"
import "code.google.com/p/goauth2/oauth"
import "os"
import "os/exec"
import "os/user"
import "strings"
import "flag"
import "bufio"

func expandTilde(path string) string {
	if path[:2] == "~/" {
		u,_ := user.Current()
		return strings.Replace(path,"~",u.HomeDir,1)
	}
	return path
}

func gitRepoExists(path string, name string) bool {
	_, err := os.Stat(path+"/"+name)
	return !os.IsNotExist(err)
}

func gitClone(path string, giturl string) {
	cmd := exec.Command("git","clone",giturl)
	cmd.Dir = path
	cmd.Run()
}

func gitUpdate(path string,repo string) {
	cmd := exec.Command("git","pull")
	cmd.Dir = path + "/" + repo
	cmd.Run()
}

func gitSetUpstream(path string,repo string,upstreamurl string) {
	cmd := exec.Command("git","remote","add","upstream",upstreamurl)
	cmd.Dir = path + "/" + repo
	cmd.Run()
}

func main() {
	var tokenPath = flag.String("token","./github-token","path to file containing your github API token")
	var update = flag.Bool("update",false,"run updates on all repos")
	flag.Parse()
	people := flag.Args()
	tokf,_ := os.Open(*tokenPath)
	scan := bufio.NewScanner(tokf)
	token := ""
	for scan.Scan() {
		token += strings.TrimSpace(scan.Text())
	}
	t := &oauth.Transport{
		Token : &oauth.Token{AccessToken: token},
	}
	client := github.NewClient(t.Client())
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100,},
	}
	for _,person := range people {
		path := expandTilde("~/Repos/"+person)
		os.MkdirAll(path,0755)
		repos,_,_ := client.Repositories.List(person,opt)
		for _,repo := range repos {
			fmt.Printf("REPO %v:",*repo.Name)
			if gitRepoExists(path,*repo.Name) {
				if *update {
					fmt.Printf("Updating %v ...",*repo.FullName)
					gitUpdate(path,*repo.Name)
					fmt.Printf("done\n")
				} else {
					fmt.Printf("Repository %v already cloned\n",*repo.FullName)
				}
				continue
			}
			fmt.Printf("Cloning repository %v...",*repo.FullName)
			gitClone(path,*repo.SSHURL)
			if (*repo.Fork) {
				repo,_,_ := client.Repositories.Get(person,*repo.Name)
				parent := *repo.Parent
				gitSetUpstream(path,*repo.Name,*parent.CloneURL)
			}
			fmt.Printf(" done.\n")
		}
	}
}
