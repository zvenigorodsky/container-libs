#!/usr/bin/env bash

TMPDIR=${TMPDIR:-/var/tmp}

btrfs() {
	[ $(stat -f -c %T ${TMPDIR}) = btrfs ]
}

overlay() {
	modprobe overlay 2> /dev/null
	grep -E -q '	overlay$' /proc/filesystems
}

zfs() {
	[ "$(stat -f -c %T ${TMPDIR:-/tmp})" = zfs ]
}

if [ "$STORAGE_DRIVER" = "" ] ; then
	drivers=vfs
	if btrfs; then
		drivers="$drivers btrfs"
	fi
	if overlay; then
		drivers="$drivers overlay"
	fi
	if zfs; then
		drivers="$drivers zfs"
	fi
else
	drivers="$STORAGE_DRIVER"
fi
set -e
for driver in $drivers ; do
	echo '['STORAGE_DRIVER="$driver"']'
	env STORAGE_DRIVER="$driver" $(dirname ${BASH_SOURCE})/test_runner.bash "$@"
done
