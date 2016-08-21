FROM busybox

ADD ./voletc /

EXPOSE 8989

ENTRYPOINT ["/voletc"]

CMD ["-b", "0.0.0.0:8989"]