FROM gcc:latest
COPY . /usr/src/litmus
WORKDIR /usr/src/litmus
RUN ./configure
RUN make install
CMD litmus -k http://192.168.1.37:6001/ipfs/
