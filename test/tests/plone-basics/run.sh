#!/bin/bash
set -eo pipefail

dir="$(dirname "$(readlink -f "$BASH_SOURCE")")"

image="$1"

PLONE_TEST_SLEEP=3
PLONE_TEST_TRIES=5

cname="plone-container-$RANDOM-$RANDOM"
cid="$(docker run -d --name "$cname" "$image")"
trap "docker rm -vf $cid > /dev/null" EXIT

get() {
	docker run --rm -i \
		--link "$cname":plone \
		--entrypoint python \
		"$image" \
		-c "import urllib2; con = urllib2.urlopen('$1'); print con.read()"
}

get_auth() {
	docker run --rm -i \
		--link "$cname":plone \
		--entrypoint python \
		"$image" \
		-c "import urllib2; request = urllib2.Request('$1'); request.add_header('Authorization', 'Basic $2'); print urllib2.urlopen(request).read()"
}


. "$dir/../../retry.sh" --tries "$PLONE_TEST_TRIES" --sleep "$PLONE_TEST_SLEEP" get "http://plone:8080"

# Plone is up and running
[[ "$(get 'http://plone:8080')" == *"Plone is up and running"* ]]

# Create a Plone site
[[ "$(get_auth 'http://plone:8080/@@plone-addsite' "$(echo -n 'admin:admin' | base64)")" == *"Create a Plone site"* ]]