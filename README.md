# go-whosonfirst-www

Go package for Who's On First www-related utilities.

## Tools

### wof-clone-website

Bare bones tools for cloning a static website to S3. This does not support exclusions yet.

```
./bin/wof-clone-website -strict -mime-type .yaml=text/x-yaml -s3-credentials shared:~.aws/credentials:default -s3-prefix nearby -source ../whosonfirst-www-nearby/www/
```
