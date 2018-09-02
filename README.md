# DEPRECATED

This package is deprecated since IPFS ver. 0.4.3 (i guess), because developers of IPFS had changed the semantics of /add method, that doesn't accept file body anymore.

# mipfs (middleware for ipfs)

install ipfs

`docker run -d --name ipfs_host --restart always -p 8080:8080 -p 4001:4001 -p 127.0.0.1:5001:5001 ipfs/go-ipfs:latest`

`docker run -it --volumes-from ipfs_host ipfs/go-ipfs:latest`

`docker restart ipfs_host`

## ipfs2webdav

`git clone https://github.com/kpmy/mipfs.git`

`cd mipfs`

`docker build -t kpmy/mipfs:0.1 .`

`docker create --restart always --name ipfs_webdav --link ipfs_host:ipfs -p 0.0.0.0:6001:6001 kpmy/mipfs:0.1`

`docker start ipfs_webdav`

### how to

connect to webdav ipfs `cadaver http://<addr>:6001/ipfs/` 

upload some files `cadaver put /path/to/file`

then look in browser `http://<addr>:6001/hash`

default user/password: root:changeme

add new user: `curl -H "Content-Type: application/json" -d '{"login": "user", "password": "password"}' http://<host>:6001/user`

