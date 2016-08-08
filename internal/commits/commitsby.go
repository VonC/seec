package commits

import (
	"fmt"
	"time"

	"seec2/internal/gh"

	"github.com/google/go-github/github"
)

type CommitsByAuthor struct {
	author string
	cbd    []*CommitsByDate
}
type CommitsByAuthors map[string]*CommitsByAuthor

// Because of seec 709cd912d4663af87903d3d278a3bab9d4d84153
type CommitsByDate struct {
	date    *time.Time
	commits []*github.Commit
}

func (cba *CommitsByAuthor) String() string {
	res := ""
	first := true
	for i, acbd := range cba.cbd {
		if !first {
			res = res + ", "
		}
		first = false
		if i == len(cba.cbd)-1 && i > 0 {
			res = res + "and "
		}
		res = res + acbd.String()
	}
	return fmt.Sprintf("%s=>%s", cba.author, res)
}

func (cbd *CommitsByDate) String() string {
	res := ""
	first := true
	for i, commit := range cbd.commits {
		if !first {
			res = res + ", "
		}
		first = false
		if i == len(cbd.commits)-1 && i > 0 {
			res = res + "and "
		}
		res = res + (*commit.SHA)[:7]
	}
	return fmt.Sprintf("%s (%s)", res, cbd.date.Format("02 Jan 2006"))
}

func NewCommitsByAuthor(authorname string) *CommitsByAuthor {
	return &CommitsByAuthor{authorname, []*CommitsByDate{}}
}

func (cbas CommitsByAuthors) Add(somecbas CommitsByAuthors) {

}

func (cba *CommitsByAuthor) AddCommit(commit *gh.Commit) {

}
