package gitlab

import (
	"fmt"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/sync/errgroup"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type issueFetcher struct {
}

func (f *issueFetcher) fetchPath(path string, client *gitlab.Client, isDebugLogging bool) (*Page, error) {
	re := regexp.MustCompile("^([^/]+)/([^/]+)/issues/(\\d+)")
	matched := re.FindStringSubmatch(path)

	if matched == nil {
		return nil, nil
	}

	projectName := matched[1] + "/" + matched[2]

	var eg errgroup.Group

	var issue *gitlab.Issue
	description := ""
	authorName := ""
	authorAvatarURL := ""
	var footerTime *time.Time

	eg.Go(func() error {
		issueID, _ := strconv.Atoi(matched[3])
		_issue, _, err := client.Issues.GetIssue(projectName, issueID)

		if err != nil {
			return err
		}

		issue = _issue

		if isDebugLogging {
			fmt.Printf("[DEBUG] issueFetcher: issue=%+v\n", issue)
		}

		description = issue.Description
		authorName = issue.Author.Name
		authorAvatarURL = issue.Author.AvatarURL
		footerTime = issue.CreatedAt

		re2 := regexp.MustCompile("#note_(\\d+)$")
		matched2 := re2.FindStringSubmatch(path)

		if matched2 != nil {
			noteID, _ := strconv.Atoi(matched2[1])
			note, _, err := client.Notes.GetIssueNote(projectName, issueID, noteID)

			if err != nil {
				return err
			}

			if isDebugLogging {
				fmt.Printf("[DEBUG] issueFetcher: note=%+v\n", note)
			}

			description = note.Body
			authorName = note.Author.Name
			authorAvatarURL = note.Author.AvatarURL
			footerTime = note.CreatedAt
		}

		return nil
	})

	var project *gitlab.Project
	eg.Go(func() error {
		_project, _, err := client.Projects.GetProject(projectName, nil)

		if err != nil {
			return err
		}

		project = _project

		if isDebugLogging {
			fmt.Printf("[DEBUG] issueFetcher: project=%+v\n", project)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	page := &Page{
		Title:                  strings.Join([]string{issue.Title, "Issues", project.NameWithNamespace, "GitLab"}, titleSeparator),
		Description:            description,
		AuthorName:             authorName,
		AuthorAvatarURL:        authorAvatarURL,
		AvatarURL:              project.AvatarURL,
		CanTruncateDescription: true,
		FooterTitle:            project.PathWithNamespace,
		FooterURL:              project.WebURL,
		FooterTime:             footerTime,
	}

	return page, nil
}
