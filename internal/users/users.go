package users

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"seec2/internal/gh"
	"strings"

	"github.com/google/go-github/github"
)

type User struct {
	Login string
	Name  string
}

func AuthorNameAndLogin(c *gh.Commit) *User {
	clogin := login(*c.Author.Email, *c.Author.Name, *c.SHA)
	cname := *c.Author.Name
	return &User{Login: clogin, Name: cname}
}
func CommitterNameAndLogin(c *gh.Commit) *User {
	clogin := login(*c.Committer.Email, *c.Committer.Name, *c.SHA)
	cname := *c.Committer.Name
	return &User{Login: clogin, Name: cname}
}

func Login(email string, name string) string {
	return login(email, name, "")
}

func login(email string, name string, SHA string) string {
	// fmt.Printf("email='%s', name='%s'\n", email, name)
	if login := cacheLogins(email, name); login != "" {
		return login
	}
	if login := scrapPage(SHA); login != "" {
		return addToCacheLogins(email, name, login)
	}
	opts := &github.SearchOptions{Order: "desc"}
	var res *github.UsersSearchResult
	var err error
	if email != "" {
		res, _, err = gh.Client.Search.Users(email, opts)
		if err != nil {
			fmt.Printf("Unable to search user '%s': err '%v'", email, err)
			gh.GHex.Exit(1)
		}
	}
	nameNoDash := name
	if res == nil || *res.Total == 0 {
		nameNoDash = strings.Replace(name, "-", " ", -1)
		res, _, err = gh.Client.Search.Users(nameNoDash, opts)
		if err != nil {
			fmt.Printf("Unable to search user '%s': err '%v'", nameNoDash, err)
			gh.GHex.Exit(1)
		}
	}
	if res == nil || *res.Total == 0 {
		var resIssues *github.IssuesSearchResult
		issueSearch := fmt.Sprintf(`"Signed-off-by: %s <%s>"`, nameNoDash, email)
		resIssues, _, err = gh.Client.Search.Issues(issueSearch, opts)
		if err != nil {
			fmt.Printf("Unable to search issue '%s': err '%v'", issueSearch, err)
			gh.GHex.Exit(1)
		}
		if resIssues == nil || *resIssues.Total == 0 {
			return ""
		}
		issue := resIssues.Issues[0]
		return addToCacheLogins(email, name, *issue.User.Login)
	}
	if res == nil || *res.Total == 0 {
		return ""
	}
	return addToCacheLogins(email, name, *res.Users[0].Login)
}

var cacheLoginsRE = regexp.MustCompile(`^(.*)#~#(.*?)\^\((.*?)\)\^(.*?)$`)

func cacheLogins(email string, name string) string {
	login := ""
	fi, err := os.OpenFile("logins.txt", os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	var part []byte
	var prefix bool
	reader := bufio.NewReader(fi)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			line := buffer.String()
			re := cacheLoginsRE.FindAllStringSubmatch(line, -1)
			if len(re) == 1 {
				femail := re[0][1]
				fname := re[0][2]
				flogin := re[0][4]
				if femail == email && fname == name {
					// fmt.Printf("femail='%s', fname='%s' => login '%s'\n", femail, fname, flogin)
					login = flogin
					break
				}
			}
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		panic(err)
	}
	// fmt.Printf("cacheLogins='%s'\n", login)
	return login
}

var contributorRE = regexp.MustCompile(`.*?<a\s+href="[^"]+"\s+rel="contributor"\s*?>([^<]+)</a>`)

func scrapPage(SHA string) string {
	if SHA == "" {
		return ""
	}
	response, err := http.Get("https://github.com/git/git/commit/" + SHA)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		lines := strings.Split(string(contents), `\n`)
		for _, line := range lines {
			if strings.Contains(line, "contributor") {
				r := contributorRE.FindAllStringSubmatch(line, -1)
				if len(r) > 0 {
					if len(r[0]) == 2 {
						// fmt.Printf("r[0]='%+v' => res='%s'\n", r[0], r[0][1])
						return r[0][1]
					}
				}
				fmt.Println("line='%+v'\n", r)
			}
		}
	}
	return ""
}
func addToCacheLogins(email string, name string, login string) string {
	fmt.Printf("addToCacheLogins '%s'\n", login)
	fi, err := os.OpenFile("logins.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := fi.WriteString(fmt.Sprintf("%s#~#%s^()^%s\n", email, name, login)); err != nil {
		panic(err)
	}
	return login

}
