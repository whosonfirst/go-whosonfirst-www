package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/whosonfirst/go-whosonfirst-crawl"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// used to pass custom -mime-type .foo=text/plain args that are
// set below using mime.AddExtensionType

type MimeTypes []string

func (m *MimeTypes) String() string {
	return strings.Join(*m, " ")
}

func (m *MimeTypes) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// all of this S3 stuff is cloned from https://github.com/thisisaaronland/go-iiif/blob/master/aws/s3.go
// and probably deserves to be moved in to a bespoke package some day... (20170131/thisisaaronland)

type S3Connection struct {
	service *s3.S3
	bucket  string
	prefix  string
}

type S3Config struct {
	Bucket      string
	Prefix      string
	Region      string
	Credentials string // see notes below
}

func NewS3Connection(s3cfg S3Config) (*S3Connection, error) {

	// https://docs.aws.amazon.com/sdk-for-go/v1/developerguide/configuring-sdk.html
	// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/

	cfg := aws.NewConfig()
	cfg.WithRegion(s3cfg.Region)

	if strings.HasPrefix(s3cfg.Credentials, "env:") {

		creds := credentials.NewEnvCredentials()
		cfg.WithCredentials(creds)

	} else if strings.HasPrefix(s3cfg.Credentials, "shared:") {

		details := strings.Split(s3cfg.Credentials, ":")

		if len(details) != 3 {
			return nil, errors.New("Shared credentials need to be defined as 'shared:CREDENTIALS_FILE:PROFILE_NAME'")
		}

		creds := credentials.NewSharedCredentials(details[1], details[2])
		cfg.WithCredentials(creds)

	} else if strings.HasPrefix(s3cfg.Credentials, "iam:") {

		// assume an IAM role suffient for doing whatever

	} else {

		return nil, errors.New("Unknown S3 config")
	}

	sess := session.New(cfg)

	if s3cfg.Credentials != "" {

		_, err := sess.Config.Credentials.Get()

		if err != nil {
			return nil, err
		}
	}

	service := s3.New(sess)

	c := S3Connection{
		service: service,
		bucket:  s3cfg.Bucket,
		prefix:  s3cfg.Prefix,
	}

	return &c, nil
}

func (conn *S3Connection) Put(key string, body []byte, content_type string) error {

	key = conn.prepareKey(key)

	params := &s3.PutObjectInput{
		Bucket:      aws.String(conn.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(content_type),
	}

	_, err := conn.service.PutObject(params)

	if err != nil {
		return err
	}

	return nil
}

func (conn *S3Connection) prepareKey(key string) string {

	if conn.prefix == "" {
		return key
	}

	return filepath.Join(conn.prefix, key)
}

func main() {

	var mime_types MimeTypes

	flag.Var(&mime_types, "mime-type", "...")

	var s3_credentials = flag.String("s3-credentials", "", "...")
	var s3_bucket = flag.String("s3-bucket", "whosonfirst.mapzen.com", "...")
	var s3_prefix = flag.String("s3-prefix", "", "...")
	var s3_region = flag.String("s3-region", "us-east-1", "...")

	var source = flag.String("source", "", "...")

	var strict = flag.Bool("strict", false, "...")
	var dryrun = flag.Bool("dryrun", false, "...")

	flag.Parse()

	if *dryrun {
		*strict = true
	}

	root, err := filepath.Abs(*source)

	if err != nil {
		log.Fatal(err)
	}

	cfg := S3Config{
		Bucket:      *s3_bucket,
		Prefix:      *s3_prefix,
		Region:      *s3_region,
		Credentials: *s3_credentials,
	}

	conn, err := NewS3Connection(cfg)

	if err != nil {
		log.Fatal(err)
	}

	for _, str_pair := range mime_types {
		pair := strings.Split(str_pair, "=")
		mime.AddExtensionType(pair[0], pair[1])
	}

	mime.AddExtensionType(".yaml", "text/x-yaml")
	mime.AddExtensionType(".example", "text/plain")

	callback := func(path string, info os.FileInfo) error {

		if info.IsDir() {
			return nil
		}

		rel_path := strings.Replace(path, root, "", -1)

		if strings.HasPrefix(rel_path, "/") {
			rel_path = strings.Replace(rel_path, "/", "", 1)
		}

		log.Printf("clone %s to s3://%s/%s/%s\n", path, cfg.Bucket, cfg.Prefix, rel_path)

		ext := filepath.Ext(path)

		content_type := mime.TypeByExtension(ext)

		if content_type == "" && *strict {
			msg := fmt.Sprintf("Unable to determine content type for %s", path)
			return errors.New(msg)
		}

		if *dryrun {
			return nil
		}

		body, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		err = conn.Put(rel_path, body, content_type)

		if err != nil {
			return err
		}

		return nil
	}

	c := crawl.NewCrawler(root)
	err = c.Crawl(callback)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
