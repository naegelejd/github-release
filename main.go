package main

import (
    "os"
    "fmt"
    "log"
    "path/filepath"
    "flag"
    "github.com/google/go-github/github"
)

var logger *log.Logger

func init () {
    prog := filepath.Base(os.Args[0])
    /* flags := log.Ldate | log.Ltime | log.Lshortfile */
    flags := log.Lshortfile
    logger = log.New(os.Stderr, prog + ": ", flags)
}

func uploadFiles(c *github.Client, owner, repo string,
        releaseID int, filenames []string) error {
    for _,fname := range filenames {
        basename := filepath.Base(fname)
        f, err := os.Open(fname)
        if err != nil {
            return err
        }
        defer f.Close()

        opts := &github.UploadOptions{Name: basename}
        asset, _, err := c.Repositories.UploadReleaseAsset(owner, repo,
                releaseID, opts, f)
        if err != nil {
            return err
        }
        fmt.Printf("Uploaded %s to %s\n", basename, *asset.BrowserDownloadUrl)
    }
    return nil
}

func main () {
    var accessToken string
    var owner, repo, tag, target, name, body string
    var draft, pre bool
    const required = "<required>"
    const sameastag = "<same as tag>"

    flag.StringVar(&owner, "owner", required, "Github repository owner")
    flag.StringVar(&repo, "repo", required, "Github repository name")
    flag.StringVar(&tag, "tag", required, "tag name")
    flag.StringVar(&target, "target", "master", "target commit/branch")
    flag.StringVar(&name, "name", sameastag, "name of release")
    flag.StringVar(&body, "body", "", "release information")
    flag.BoolVar(&draft, "draft", false, "is release a draft?")
    flag.BoolVar(&pre, "pre", false, "is release a pre-release?")
    flag.StringVar(&accessToken, "token", "", "github personal access token")
    flag.Parse()

    if owner == required {
        logger.Fatal("Missing -owner=<owner>")
    }
    if repo == required {
        logger.Fatal("Missing -repo=<repo>")
    }
    if tag == required {
        logger.Fatal("Missing -tag=<tag>")
    }
    if name == sameastag {
        name = tag
    }
    if flag.NArg() == 0 {
        logger.Fatal("Must specify at least 1 release file to upload")
    }

    t := &Transport{Username: accessToken, Password: ""}
    c := github.NewClient(t.Client())

    release := &github.RepositoryRelease{
        TagName: github.String(tag),
        TargetCommitish: github.String(target),
        Name: github.String(name),
        Body: github.String(body),
        Draft: github.Bool(draft),
        Prerelease: github.Bool(pre),
    }

    release, _, err := c.Repositories.CreateRelease(owner, repo, release)
    if err != nil {
        logger.Fatal(err)
    }

    if err := uploadFiles(c, owner, repo, *release.ID, flag.Args()); err != nil {
        logger.Fatal(err)
    }
}
