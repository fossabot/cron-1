package cron

import (
	"context"
	"fmt"
	"os"

	"github.com/icco/cron/goodreads"
	"github.com/icco/cron/pinboard"
	"github.com/icco/cron/spider"
	"github.com/icco/cron/tweets"
	"github.com/icco/cron/updater"
	"go.opencensus.io/tag"
)

// Act takes a job and calls a sub project to do work.
func Act(octx context.Context, job string) error {
	jobKey, err := tag.NewKey("natwelch.com/keys/job")
	if err != nil {
		log.WithError(err).Warn("could not create oc tag")
	}
	ctx, err := tag.New(octx,
		tag.Upsert(jobKey, job),
	)
	if err != nil {
		log.WithError(err).Warn("could not add oc tag")
	}

	gqlToken := os.Getenv("GQL_TOKEN")
	if gqlToken == "" {
		return fmt.Errorf("GQL_TOKEN is unset")
	}

	twitterAuth := &tweets.TwitterAuth{
		ConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		AccessToken:    os.Getenv("TWITTER_ACCESS_TOKEN"),
		AccessSecret:   os.Getenv("TWITTER_ACCESS_SECRET"),
	}

	pinboardToken := os.Getenv("PINBOARD_TOKEN")
	if pinboardToken == "" {
		return fmt.Errorf("PINBOARD_TOKEN is unset")
	}

	goodreadsToken := os.Getenv("GOODREADS_TOKEN")
	if goodreadsToken == "" {
		return fmt.Errorf("GOODREADS_TOKEN is unset")
	}

	switch job {
	case "minute":
		log.Info("> heartbeat")
	case "update-deployments":
		updater.UpdateWorkspaces(ctx, &updater.Config{Log: log})
	case "spider":
		spider.Crawl(ctx, &spider.Config{Log: log, URL: "https://writing.natwelch.com/"})
	case "user-tweets":
		t := tweets.Twitter{
			TwitterAuth:  twitterAuth,
			Log:          log,
			GraphQLToken: gqlToken,
		}
		err := t.SaveUserTweets(ctx)
		if err != nil {
			return err
		}
	case "pinboard":
		p := &pinboard.Pinboard{
			Token:        pinboardToken,
			Log:          log,
			GraphQLToken: gqlToken,
		}
		err := p.UpdatePins(ctx)
		if err != nil {
			return err
		}
	case "random-tweets":
		t := &tweets.Twitter{
			TwitterAuth:  twitterAuth,
			Log:          log,
			GraphQLToken: gqlToken,
		}
		err := t.CacheRandomTweets(ctx)
		if err != nil {
			return err
		}
	case "goodreads":
		g := &goodreads.Goodreads{
			Log:          log,
			Token:        goodreadsToken,
			GraphQLToken: gqlToken,
		}
		err := g.UpsertBooks(ctx)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown job type: %s", job)
	}

	return nil
}
