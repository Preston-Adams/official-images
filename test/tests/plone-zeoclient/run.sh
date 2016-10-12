#!/bin/bash
set -eo pipefail

dir="$(dirname "$(readlink -f "$BASH_SOURCE")")"

image="$1"

PLONE_TEST_SLEEP=3
PLONE_TEST_TRIES=5

# Start ZEO server
zname="zeo-container-$RANDOM-$RANDOM"
zid="$(docker run -d --name "$zname" "$image" zeoserver)"

# Start Plone as ZEO Client
pname="plone-container-$RANDOM-$RANDOM"
pid="$(docker run -d --name "$pname" --link=$zname:zeo -e ZEO_ADDRESS=zeo:8100 "$image")"

# Tear down
trap "docker rm -vf $pid $zid > /dev/null" EXIT

get() {
	docker run --rm -i \
		--link "$pname":plone \
		--entrypoint python \
		"$image" \
		-c "import urllib2; con = urllib2.urlopen('$1'); print con.read()"
}

get_auth() {
	docker run --rm -i \
		--link "$pname":plone \
		--entrypoint python \
		"$image" \
		-c "import urllib2; request = urllib2.Request('$1'); request.add_header('Authorization', 'Basic $2'); print urllib2.urlopen(request).read()"
}

. "$dir/../../retry.sh" --tries "$PLONE_TEST_TRIES" --sleep "$PLONE_TEST_SLEEP" get "http://plone:8080"

# Plone is up and running
[[ "$(get 'http://plone:8080')" == *"Plone is up and running"* ]]

# Create a Plone site
[[ "$(get_auth 'http://plone:8080/@@plone-addsite' "$(echo -n 'admin:admin' | base64)")" == *"Create a Plone site"* ]]