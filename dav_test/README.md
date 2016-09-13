# тест litmus

http://www.webdav.org/neon/litmus/

`wget http://www.webdav.org/neon/litmus/litmus-0.13.tar.gz`

распаковать

взять отсюда Dockerfile

`docker build -t kpmy/litmus:0.13 .`

`docker run -it --name litmus kpmy/litmus:0.13`

`docker start -i litmus`...
