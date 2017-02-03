# go-whosonfirst-www

Go package for Who's On First www-related utilities.

## Tools

### wof-clone-website

A bare bones tool for cloning a static website to S3. This does not support exclusions yet.

```
./bin/wof-clone-website -strict -mime-type .yaml=text/x-yaml -s3-credentials shared:~.aws/credentials:default -s3-prefix nearby -source ../whosonfirst-www-nearby/www/
```

### wof-mk-static

A bare bones tool for generating ID-based static files for Who's On First placetypes and sources.

```
./bin/wof-mk-static -static ../whosonfirst-www/www/placetypes ../whosonfirst-placetypes/placetypes/*.json
./bin/wof-mk-static -id id -static ../whosonfirst-www/www/sources ../whosonfirst-sources/sources/*.json
```

_See the way we have to pass an `-id` flag to the second command? That's a thing we need to fix..._

Or something like this:

```
./bin/wof-mk-static -id id -static ../whosonfirst-sources/static ../whosonfirst-sources/sources/*.json
./bin/wof-clone-website -s3-credentials shared:/path/to/.aws/credentials:default -s3-bucket whosonfirst.mapzen.com -s3-prefix data -source ../whosonfirst-sources/static/
```      
