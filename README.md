# mipfs
ipfs2webdav

установка ipfs

`docker run -d --name ipfs_host -p 8080:8080 -p 4001:4001 -p 5001:5001 ipfs/go-ipfs:latest`

`docker run -it --volumes-from ipfs_host ipfs/go-ipfs:latest`

установка ipfs2webdav

`git clone https://github.com/kpmy/mipfs.git`

`docker build -t kpmy/mipfs:0.1 .`

`docker create --restart always --name ipfs_webdav --link ipfs_host:ipfs -p 0.0.0.0:6001:6001 kpmy/mipfs:0.1`

`docker start ipfs_webdav`
